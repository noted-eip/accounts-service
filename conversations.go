package main

import (
	"accounts-service/auth"
	"accounts-service/models"
	conversationsv1 "accounts-service/protorepo/noted/accounts/v1"
	"accounts-service/validators"
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type conversationsAPI struct {
	conversationsv1.UnimplementedConversationsAPIServer

	auth   auth.Service
	logger *zap.Logger
	repo   models.ConversationsRepository
}

var _ conversationsv1.ConversationsAPIServer = &conversationsAPI{}

func (server *conversationsAPI) CreateConversation(ctx context.Context, in *conversationsv1.CreateConversationRequest) (*conversationsv1.CreateConversationResponse, error) {
	err := validators.ValidateCreateConversationRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	groupConversations, err := server.repo.List(ctx, &models.AllConversationsFilter{ID: in.GroupId})

	for _, conv := range groupConversations {
		res := conv.Title == in.Title
		if res {
			return nil, status.Error(codes.Internal, "a conversations with this title always exist")
		}
	}

	conversation, err := server.repo.Create(ctx, &models.ConversationInfo{Title: in.Title, GroupID: in.GroupId})
	if err != nil {
		return nil, statusFromModelError(err)
	}
	return &conversationsv1.CreateConversationResponse{
		Conversation: &conversationsv1.Conversation{
			Id:      conversation.ID,
			GroupId: conversation.GroupID,
			Title:   conversation.Title,
		},
	}, nil
}

func (server *conversationsAPI) GetConversation(ctx context.Context, in *conversationsv1.GetConversationRequest) (*conversationsv1.GetConversationResponse, error) {
	err := validators.ValidateGetConversationRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not get Conversation")
	}

	conversation, err := server.repo.Get(ctx, &models.OneConversationFilter{ID: in.ConversationId})
	if err != nil {
		server.logger.Error("failed to get conversation from conversation id", zap.Error(err))
	}

	return &conversationsv1.GetConversationResponse{
		Conversation: &conversationsv1.Conversation{
			Id:      conversation.ID,
			GroupId: conversation.GroupID,
			Title:   conversation.Title,
		},
	}, nil
}

func (server *conversationsAPI) DeleteConversation(ctx context.Context, in *conversationsv1.DeleteConversationRequest) (*conversationsv1.DeleteConversationResponse, error) {
	err := validators.ValidateDeleteConversationRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = server.repo.Delete(ctx, &models.OneConversationFilter{ID: in.ConversationId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	return &conversationsv1.DeleteConversationResponse{}, nil
}

func (server *conversationsAPI) UpdateConversation(ctx context.Context, in *conversationsv1.UpdateConversationRequest) (*conversationsv1.UpdateConversationResponse, error) {
	err := validators.ValidateUpdateConversationRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not validate update conversation request")
	}

	filter := models.OneConversationFilter{ID: in.ConversationId}

	conversation, err := server.repo.Get(ctx, &filter)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get conversation to update")
	}

	if conversation.Title == in.Title {
		return nil, status.Error(codes.Internal, "new title same as old title")
	}

	groupConversations, err := server.repo.List(ctx, &models.AllConversationsFilter{ID: conversation.GroupID})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get list of all conversations of group")
	}

	for _, conv := range groupConversations {
		res := conv.Title == in.Title
		if res {
			return nil, status.Error(codes.Internal, "a conversations with this title always exist")
		}
	}

	_, err = server.repo.Update(ctx, &filter, &models.ConversationTitle{Title: in.Title})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update conversation title")
	}

	return &conversationsv1.UpdateConversationResponse{
		Conversation: &conversationsv1.Conversation{
			Id:      in.ConversationId,
			GroupId: conversation.GroupID,
			Title:   in.Title,
		},
	}, nil
}

func (server *conversationsAPI) ListConversations(ctx context.Context, in *conversationsv1.ListConversationsRequest) (*conversationsv1.ListConversationsResponse, error) {
	err := validators.ValidateListConversationRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not validate list conversation request")
	}

	conversations, err := server.repo.List(ctx, &models.AllConversationsFilter{ID: in.GroupId})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list conversations")
	}

	var groupConversations []*conversationsv1.Conversation

	for _, conversation := range conversations {
		if conversation.GroupID == in.GroupId {
			groupConversation := &conversationsv1.Conversation{Id: conversation.ID, GroupId: conversation.GroupID, Title: conversation.Title}
			if err != nil {
				server.logger.Error("failed to decode conversation", zap.Error(err))
			}
			groupConversations = append(groupConversations, groupConversation)
		}
	}
	return &conversationsv1.ListConversationsResponse{
		Conversations: groupConversations,
	}, nil
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
