package models

import (
	"context"
)

type Invite struct {
	ID                 string  `json:"id" bson:"id,omitempty"`
	SenderAccountID    *string `json:"sender_account_id" bson:"sender_account_id,omitempty"`
	RecipientAccountID *string `json:"recipient_account_id" bson:"recipient_account_id,omitempty"`
	GroupID            *string `json:"group_id" bson:"group_id,omitempty"`
}

type InvitePayload struct {
	SenderAccountID    *string `json:"sender_account_id" bson:"sender_account_id,omitempty"`
	RecipientAccountID *string `json:"recipient_account_id" bson:"recipient_account_id,omitempty"`
	GroupID            *string `json:"group_id" bson:"group_id,omitempty"`
}

type OneInviteFilter struct {
	ID string `json:"id" bson:"id,omitempty"`
}

type ManyInvitesFilter struct {
	SenderAccountID *string `json:"sender_account_id" bson:"sender_account_id,omitempty"`
	GroupID         *string `json:"group_id" bson:"group_id,omitempty"`
}

// InvitesRepository is safe for use in multiple goroutines.
type InvitesRepository interface {
	Create(ctx context.Context, filter *InvitePayload) (*Invite, error)

	Get(ctx context.Context, filter *OneInviteFilter) (*Invite, error)

	Delete(ctx context.Context, filter *OneInviteFilter) error

	Update(ctx context.Context, filter *OneInviteFilter, Invite *InvitePayload) (*Invite, error)

	List(ctx context.Context, filter *ManyInvitesFilter, pagination *Pagination) ([]Invite, error)
}
