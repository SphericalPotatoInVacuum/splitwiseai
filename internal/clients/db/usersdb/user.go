package usersdb

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type UserState int

const (
	New UserState = iota
	AwaitingOAuthCode
	IncompleteProfile
	Ready
	Uploaded
	SelectingUsers
	Splitting
)

func (s UserState) String() string {
	if s == New {
		return "new"
	}
	if s == AwaitingOAuthCode {
		return "awaiting_oauth_code"
	}
	if s == IncompleteProfile {
		return "incomplete_profile"
	}
	if s == Ready {
		return "ready"
	}
	if s == Uploaded {
		return "uploaded"
	}
	if s == SelectingUsers {
		return "selecting_users"
	}
	if s == Splitting {
		return "splitting"
	}
	return "unknown"
}

type AuthorizedState string

const (
	Authorized   AuthorizedState = "true"
	Unauthorized AuthorizedState = "false"
)

type User struct {
	TelegramId          string `dynamodbav:"telegram_id"`
	State               string `dynamodbav:"state"`
	SplitwiseOAuthState string `dynamodbav:"splitwise_oauth_state"`
	SplitwiseGroupId    string `dynamodbav:"splitwise_group_id"`
	Currency            string `dynamodbav:"currency"`
	Authorized          string `dynamodbav:"authorized"`
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
