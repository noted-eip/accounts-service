// Package models defines the data types, payloads and repository interfaces
// of the accounts service.
package models

import (
	"context"

	"github.com/google/uuid"
)

type Account struct {
	ID    uuid.UUID `json:"id" bson:"_id,omitempty"`
	Email string    `json:"email" bson:"email,omitempty"`
	Name  string    `json:"name" bson:"name,omitempty"`
	Hash  *[]byte   `json:"hash" bson:"hash,omitempty"`
}

type AccountPayload struct {
	Name  *string `json:"name" bson:"name,omitempty"`
	Email *string `json:"email" bson:"email,omitempty"`
	Hash  *[]byte `json:"hash" bson:"hash,omitempty"`
}

type OneAccountFilter struct {
	ID    uuid.UUID `json:"id" bson:"_id,omitempty"`
	Email string    `json:"email" bson:"email,omitempty"`
}

type ManyAccountFilter struct{}

// AccountsRepository is safe for use in multiple goroutines.
type AccountsRepository interface {
	Create(ctx context.Context, filter *AccountPayload) error

	Get(ctx context.Context, filter *OneAccountFilter) (*Account, error)

	Delete(ctx context.Context, filter *OneAccountFilter) error

	Update(ctx context.Context, filter *OneAccountFilter, account *AccountPayload) error

	List(ctx context.Context) (*[]Account, error)
}
