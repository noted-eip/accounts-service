// Package models defines the data types, payloads and repository interfaces
// of the accounts service.
package models

import (
	"context"

	"github.com/google/uuid"
)

type Group struct {
	ID      uuid.UUID `json:"id" bson:"_id,omitempty"`
	Name    *string   `json:"name" bson:"name,omitempty"`
	Members *[]Member `json:"members" bson:"members,omitempty"`
}

type GroupPayload struct {
	Name    *string   `json:"name" bson:"name,omitempty"`
	Members *[]Member `json:"members" bson:"members,omitempty"`
}

type Member struct {
	ID uuid.UUID `json:"account_id" bson:"_id,omitempty"`
}

type OneGroupFilter struct {
	ID   uuid.UUID `json:"id" bson:"_id,omitempty"`
	Name string    `json:"name" bson:"name,omitempty"`
}

// GroupsRepository is safe for use in multiple goroutines.
type GroupsRepository interface {
	Create(ctx context.Context, filter *GroupPayload) error

	Delete(ctx context.Context, filter *OneGroupFilter) error

	Get(ctx context.Context, filter *OneGroupFilter) (*Group, error)

	Update(ctx context.Context, filter *OneGroupFilter, account *GroupPayload) error

	List(ctx context.Context) ([]Group, error)
}
