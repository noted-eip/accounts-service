package models

import (
	"context"
)

type Conversation struct {
	ID      string `json:"id" bson:"_id,omitempty"`
	GroupID string `json:"group_id" bson:"group_id,omitempty"`
	Title   string `json:"title" bson:"title,omitempty"`
}

type CreateConversationPayload struct {
	Title   string `json:"title" bson:"title,omitempty"`
	GroupID string `json:"group_id" bson:"group_id,omitempty"`
}

type UpdateConversationPayload struct {
	Title string `json:"title" bson:"title,omitempty"`
}

type OneConversationFilter struct {
	ID string `json:"id" bson:"_id,omitempty"`
}

type ManyConversationsFilter struct {
	GroupID string `json:"group_id" bson:"group_id,omitempty"`
}

type ConversationsRepository interface {
	Create(ctx context.Context, filter *CreateConversationPayload) (*Conversation, error)

	Get(ctx context.Context, filter *OneConversationFilter) (*Conversation, error)

	Delete(ctx context.Context, filter *OneConversationFilter) error

	Update(ctx context.Context, filter *OneConversationFilter, info *UpdateConversationPayload) (*Conversation, error)

	List(ctx context.Context, filter *ManyConversationsFilter) ([]Conversation, error)
}
