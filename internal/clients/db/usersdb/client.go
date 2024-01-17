package usersdb

import (
	"context"
	"errors"
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
		_, err = client.createUserTable()
		if err != nil {
			zap.S().Panicw("Failed to create table", zap.Error(err))
		}
	}
	log.Info("Ensured table exists")
	return client, nil
}

// TableExists determines whether a DynamoDB table exists.
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

// CreateUserTable creates a DynamoDB table with a composite primary key defined as
// a string sort key named `title`, and a numeric partition key named `year`.
// This function uses NewTableExistsWaiter to wait for the table to be created by
// DynamoDB before it returns.
func (c *client) createUserTable() (*types.TableDescription, error) {
	log := zap.S().With("table", c.TableName)

	var tableDesc *types.TableDescription
	table, err := c.DynamoDbClient.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: aws.String("telegram_id"),
			AttributeType: types.ScalarAttributeTypeN,
		}, {
			AttributeName: aws.String("splitwise_group_id"),
			AttributeType: types.ScalarAttributeTypeN,
		}, {
			AttributeName: aws.String("currency"),
			AttributeType: types.ScalarAttributeTypeS,
		}, {
			AttributeName: aws.String("splitwise_oauth_state"),
			AttributeType: types.ScalarAttributeTypeS,
		}, {
			AttributeName: aws.String("state"),
			AttributeType: types.ScalarAttributeTypeS,
		}, {
			AttributeName: aws.String("authorized"),
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

// AddUser adds a user the DynamoDB table.
func (c *client) CreateUser(ctx context.Context, user *User) error {
	log := c.log.With("user", user)

	item, err := attributevalue.MarshalMap(user)
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

// UpdateUser updates the status of a user that already exists in the
// DynamoDB table. This function uses the `expression` package to build the update
// expression.
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
		log.Errorw("Couldn't build expression for update", zap.Error(err))
	} else {
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
			log.Errorw("Couldn't update user", zap.Error(err))
		} else {
			err = attributevalue.UnmarshalMap(response.Attributes, &attributeMap)
			if err != nil {
				log.Errorw("Couldn't unmarshall update response", zap.Error(err))
			}
		}
	}
	return attributeMap, err
}

// GetUser gets user data from the DynamoDB table by using the primary key id
func (c *client) GetUser(ctx context.Context, telegramId int64) (User, error) {
	log := zap.S().With("telegram_id", telegramId)
	log.Debug("Getting user")
	user := User{TelegramId: telegramId}
	getItemInput := dynamodb.GetItemInput{
		Key: user.GetKey(), TableName: aws.String(c.TableName),
	}
	response, err := c.DynamoDbClient.GetItem(ctx, &getItemInput)
	if err != nil {
		log.Errorw("Couldn't get info about user", zap.Error(err))
	} else {
		err = attributevalue.UnmarshalMap(response.Item, &user)
		if err != nil {
			log.Errorw("Couldn't unmarshal response", zap.Error(err))
		}
	}
	log.Debugw("Got user", "user", user)
	return user, err
}

// DeleteUser removes a user from the DynamoDB table.
func (c *client) DeleteUser(user User) error {
	log := c.log.With("user_id", user.TelegramId)
	_, err := c.DynamoDbClient.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(c.TableName), Key: user.GetKey(),
	})
	if err != nil {
		log.Errorw("Couldn't delete user from the table", zap.Error(err))
	}
	return err
}

// DeleteTable deletes the DynamoDB table and all of its data.
func (c *client) DeleteTable() error {
	log := c.log
	_, err := c.DynamoDbClient.DeleteTable(context.TODO(), &dynamodb.DeleteTableInput{
		TableName: aws.String(c.TableName)})
	if err != nil {
		log.Errorw("Couldn't delete table", c.TableName, zap.Error(err))
	}
	return err
}
