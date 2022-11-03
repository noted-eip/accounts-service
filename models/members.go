package models

import (
	"context"
	"time"
)

type Member struct {
	ID        string    `json:"_id" bson:"_id,omitempty"`
	Account   *string   `json:"account_id" bson:"account_id,omitempty"`
	Group     *string   `json:"group_id" bson:"group_id,omitempty"`
	Role      string    `json:"role" bson:"role,omitempty"`
	CreatedAt time.Time `json:"created_at" bson:"created_at,omitempty"`
}

type MemberPayload struct {
	Account   *string   `json:"account_id" bson:"account_id,omitempty"`
	Group     *string   `json:"group_id" bson:"group_id,omitempty"`
	Role      string    `json:"role" bson:"role,omitempty"`
	CreatedAt time.Time `json:"created_at" bson:"created_at,omitempty"`
}

type MemberFilter struct {
	Account *string `json:"account_id" bson:"account_id,omitempty"`
	Group   *string `json:"group_id" bson:"group_id,omitempty"`
}

type MembersRepository interface {
	Create(ctx context.Context, filter *MemberPayload) (*Member, error)

	DeleteOne(ctx context.Context, filter *MemberFilter) error

	DeleteMany(ctx context.Context, filter *MemberFilter) error

	Get(ctx context.Context, filter *MemberFilter) (*Member, error)

	Update(ctx context.Context, filter *MemberFilter, account *MemberPayload) (*Member, error)

	List(ctx context.Context, filter *MemberFilter) ([]Member, error)
}
