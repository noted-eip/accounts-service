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

type PendingAccountPayload struct {
	Name  *string `json:"name" bson:"name,omitempty"`
	Email *string `json:"email" bson:"email,omitempty"`
	Hash  *[]byte `json:"hash" bson:"hash,omitempty"`
}

type OnePendingAccountFilter struct {
	ID    string `json:"id" bson:"_id,omitempty"`
	Email string `json:"email" bson:"email,omitempty"`
}

type PendingAccountSecretToken struct {
	ID    string `json:"id" bson:"_id,omitempty"`
	Token string `json:"token" bson:"token,omitempty"`
}

type PendingAccountsRepository interface {
	Create(ctx context.Context, filter *PendingAccountPayload) (*PendingAccount, error)

	Get(ctx context.Context, filter *OnePendingAccountFilter) (*PendingAccount, error)

	GetMailsFromIDs(ctx context.Context, filter []*OnePendingAccountFilter) ([]string, error)

	Delete(ctx context.Context, filter *OnePendingAccountFilter) error

	Update(ctx context.Context, filter *OnePendingAccountFilter, account *PendingAccountPayload) (*PendingAccount, error)
}
