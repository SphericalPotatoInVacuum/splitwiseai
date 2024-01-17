package usersdb

import "context"

type Client interface {
	GetUser(ctx context.Context, telegramId string) (User, error)
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) (map[string]interface{}, error)
}
