// Package models defines the data types, payloads and repository interfaces
// of the accounts service.
package models

import "github.com/google/uuid"

type Account struct {
	ID uuid.UUID `json:"id"`
}

type AccountQuery struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

// AccountsRepository is safe for use in multiple goroutines.
type AccountsRepository interface {
	Get(filter *AccountQuery) (*Account, error)

	Delete(filter *AccountQuery) error
}
