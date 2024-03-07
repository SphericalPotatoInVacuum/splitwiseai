package models

import (
	tokensdb "github.com/SphericalPotatoInVacuum/splitwiseai/internal/models/tokens"
	usersdb "github.com/SphericalPotatoInVacuum/splitwiseai/internal/models/users"
)

type Models interface {
	User() usersdb.Model
	Token() tokensdb.Model
}
