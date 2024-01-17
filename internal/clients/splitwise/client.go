package splitwise

import (
	"context"
	"fmt"

	splitwiseApi "github.com/aanzolaavila/splitwise.go"
	"github.com/aanzolaavila/splitwise.go/resources"
	"golang.org/x/oauth2"
)

type Config struct {
	ClientId     string `env:"SPLITWISE_CLIENT_ID"`
	ClientSecret string `env:"SPLITWISE_CLIENT_SECRET"`
}

type client struct {
	instances map[int64]*instance
	oauthConf *oauth2.Config
}

type instance struct {
	splitwiseClient splitwiseApi.Client
}

func (c *client) GetOAuthUrl(state string) string {
	return c.oauthConf.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (c *client) GetOAuthToken(ctx context.Context, code string) (string, error) {
	tok, err := c.oauthConf.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("error setting exchange: %w", err)
	}

	return tok.AccessToken, nil
}

func (c *client) AddInstanceFromOAuthToken(ctx context.Context, key int64, token string) (Instance, error) {
	tok := oauth2.Token{AccessToken: token}
	httpClient := c.oauthConf.Client(ctx, &tok)

	splitwiseClient := splitwiseApi.Client{
		HttpClient: httpClient,
	}

	_, err := splitwiseClient.GetCurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting current user: %w", err)
	}

	inst := &instance{
		splitwiseClient: splitwiseClient,
	}

	c.instances[key] = inst

	return inst, nil
}

func (c *client) GetInstance(key int64) (Instance, bool) {
	instance, ok := c.instances[key]
	return instance, ok
}

func NewClient(cfg Config) (Client, error) {
	return &client{
		instances: map[int64]*instance{},
		oauthConf: &oauth2.Config{
			ClientID:     cfg.ClientId,
			ClientSecret: cfg.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://secure.splitwise.com/oauth/authorize",
				TokenURL: "https://secure.splitwise.com/oauth/token",
			},
			RedirectURL: "https://primary-mutual-peacock.ngrok-free.app/splitwise/callback",
		},
	}, nil
}

func (c *instance) GetGroup(ctx context.Context, groupId int) (*resources.Group, error) {
	group, err := c.splitwiseClient.GetGroup(ctx, groupId)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (c *instance) GetGroups(ctx context.Context) ([]resources.Group, error) {
	groups, err := c.splitwiseClient.GetGroups(ctx)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (c *instance) GetGroupUsers(ctx context.Context, groupId int) ([]resources.User, error) {
	group, err := c.splitwiseClient.GetGroup(ctx, groupId)
	if err != nil {
		return nil, err
	}

	return group.Members, nil
}

func (c *instance) GetCurrencies(ctx context.Context) ([]resources.Currency, error) {
	currencies, err := c.splitwiseClient.GetCurrencies(ctx)
	if err != nil {
		return nil, err
	}

	return currencies, nil
}

func (c *instance) CreateExpense(ctx context.Context) error {
	c.splitwiseClient.CreateExpenseByShares(
		ctx,
		0.0,
		"some expense",
		123,
		splitwiseApi.CreateExpenseParams{
			splitwiseApi.CreateExpenseDate:         "2021-01-01",
			splitwiseApi.CreateExpenseCurrencyCode: "USD",
		},
		[]splitwiseApi.ExpenseUser{
			{
				Id:        123,
				PaidShare: 0.0,
				OwedShare: 0.0,
			},
		},
	)
	return fmt.Errorf("not implemented")
}
