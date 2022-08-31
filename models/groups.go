package models

import (
	"context"
)

type Group struct {
	ID          string         `json:"id" bson:"_id,omitempty"`
	Name        *string        `json:"name" bson:"name,omitempty"`
	OwnerID     string         `json:"owner_id" bson:"owner_id,omitempty"`
	Description *string        `json:"description" bson:"description,omitempty"`
	Members     *[]GroupMember `json:"members" bson:"members,omitempty"`
}

type GroupPayload struct {
	Name        *string        `json:"name" bson:"name,omitempty"`
	Description *string        `json:"description" bson:"description,omitempty"`
	OwnerID     string         `json:"owner_id" bson:"owner_id,omitempty"`
	Members     *[]GroupMember `json:"members" bson:"members,omitempty"`
}

type GroupMember struct {
	ID string `json:"account_id" bson:"account_id,omitempty"`
}

type OneGroupFilter struct {
	ID      string `json:"id" bson:"_id,omitempty"`
	OwnerID string `json:"owner_id" bson:"owner_id,omitempty"`
}

type ManyGroupsFilter struct {
}

// GroupsRepository is safe for use in multiple goroutines.
type GroupsRepository interface {
	Create(ctx context.Context, filter *GroupPayload) (*Group, error)

	Delete(ctx context.Context, filter *OneGroupFilter) error

	Get(ctx context.Context, filter *OneGroupFilter) (*Group, error)

	Update(ctx context.Context, filter *OneGroupFilter, account *GroupPayload) (*Group, error)

	List(ctx context.Context, filter *ManyGroupsFilter, pagination *Pagination) ([]Group, error)
}
