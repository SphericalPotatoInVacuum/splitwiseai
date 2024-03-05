package usersdb

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type User struct {
	TelegramId          int64  `dynamodbav:"telegram_id"`
	State               string `dynamodbav:"state"`
	SplitwiseOAuthState string `dynamodbav:"splitwise_oauth_state"`
	SplitwiseGroupId    int64  `dynamodbav:"splitwise_group_id"`
	Currency            string `dynamodbav:"currency"`
	Authorized          bool   `dynamodbav:"authorized"`
}

func (u User) GetKey() map[string]types.AttributeValue {
	telegramId, err := attributevalue.Marshal(u.TelegramId)
	if err != nil {
		panic(err)
	}
	return map[string]types.AttributeValue{
		"telegram_id": telegramId,
	}
}
