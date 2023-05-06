package models

import (
	"context"
	"time"
)

type Account struct {
	ID         string    `json:"id" bson:"_id,omitempty"`
	Email      *string   `json:"email" bson:"email,omitempty"`
	Name       *string   `json:"name" bson:"name,omitempty"`
	Hash       *[]byte   `json:"hash" bson:"hash,omitempty"`
	Token      string    `json:"token" bson:"token,omitempty"`
	ValidUntil time.Time `json:"valid_until" bson:"valid_until,omitempty"`
}

type AccountPayload struct {
	Name  *string `json:"name" bson:"name,omitempty"`
	Email *string `json:"email" bson:"email,omitempty"`
	Hash  *[]byte `json:"hash" bson:"hash,omitempty"`
}

type OneAccountFilter struct {
	ID    string `json:"id" bson:"_id,omitempty"`
	Email string `json:"email" bson:"email,omitempty"`
}

type AccountSecretToken struct {
	ID         string    `json:"id" bson:"_id,omitempty"`
	Token      string    `json:"token" bson:"token,omitempty"`
	ValidUntil time.Time `json:"valid_until" bson:"valid_until,omitempty"`
}

type ManyAccountsFilter struct{}

// AccountsRepository is safe for use in multiple goroutines.
type AccountsRepository interface {
	Create(ctx context.Context, filter *AccountPayload) (*Account, error)

	Get(ctx context.Context, filter *OneAccountFilter) (*Account, error)

	GetMailsFromIDs(ctx context.Context, filter []*OneAccountFilter) ([]string, error)

	Delete(ctx context.Context, filter *OneAccountFilter) error

	Update(ctx context.Context, filter *OneAccountFilter, account *AccountPayload) (*Account, error)

	List(ctx context.Context, filter *ManyAccountsFilter, pagination *Pagination) ([]Account, error)

	UpdateAccountWithResetPasswordToken(ctx context.Context, filter *OneAccountFilter) (*AccountSecretToken, error)

	UpdateAccountPassword(ctx context.Context, filter *OneAccountFilter, account *AccountPayload) (*Account, error)
}
