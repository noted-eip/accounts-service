package main

import (
	"accounts-service/auth"
	"accounts-service/models"
	conversationsv1 "accounts-service/protorepo/noted/accounts/v1"
	"context"

	"go.uber.org/zap"
)

type conversationsAPI struct {
	conversationsv1.UnimplementedConversationsAPIServer

	auth   auth.Service
	logger *zap.Logger
	repo   models.ConversationsRepository
}

var _ conversationsv1.ConversationsAPIServer = &conversationsAPI{}

func (server *conversationsAPI) CreateConversation(ctx context.Context, in *conversationsv1.CreateConversationRequest) (*conversationsv1.CreateConversationResponse, error) {
	return nil, nil
}

func (server *conversationsAPI) GetConversation(ctx context.Context, in *conversationsv1.GetConversationRequest) (*conversationsv1.GetConversationResponse, error) {
	return nil, nil
}

func (server *conversationsAPI) DeleteConversation(ctx context.Context, in *conversationsv1.DeleteConversationRequest) (*conversationsv1.DeleteConversationResponse, error) {
	return nil, nil
}

func (server *conversationsAPI) UpdateConversation(ctx context.Context, in *conversationsv1.UpdateConversationRequest) (*conversationsv1.UpdateConversationResponse, error) {
	return nil, nil
}

func (server *conversationsAPI) SendConversationMessage(ctx context.Context, in *conversationsv1.SendConversationMessageRequest) (*conversationsv1.SendConversationMessageResponse, error) {
	return nil, nil
}

func (server *conversationsAPI) DeleteConversationMessage(ctx context.Context, in *conversationsv1.DeleteConversationMessageRequest) (*conversationsv1.DeleteConversationMessageResponse, error) {
	return nil, nil
}

func (server *conversationsAPI) GetConversationMessage(ctx context.Context, in *conversationsv1.GetConversationMessageRequest) (*conversationsv1.GetConversationMessageResponse, error) {
	return nil, nil
}

func (server *conversationsAPI) UpdateConversationMessage(ctx context.Context, in *conversationsv1.UpdateConversationMessageRequest) (*conversationsv1.UpdateConversationMessageResponse, error) {
	return nil, nil
}

func (server *conversationsAPI) ListConversationMessages(ctx context.Context, in *conversationsv1.ListConversationMessagesRequest) (*conversationsv1.ListConversationMessagesResponse, error) {
	return nil, nil
}

func (server *conversationsAPI) ListConversations(ctx context.Context, in *conversationsv1.ListConversationsRequest) (*conversationsv1.ListConversationsResponse, error) {
	return nil, nil
}
