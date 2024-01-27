package tokensdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.uber.org/zap"
)

type Config struct {
	Endpoint     string `env:"DB_ENDPOINT"`
	AwsKeyId     string `env:"DB_AWS_KEY_ID"`
	AwsSecretKey string `env:"DB_AWS_SECRET_KEY"`
	TableName    string `env:"TOKENS_TABLE_NAME"`
}

type client struct {
	DynamoDbClient *dynamodb.Client
	TableName      string
	log            *zap.SugaredLogger
}

func NewClient(cfg Config) (Client, error) {
	log := zap.S().With("table", cfg.TableName)

	log.Debug("Creating DynamoDB client")
	DynamoDbClient := dynamodb.NewFromConfig(
		aws.Config{
			Region:      "ru-central1",
			Credentials: credentials.NewStaticCredentialsProvider(cfg.AwsKeyId, cfg.AwsSecretKey, ""),
		},
		func(o *dynamodb.Options) {
			o.BaseEndpoint = &cfg.Endpoint
		},
	)
	client := &client{DynamoDbClient: DynamoDbClient, TableName: cfg.TableName, log: log}
	exists, err := client.tableExists()
	if err != nil {
		return nil, fmt.Errorf("failed to check if table exists: %w", err)
	}
	if !exists {
		_, err = client.createTable()
		if err != nil {
			return nil, fmt.Errorf("failed to create table: %w", err)
		}
	}
	log.Debug("Ensured table exists")
	return client, nil
}

func (c *client) tableExists() (bool, error) {
	log := zap.S().With("table", c.TableName)
	exists := true
	_, err := c.DynamoDbClient.DescribeTable(
		context.TODO(), &dynamodb.DescribeTableInput{TableName: aws.String(c.TableName)},
	)
	if err != nil {
		var notFoundEx *types.ResourceNotFoundException
		if errors.As(err, &notFoundEx) {
			err = nil
		} else {
			log.Errorw("Couldn't determine existence of table", zap.Error(err))
		}
		exists = false
	}
	return exists, err
}

func (c *client) createTable() (*types.TableDescription, error) {
	log := zap.S().With("table", c.TableName)

	var tableDesc *types.TableDescription
	table, err := c.DynamoDbClient.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: aws.String("telegram_id"),
			AttributeType: types.ScalarAttributeTypeN,
		}},
		KeySchema: []types.KeySchemaElement{{
			AttributeName: aws.String("telegram_id"),
			KeyType:       types.KeyTypeHash,
		}},
		TableName: aws.String(c.TableName),
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create the table: %w", err)
	} else {
		waiter := dynamodb.NewTableExistsWaiter(c.DynamoDbClient)
		err = waiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{
			TableName: aws.String(c.TableName)}, 5*time.Minute)
		if err != nil {
			return nil, fmt.Errorf("failed to wait for table existance: %w", err)
		}
		tableDesc = table.TableDescription
	}

	log.Infoln("Created table")

	return tableDesc, err
}

func (c *client) PutToken(ctx context.Context, token *Token) error {
	log := c.log.With("telegram_id", token.TelegramId)
	log.Debug("Putting token")

	item, err := attributevalue.MarshalMap(token)
	if err != nil {
		panic(err)
	}
	_, err = c.DynamoDbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(c.TableName), Item: item,
	})
	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}
	log.Debug("Put token")
	return nil
}

func (c *client) GetToken(ctx context.Context, telegramId int64) (*Token, error) {
	log := zap.S().With("telegram_id", telegramId)
	log.Debug("Getting token")

	token := Token{TelegramId: telegramId}
	getItemInput := dynamodb.GetItemInput{
		Key: token.GetKey(), TableName: aws.String(c.TableName),
	}

	response, err := c.DynamoDbClient.GetItem(ctx, &getItemInput)
	if err != nil {
		log.Errorw("Couldn't get info about token", zap.Error(err))
		return nil, fmt.Errorf("failed to get token from DynamoDB: %w", err)
	}

	if response.Item == nil {
		log.Debug("Token not found")
		return nil, nil
	}

	err = attributevalue.UnmarshalMap(response.Item, &token)
	if err != nil {
		log.Errorw("Couldn't unmarshal response", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal token data: %w", err)
	}

	log.Debug("Got token")
	return &token, nil
}

func (c *client) DeleteToken(ctx context.Context, telegramId int64) error {
	log := c.log.With("telegram_id", telegramId)
	log.Debug("Deleting token")
	_, err := c.DynamoDbClient.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(c.TableName), Key: Token{TelegramId: telegramId}.GetKey(),
	})
	if err != nil {
		log.Errorw("Couldn't delete token from the table", zap.Error(err))
		return fmt.Errorf("failed to delete token from DynamoDB: %w", err)
	}
	log.Debug("Deleted token")
	return nil
}

func (c *client) DeleteTable() error {
	log := c.log
	_, err := c.DynamoDbClient.DeleteTable(context.TODO(), &dynamodb.DeleteTableInput{
		TableName: aws.String(c.TableName)})
	if err != nil {
		log.Errorw("Couldn't delete table", c.TableName, zap.Error(err))
	}
	return err
}
