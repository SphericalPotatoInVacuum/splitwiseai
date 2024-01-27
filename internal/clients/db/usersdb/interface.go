package usersdb

import "context"

type Client interface {
	GetUser(ctx context.Context, telegramId int64) (*User, error)
	PutUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) (map[string]interface{}, error)
}
