package models

import (
	"context"
)

// type Account struct {
// 	ID    string  `json:"id" bson:"_id,omitempty"`
// 	Email *string `json:"email" bson:"email,omitempty"`
// 	Name  *string `json:"name" bson:"name,omitempty"`
// 	Hash  *[]byte `json:"hash" bson:"hash,omitempty"`
// }

// type AccountPayload struct {
// 	Name  *string `json:"name" bson:"name,omitempty"`
// 	Email *string `json:"email" bson:"email,omitempty"`
// 	Hash  *[]byte `json:"hash" bson:"hash,omitempty"`
// }

// type OneAccountFilter struct {
// 	ID    string  `json:"id" bson:"_id,omitempty"`
// 	Email *string `json:"email" bson:"email,omitempty"`
// }

// type ManyAccountsFilter struct{}

type TchatsRepository interface {
	Create(ctx context.Context) (error)

	Get(ctx context.Context) (error)

	Delete(ctx context.Context) (error)

	Update(ctx context.Context) (error)

	// List(ctx context.Context, filter *ManyAccountsFilter, pagination *Pagination) ([]Account, error)
}
