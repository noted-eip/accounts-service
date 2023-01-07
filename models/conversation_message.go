package models

import (
	"context"
	"time"
)

type ConversationMessage struct {
	ID              string    `json:"id" json:"id,omitempty"`
	ConversationID  string    `json:"conversation_id" json:"conversation_id,omitempty"`
	SenderAccountID string    `json:"sender_account_id" bson:"sender_account_id,omitempty"`
	Content         string    `json:"content" bson:"content,omitempty"`
	CreatedAt       time.Time `json:"created_at" bson:"created_at,omitempty"`
}

type CreateConversationMessagePayload struct {
	ConversationID  string `json:"conversation_id" json:"conversation_id,omitempty"`
	SenderAccountID string `json:"sender_account_id" bson:"sender_account_id,omitempty"`
	Content         string `json:"content" bson:"content,omitempty"`
}

type UpdateConversationMessagePayload struct {
	Content string `json:"content" bson:"content,omitempty"`
}

type OneConversationMessageFilter struct {
	ID             string `json:"id" json:"id,omitempty"`
	ConversationID string `json:"conversation_id" json:"conversation_id,omitempty"`
}

type ManyConversationMessagesFilter struct {
	ConversationID string `json:"conversation_id" json:"conversation_id,omitempty"`
}

type ConversationMessagesRepository interface {
	Create(ctx context.Context, filter *CreateConversationMessagePayload) (*ConversationMessage, error)

	Get(ctx context.Context, filter *OneConversationMessageFilter) (*ConversationMessage, error)

	Delete(ctx context.Context, filter *OneConversationMessageFilter) error

	Update(ctx context.Context, filter *OneConversationMessageFilter, info *UpdateConversationMessagePayload) (*ConversationMessage, error)

	List(ctx context.Context, filter *ManyConversationMessagesFilter, pagination *Pagination) ([]ConversationMessage, error)
}
