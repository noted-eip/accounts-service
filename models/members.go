package models

import (
	"context"
)

type Member struct {
	ID      string  `json:"member_id" bson:"member_id,omitempty"`
	Account string  `json:"account_id" bson:"account_id,omitempty"`
	Group   string  `json:"group_id" bson:"group_id,omitempty"`
	Role    *string `json:"role" bson:"role,omitempty"`
}

type MemberPayload struct {
	Account string  `json:"account_id" bson:"account_id,omitempty"`
	Group   string  `json:"group_id" bson:"group_id,omitempty"`
	Role    *string `json:"role" bson:"role,omitempty"`
}

type MemberFilter struct {
	Account string `json:"account_id" bson:"account_id,omitempty"`
	Group   string `json:"group_id" bson:"group_id,omitempty"`
}

type MembersRepository interface {
	Create(ctx context.Context, filter *MemberPayload) (*Member, error)

	Delete(ctx context.Context, filter *MemberFilter) error

	Get(ctx context.Context, filter *MemberFilter) (*Member, error)

	Update(ctx context.Context, filter *MemberFilter, account *MemberPayload) (*Member, error)

	List(ctx context.Context, filter *MemberFilter, pagination *Pagination) ([]Member, error)
}
