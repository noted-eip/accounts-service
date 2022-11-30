package main

import (
	"accounts-service/auth"
	"accounts-service/models"
	tchatsv1 "accounts-service/protorepo/noted/accounts/v1"
	"context"

	"go.uber.org/zap"
)

type tchatsAPI struct {
	tchatsv1.UnimplementedConversationsAPIServer

	auth   auth.Service
	logger *zap.Logger
	repo   models.TchatsRepository
}

var _ tchatsv1.ConversationsAPIServer = &tchatsAPI{}

func (server *tchatsAPI) CreateConversation(ctx context.Context, in *tchatsv1.CreateConversationRequest) (*tchatsv1.CreateConversationResponse, error) {
	return nil, nil
}

func (server *tchatsAPI) UpdateConversation(ctx context.Context, in *tchatsv1.UpdateConversationRequest) (*tchatsv1.UpdateConversationResponse, error) {
	return nil, nil
}

func (server *tchatsAPI) SendConversationMessage(ctx context.Context, in *tchatsv1.SendConversationMessageRequest) (*tchatsv1.SendConversationMessageResponse, error) {
	return nil, nil
}

func (server *tchatsAPI) DeleteConversationMessage(ctx context.Context, in *tchatsv1.DeleteConversationMessageRequest) (*tchatsv1.DeleteConversationMessageResponse, error) {
	return nil, nil
}

func (server *tchatsAPI) GetConversationMessage(ctx context.Context, in *tchatsv1.GetConversationMessageRequest) (*tchatsv1.GetConversationMessageResponse, error) {
	return nil, nil
}

func (server *tchatsAPI) UpdateConversationMessage(ctx context.Context, in *tchatsv1.UpdateConversationMessageRequest) (*tchatsv1.UpdateConversationMessageResponse, error) {
	return nil, nil
}

func (server *tchatsAPI) ListConversationMessages(ctx context.Context, in *tchatsv1.ListConversationMessagesRequest) (*tchatsv1.ListConversationMessagesResponse, error) {
	return nil, nil
}

func (server *tchatsAPI) ListConversations(ctx context.Context, in *tchatsv1.ListConversationsRequest) (*tchatsv1.ListConversationsResponse, error) {
	return nil, nil
}
