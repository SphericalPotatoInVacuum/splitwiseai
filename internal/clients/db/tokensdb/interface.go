package tokensdb

import "context"

type Client interface {
	PutToken(ctx context.Context, token *Token) error
	GetToken(ctx context.Context, telegramId int64) (*Token, error)
	DeleteToken(ctx context.Context, telegramId int64) error
}
