package models

import (
	"context"
	"time"
)

type Account struct {
	ID              string    `json:"id" bson:"_id,omitempty"`
	Email           *string   `json:"email" bson:"email,omitempty"`
	Name            *string   `json:"name" bson:"name,omitempty"`
	Hash            *[]byte   `json:"hash" bson:"hash,omitempty"`
	ValidationToken string    `json:"validation_token" bson:"validation_token,omitempty"`
	IsValidated     bool      `json:"is_validated" bson:"is_validated"`
	IsInMobileBeta  bool      `json:"is_in_mobile_beta" bson:"is_in_mobile_beta,omitempty"`
	Token           string    `json:"token" bson:"token,omitempty"`
	ValidUntil      time.Time `json:"valid_until" bson:"valid_until,omitempty"`
}

type AccountPayload struct {
	Name  *string `json:"name" bson:"name,omitempty"`
	Email *string `json:"email" bson:"email,omitempty"`
	Hash  *[]byte `json:"hash" bson:"hash,omitempty"`
}

type OneAccountFilter struct {
	ID          string `json:"id" bson:"_id,omitempty"`
	Email       string `json:"email" bson:"email,omitempty"`
	IsValidated bool   `json:"is_validated" bson:"is_validated,omitempty"`
}

type AccountSecretToken struct {
	ID         string    `json:"id" bson:"_id,omitempty"`
	Token      string    `json:"token" bson:"token,omitempty"`
	ValidUntil time.Time `json:"valid_until" bson:"valid_until,omitempty"`
}

type ManyAccountsFilter struct{}

// AccountsRepository is safe for use in multiple goroutines.
type AccountsRepository interface {
	Create(ctx context.Context, filter *AccountPayload, isValidated bool) (*Account, error)

	Get(ctx context.Context, filter *OneAccountFilter) (*Account, error)

	GetMailsFromIDs(ctx context.Context, filter []*OneAccountFilter) ([]string, error)

	Delete(ctx context.Context, filter *OneAccountFilter) error

	Update(ctx context.Context, filter *OneAccountFilter, account *AccountPayload) (*Account, error)

	List(ctx context.Context, filter *ManyAccountsFilter, pagination *Pagination) ([]Account, error)

	UpdateAccountWithResetPasswordToken(ctx context.Context, filter *OneAccountFilter) (*AccountSecretToken, error)

	UpdateAccountPassword(ctx context.Context, filter *OneAccountFilter, account *AccountPayload) (*Account, error)

	UpdateAccountValidationState(ctx context.Context, filter *OneAccountFilter) (*Account, error)

	RegisterUserToMobileBeta(ctx context.Context, filter *OneAccountFilter) (*Account, error)

	UnsetAccountPasswordAndSetValidationState(ctx context.Context, filter *OneAccountFilter) (*Account, error)
}
