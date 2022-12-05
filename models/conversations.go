package models

import (
	"context"
)

type Conversation struct {
	ID      string `json:"id" bson:"_id,omitempty"`
	GroupID string `json:"groupId" bson:"group_id,omitempty"`
	Title   string `json:"title" bson:"title,omitempty"`
}

type ConversationInfo struct {
	Title   string `json:"title" bson:"title,omitempty"`
	GroupID string `json:"groupId" bson:"group_id,omitempty"`
}

type OneConversationFilter struct {
	ID string `json:"id" bson:"_id"`
}

type AllConversationsFilter struct {
	ID string `json:"groupId" bson:"group_id"`
}

type ConversationsRepository interface {
	Create(ctx context.Context, filter *ConversationInfo) (*Conversation, error)

	Get(ctx context.Context, filter *OneConversationFilter) (*Conversation, error)

	Delete(ctx context.Context) error

	Update(ctx context.Context) error

	List(ctx context.Context, filter *AllConversationsFilter) ([]Conversation, error)
}
