package main

import (
	"accounts-service/auth"
	"accounts-service/models"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"accounts-service/validators"
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type conversationsAPI struct {
	accountsv1.UnimplementedConversationsAPIServer

	auth         auth.Service
	logger       *zap.Logger
	messageRepo  models.ConversationMessagesRepository
	repo         models.ConversationsRepository
	groupService accountsv1.GroupsAPIServer
}

var _ accountsv1.ConversationsAPIServer = &conversationsAPI{}

func (server *conversationsAPI) CreateConversation(ctx context.Context, in *accountsv1.CreateConversationRequest) (*accountsv1.CreateConversationResponse, error) {
	err := validators.ValidateCreateConversationRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	conversation, err := server.repo.Create(ctx, &models.CreateConversationPayload{GroupID: in.GroupId, Title: in.Title})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	return &accountsv1.CreateConversationResponse{
		Conversation: &accountsv1.Conversation{
			Id:      conversation.ID,
			GroupId: conversation.GroupID,
			Title:   conversation.Title,
		},
	}, nil
}

func (server *conversationsAPI) GetConversation(ctx context.Context, in *accountsv1.GetConversationRequest) (*accountsv1.GetConversationResponse, error) {
	err := validators.ValidateGetConversationRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := server.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	conversation, err := server.repo.Get(ctx, &models.OneConversationFilter{ID: in.ConversationId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	accountId := token.UserID.String()

	_, err = server.groupService.GetGroupMember(ctx, &accountsv1.GetGroupMemberRequest{GroupId: conversation.GroupID, AccountId: accountId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "not in group")
	}

	return &accountsv1.GetConversationResponse{
		Conversation: &accountsv1.Conversation{
			Id:      conversation.ID,
			GroupId: conversation.GroupID,
			Title:   conversation.Title,
		},
	}, nil
}

func (server *conversationsAPI) DeleteConversation(ctx context.Context, in *accountsv1.DeleteConversationRequest) (*accountsv1.DeleteConversationResponse, error) {
	err := validators.ValidateDeleteConversationRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := server.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	conversation, err := server.repo.Get(ctx, &models.OneConversationFilter{ID: in.ConversationId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	accountId := token.UserID.String()

	_, err = server.groupService.GetGroupMember(ctx, &accountsv1.GetGroupMemberRequest{GroupId: conversation.GroupID, AccountId: accountId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "not in group")
	}

	err = server.repo.Delete(ctx, &models.OneConversationFilter{ID: in.ConversationId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	return &accountsv1.DeleteConversationResponse{}, nil
}

func (server *conversationsAPI) UpdateConversation(ctx context.Context, in *accountsv1.UpdateConversationRequest) (*accountsv1.UpdateConversationResponse, error) {
	err := validators.ValidateUpdateConversationRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := server.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	conversation, err := server.repo.Get(ctx, &models.OneConversationFilter{ID: in.ConversationId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	accountId := token.UserID.String()

	_, err = server.groupService.GetGroupMember(ctx, &accountsv1.GetGroupMemberRequest{GroupId: conversation.GroupID, AccountId: accountId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "not in group")
	}

	filter := models.OneConversationFilter{ID: in.ConversationId}

	updatedConversation, err := server.repo.Update(ctx, &filter, &models.UpdateConversationPayload{Title: in.Title})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	return &accountsv1.UpdateConversationResponse{
		Conversation: &accountsv1.Conversation{
			Id:      in.ConversationId,
			GroupId: updatedConversation.GroupID,
			Title:   in.Title,
		},
	}, nil
}

func (server *conversationsAPI) ListConversations(ctx context.Context, in *accountsv1.ListConversationsRequest) (*accountsv1.ListConversationsResponse, error) {
	err := validators.ValidateListConversationRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := server.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	accountId := token.UserID.String()

	_, err = server.groupService.GetGroupMember(ctx, &accountsv1.GetGroupMemberRequest{GroupId: in.GroupId, AccountId: accountId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "not in group")
	}

	conversations, err := server.repo.List(ctx, &models.ManyConversationsFilter{GroupID: in.GroupId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	groupConversations := make([]*accountsv1.Conversation, 0)

	for _, conversation := range conversations {
		groupConversation := &accountsv1.Conversation{Id: conversation.ID, GroupId: conversation.GroupID, Title: conversation.Title}
		groupConversations = append(groupConversations, groupConversation)
	}
	return &accountsv1.ListConversationsResponse{
		Conversations: groupConversations,
	}, nil
}

// TODO: This function is duplicated from accountsService.authenticate().
// Find a way to extract this into a separate function or use a base class
// to share common behaviour.
func (srv *conversationsAPI) authenticate(ctx context.Context) (*auth.Token, error) {
	token, err := srv.auth.TokenFromContext(ctx)
	if err != nil {
		srv.logger.Debug("could not authenticate request", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	return token, nil
}
