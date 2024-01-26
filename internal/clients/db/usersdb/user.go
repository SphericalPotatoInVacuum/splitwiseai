package usersdb

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type UserState string

const (
	Unauthorized      UserState = "unauthorized"
	AwaitingOAuthCode UserState = "awaiting_oauth_code"
	Ready             UserState = "ready"
	Uploaded          UserState = "uploaded"
	SelectingUsers    UserState = "selecting_users"
	Splitting         UserState = "splitting"
)

type User struct {
	TelegramId          int64  `dynamodbav:"telegram_id"`
	State               string `dynamodbav:"state"`
	SplitwiseOAuthState string `dynamodbav:"splitwise_oauth_state"`
	SplitwiseGroupId    uint64 `dynamodbav:"splitwise_group_id"`
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
