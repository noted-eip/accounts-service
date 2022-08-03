// Package models defines the data types, payloads and repository interfaces
// of the accounts service.
package models

import (
	"context"
)

type Group struct {
	ID          string    `json:"id" bson:"_id,omitempty"`
	Name        *string   `json:"name" bson:"name,omitempty"`
	Description *string   `json:"description" bson:"description,omitempty"`
	Members     *[]Member `json:"members" bson:"members,omitempty"`
	Notes       *[]Note   `json:"notes" bson:"notes,omitempty"`
}

type GroupPayload struct {
	Name        *string   `json:"name" bson:"name,omitempty"`
	Description *string   `json:"description" bson:"description,omitempty"`
	Members     *[]Member `json:"members" bson:"members,omitempty"`
	Notes       *[]Note   `json:"notes" bson:"notes,omitempty"`
}

type Member struct {
	ID string `json:"account_id" bson:"_id,omitempty"`
}

type Note struct {
	ID string `json:"note_id" bson:"_id,omitempty"`
}

type OneGroupFilter struct {
	ID   string  `json:"id" bson:"_id,omitempty"`
	Name *string `json:"name" bson:"name,omitempty"`
}

// GroupsRepository is safe for use in multiple goroutines.
type GroupsRepository interface {
	Create(ctx context.Context, filter *GroupPayload) error

	Delete(ctx context.Context, filter *OneGroupFilter) error

	Get(ctx context.Context, filter *OneGroupFilter) (*Group, error)

	Update(ctx context.Context, filter *OneGroupFilter, account *GroupPayload) error

	List(ctx context.Context) ([]Group, error)
}
