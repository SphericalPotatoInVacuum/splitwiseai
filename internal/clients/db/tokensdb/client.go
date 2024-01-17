package tokensdb

import (
	"context"
	"errors"
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

	log.Debugln("Creating DynamoDB client")
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
		log.Panicw("Failed to check if table exists", zap.Error(err))
	}
	if !exists {
		_, err = client.createTable()
		if err != nil {
			zap.S().Panicw("Failed to create table", zap.Error(err))
		}
	}
	log.Info("Ensured table exists")
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
			log.Infoln("Table doesn't exist")
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
		}, {
			AttributeName: aws.String("token"),
			AttributeType: types.ScalarAttributeTypeS,
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
		log.Errorw("Couldn't create table", zap.Error(err))
	} else {
		waiter := dynamodb.NewTableExistsWaiter(c.DynamoDbClient)
		err = waiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{
			TableName: aws.String(c.TableName)}, 5*time.Minute)
		if err != nil {
			log.Errorw("Wait for table exists failed", zap.Error(err))
		}
		tableDesc = table.TableDescription
	}

	log.Infoln("Created table")

	return tableDesc, err
}

func (c *client) PutToken(ctx context.Context, token *Token) error {
	log := c.log.With("telegram_id", token.TelegramId)

	item, err := attributevalue.MarshalMap(token)
	if err != nil {
		panic(err)
	}
	_, err = c.DynamoDbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(c.TableName), Item: item,
	})
	if err != nil {
		log.Errorw("Couldn't add item to table", zap.Error(err))
	}
	return err
}

func (c *client) GetToken(ctx context.Context, telegramId int64) (Token, error) {
	log := zap.S().With("telegram_id", telegramId)
	log.Debug("Getting token")
	user := Token{TelegramId: telegramId}
	getItemInput := dynamodb.GetItemInput{
		Key: user.GetKey(), TableName: aws.String(c.TableName),
	}
	response, err := c.DynamoDbClient.GetItem(ctx, &getItemInput)
	if err != nil {
		log.Errorw("Couldn't get token", zap.Error(err))
	} else {
		err = attributevalue.UnmarshalMap(response.Item, &user)
		if err != nil {
			log.Errorw("Couldn't unmarshal response", zap.Error(err))
		}
	}
	log.Debug("Got token")
	return user, err
}

func (c *client) DeleteToken(ctx context.Context, telegramId int64) error {
	log := c.log.With("telegram_id", telegramId)
	_, err := c.DynamoDbClient.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(c.TableName), Key: Token{TelegramId: telegramId}.GetKey(),
	})
	if err != nil {
		log.Errorw("Couldn't delete token from the table", zap.Error(err))
	}
	return err
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
