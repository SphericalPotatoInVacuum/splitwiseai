package tokensdb

import "context"

type Client interface {
	PutToken(ctx context.Context, token *Token) error
	GetToken(ctx context.Context, telegramId string) (Token, error)
	DeleteToken(ctx context.Context, telegramId string) error
}
