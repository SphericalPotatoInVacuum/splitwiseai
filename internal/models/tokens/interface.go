package tokensdb

import "context"

type Model interface {
	PutToken(ctx context.Context, token *Token) error
	GetToken(ctx context.Context, telegramId int64) (*Token, error)
	DeleteToken(ctx context.Context, telegramId int64) error
}
