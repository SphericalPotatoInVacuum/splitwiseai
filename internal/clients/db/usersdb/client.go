package usersdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.uber.org/zap"
)

type Config struct {
	Endpoint     string `env:"DB_ENDPOINT"`
	AwsKeyId     string `env:"DB_AWS_KEY_ID"`
	AwsSecretKey string `env:"DB_AWS_SECRET_KEY"`
	TableName    string `env:"USERS_TABLE_NAME"`
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
		log.Panicw("Failed to check if table exists", zap.Error(err))
	}
	if !exists {
		_, err = client.createUserTable()
		if err != nil {
			zap.S().Panicw("Failed to create table", zap.Error(err))
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

func (c *client) createUserTable() (*types.TableDescription, error) {
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

	log.Debug("Created table")

	return tableDesc, err
}

func (c *client) PutUser(ctx context.Context, user *User) error {
	log := c.log.With("user", user)
	log.Debug("Adding user")
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		panic(err)
	}
	_, err = c.DynamoDbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(c.TableName), Item: item,
	})
	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}
	log.Debug("Put user")
	return nil
}

func (c *client) UpdateUser(ctx context.Context, user *User) (map[string]interface{}, error) {
	log := c.log.With("user", user)
	log.Debug("Updating user")
	var err error
	var response *dynamodb.UpdateItemOutput
	var attributeMap map[string]interface{}
	update := expression.
		Set(expression.Name("splitwise_group_id"), expression.Value(user.SplitwiseGroupId)).
		Set(expression.Name("currency"), expression.Value(user.Currency)).
		Set(expression.Name("splitwise_oauth_state"), expression.Value(user.SplitwiseOAuthState)).
		Set(expression.Name("state"), expression.Value(user.State)).
		Set(expression.Name("authorized"), expression.Value(user.Authorized))

	expr, err := expression.NewBuilder().WithUpdate(update).Build()

	if err != nil {
		return nil, fmt.Errorf("failed to build update expression: %w", err)
	}
	updateItemInput := dynamodb.UpdateItemInput{
		TableName:                 aws.String(c.TableName),
		Key:                       user.GetKey(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueUpdatedNew,
	}
	response, err = c.DynamoDbClient.UpdateItem(ctx, &updateItemInput)
	if err != nil {
		return nil, fmt.Errorf("failed to update item: %w", err)
	}
	err = attributevalue.UnmarshalMap(response.Attributes, &attributeMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return attributeMap, nil
}

func (c *client) GetUser(ctx context.Context, telegramId int64) (*User, error) {
	log := zap.S().With("telegram_id", telegramId)
	log.Debug("Getting user")

	user := User{TelegramId: telegramId}
	getItemInput := dynamodb.GetItemInput{
		Key: user.GetKey(), TableName: aws.String(c.TableName),
	}

	response, err := c.DynamoDbClient.GetItem(ctx, &getItemInput)
	if err != nil {
		log.Errorw("Couldn't get info about user", zap.Error(err))
		return nil, fmt.Errorf("failed to get user from DynamoDB: %w", err)
	}

	if response.Item == nil {
		log.Debug("User not found")
		return nil, nil
	}

	err = attributevalue.UnmarshalMap(response.Item, &user)
	if err != nil {
		log.Errorw("Couldn't unmarshal response", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	log.Debugw("Got user", "user", user)
	return &user, nil
}

func (c *client) DeleteUser(user User) error {
	log := c.log.With("telegram_id", user.TelegramId)
	_, err := c.DynamoDbClient.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(c.TableName), Key: user.GetKey(),
	})
	if err != nil {
		log.Errorw("Couldn't delete user from the table", zap.Error(err))
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
