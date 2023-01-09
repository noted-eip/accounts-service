package models

import (
	"context"
)

type Account struct {
	ID    string  `json:"id" bson:"_id,omitempty"`
	Email *string `json:"email" bson:"email,omitempty"`
	Name  *string `json:"name" bson:"name,omitempty"`
	Hash  *[]byte `json:"hash" bson:"hash,omitempty"`
}

type AccountPayload struct {
	Name  *string `json:"name" bson:"name,omitempty"`
	Email *string `json:"email" bson:"email,omitempty"`
	Hash  *[]byte `json:"hash" bson:"hash,omitempty"`
}

type OneAccountFilter struct {
	ID    string  `json:"id" bson:"_id,omitempty"`
	Email *string `json:"email" bson:"email,omitempty"`
}

type ManyAccountsFilter struct {
	EmailContains *string
}

// AccountsRepository is safe for use in multiple goroutines.
type AccountsRepository interface {
	Create(ctx context.Context, filter *AccountPayload) (*Account, error)

	Get(ctx context.Context, filter *OneAccountFilter) (*Account, error)

	Delete(ctx context.Context, filter *OneAccountFilter) error

	Update(ctx context.Context, filter *OneAccountFilter, account *AccountPayload) (*Account, error)

	List(ctx context.Context, filter *ManyAccountsFilter, pagination *Pagination) ([]Account, error)
}
