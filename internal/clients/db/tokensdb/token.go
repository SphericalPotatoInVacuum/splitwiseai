package tokensdb

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Token struct {
	TelegramId string `dynamodbav:"telegram_id"`
	Token      string `dynamodbav:"token"`
}

func (t Token) GetKey() map[string]types.AttributeValue {
	telegramId, err := attributevalue.Marshal(t.TelegramId)
	if err != nil {
		panic(err)
	}
	return map[string]types.AttributeValue{
		"telegram_id": telegramId,
	}
}
