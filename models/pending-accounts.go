package models

import (
	"context"
)

// account not validate
type PendingAccount struct {
	ID    string  `json:"id" bson:"_id,omitempty"`
	Email *string `json:"email" bson:"email,omitempty"`
	Name  *string `json:"name" bson:"name,omitempty"`
	Hash  *[]byte `json:"hash" bson:"hash,omitempty"`
	Token string  `json:"token" bson:"token,omitempty"`
}

type PendingAccountSecretToken struct {
	ID    string `json:"id" bson:"_id,omitempty"`
	Token string `json:"token" bson:"token,omitempty"`
}

type PendingAccountsRepository interface {
	Create(ctx context.Context, filter *AccountPayload) (*PendingAccount, error)

	Get(ctx context.Context, filter *OneAccountFilter) (*PendingAccount, error)

	GetMailsFromIDs(ctx context.Context, filter []*OneAccountFilter) ([]string, error)

	Delete(ctx context.Context, filter *OneAccountFilter) error

	Update(ctx context.Context, filter *OneAccountFilter, account *AccountPayload) (*PendingAccount, error)
}
