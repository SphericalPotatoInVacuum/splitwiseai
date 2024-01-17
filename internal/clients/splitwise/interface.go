package splitwise

import (
	"context"

	"github.com/aanzolaavila/splitwise.go/resources"
)

type Client interface {
	GetOAuthUrl(state string) string
	GetOAuthToken(ctx context.Context, code string) (string, error)
	AddInstanceFromOAuthToken(ctx context.Context, key string, token string) (Instance, error)
	GetInstance(key string) (Instance, bool)
}

type Instance interface {
	GetGroup(ctx context.Context, groupId int) (*resources.Group, error)
	GetGroups(ctx context.Context) ([]resources.Group, error)
	GetGroupUsers(ctx context.Context, groupID int) ([]resources.User, error)
	GetCurrencies(ctx context.Context) ([]resources.Currency, error)
	CreateExpense(ctx context.Context) error
}
