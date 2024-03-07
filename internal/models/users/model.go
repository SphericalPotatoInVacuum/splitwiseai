package usersdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.uber.org/zap"
)

type model struct {
	db        *dynamodb.Client
	tableName string
	log       *zap.SugaredLogger
}

func NewModel(DB *dynamodb.Client, tableName string) (Model, error) {
	log := zap.S().With("table", tableName)

	model := &model{db: DB, tableName: tableName, log: log}

	exists, err := model.tableExists()
	if err != nil {
		log.Panicw("Failed to check if table exists", zap.Error(err))
	}

	if !exists {
		_, err = model.createUserTable()
		if err != nil {
			zap.S().Panicw("Failed to create table", zap.Error(err))
		}
	}

	log.Debug("Ensured table exists")

	return model, nil
}

func (m *model) tableExists() (bool, error) {
	log := zap.S().With("table", m.tableName)
	exists := true
	_, err := m.db.DescribeTable(
		context.TODO(), &dynamodb.DescribeTableInput{TableName: aws.String(m.tableName)},
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

func (m *model) createUserTable() (*types.TableDescription, error) {
	log := zap.S().With("table", m.tableName)

	var tableDesc *types.TableDescription
	table, err := m.db.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: aws.String("telegram_id"),
			AttributeType: types.ScalarAttributeTypeN,
		}},
		KeySchema: []types.KeySchemaElement{{
			AttributeName: aws.String("telegram_id"),
			KeyType:       types.KeyTypeHash,
		}},
		TableName: aws.String(m.tableName),
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
	})

	if err != nil {
		log.Errorw("Couldn't create table", zap.Error(err))
	} else {
		waiter := dynamodb.NewTableExistsWaiter(m.db)
		err = waiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{
			TableName: aws.String(m.tableName)}, 5*time.Minute)
		if err != nil {
			log.Errorw("Wait for table exists failed", zap.Error(err))
		}
		tableDesc = table.TableDescription
	}

	log.Debug("Created table")

	return tableDesc, err
}

func (m *model) PutUser(ctx context.Context, user *User) error {
	log := m.log.With("user", user)
	log.Debug("Adding user")
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		panic(err)
	}
	_, err = m.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(m.tableName), Item: item,
	})
	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}
	log.Debug("Put user")
	return nil
}

func (m *model) UpdateUser(ctx context.Context, user *User) (map[string]interface{}, error) {
	log := m.log.With("user", user)
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
		TableName:                 aws.String(m.tableName),
		Key:                       user.GetKey(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueUpdatedNew,
	}
	response, err = m.db.UpdateItem(ctx, &updateItemInput)
	if err != nil {
		return nil, fmt.Errorf("failed to update item: %w", err)
	}
	err = attributevalue.UnmarshalMap(response.Attributes, &attributeMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return attributeMap, nil
}

func (m *model) GetUser(ctx context.Context, telegramId int64) (*User, error) {
	log := zap.S().With("telegram_id", telegramId)
	log.Debug("Getting user")

	user := User{TelegramId: telegramId}
	getItemInput := dynamodb.GetItemInput{
		Key: user.GetKey(), TableName: aws.String(m.tableName),
	}

	response, err := m.db.GetItem(ctx, &getItemInput)
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

func (m *model) DeleteUser(user User) error {
	log := m.log.With("telegram_id", user.TelegramId)
	_, err := m.db.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(m.tableName), Key: user.GetKey(),
	})
	if err != nil {
		log.Errorw("Couldn't delete user from the table", zap.Error(err))
	}
	return err
}

func (m *model) DeleteTable() error {
	log := m.log
	_, err := m.db.DeleteTable(context.TODO(), &dynamodb.DeleteTableInput{
		TableName: aws.String(m.tableName)})
	if err != nil {
		log.Errorw("Couldn't delete table", m.tableName, zap.Error(err))
	}
	return err
}
