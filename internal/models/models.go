package models

import (
	"fmt"

	tokensdb "github.com/SphericalPotatoInVacuum/splitwiseai/internal/models/tokens"
	usersdb "github.com/SphericalPotatoInVacuum/splitwiseai/internal/models/users"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type models struct {
	user  usersdb.Model
	token tokensdb.Model
}

type Config struct {
	UsersTableName  string `env:"USERS_TABLE_NAME"`
	TokensTableName string `env:"TOKENS_TABLE_NAME"`
}

func NewModels(db *dynamodb.Client, cfg Config) (Models, error) {
	user, err := usersdb.NewModel(db, cfg.UsersTableName)
	if err != nil {
		return nil, fmt.Errorf("failed to create users model: %w", err)
	}

	token, err := tokensdb.NewModel(db, cfg.TokensTableName)
	if err != nil {
		return nil, fmt.Errorf("failed to create tokens model: %w", err)
	}

	return &models{user: user, token: token}, nil
}

func (m *models) User() usersdb.Model {
	return m.user
}

func (m *models) Token() tokensdb.Model {
	return m.token
}
