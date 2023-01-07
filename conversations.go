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
	repo         models.ConversationsRepository
	groupService accountsv1.GroupsAPIServer
}

var _ accountsv1.ConversationsAPIServer = &conversationsAPI{}

func (server *conversationsAPI) CreateConversation(ctx context.Context, in *accountsv1.CreateConversationRequest) (*accountsv1.CreateConversationResponse, error) {
	token, err := server.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	accountId := token.UserID.String()

	_, err = server.groupService.GetGroupMember(ctx, &accountsv1.GetGroupMemberRequest{GroupId: in.GroupId, AccountId: accountId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "sender not in group")
	}

	err = validators.ValidateCreateConversationRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// groupConversations, err := server.repo.List(ctx, &models.ManyConversationsFilter{GroupID: in.GroupId})

	// for _, conv := range groupConversations {
	// 	res := conv.Title == in.Title
	// 	if res {
	// 		return nil, status.Error(codes.Internal, "a conversations with this title always exist")
	// 	}
	// }

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
	token, err := server.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	err = validators.ValidateGetConversationRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	conversation, err := server.repo.Get(ctx, &models.OneConversationFilter{ID: in.ConversationId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	accountId := token.UserID.String()

	_, err = server.groupService.GetGroupMember(ctx, &accountsv1.GetGroupMemberRequest{GroupId: conversation.GroupID, AccountId: accountId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "sender not in group")
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
	token, err := server.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	err = validators.ValidateDeleteConversationRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	conversation, err := server.repo.Get(ctx, &models.OneConversationFilter{ID: in.ConversationId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	accountId := token.UserID.String()

	_, err = server.groupService.GetGroupMember(ctx, &accountsv1.GetGroupMemberRequest{GroupId: conversation.GroupID, AccountId: accountId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "sender not in group")
	}

	err = server.repo.Delete(ctx, &models.OneConversationFilter{ID: in.ConversationId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	return &accountsv1.DeleteConversationResponse{}, nil
}

func (server *conversationsAPI) UpdateConversation(ctx context.Context, in *accountsv1.UpdateConversationRequest) (*accountsv1.UpdateConversationResponse, error) {
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
		return nil, status.Error(codes.InvalidArgument, "sender not in group")
	}

	err = validators.ValidateUpdateConversationRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
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
	token, err := server.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	accountId := token.UserID.String()

	_, err = server.groupService.GetGroupMember(ctx, &accountsv1.GetGroupMemberRequest{GroupId: in.GroupId, AccountId: accountId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "sender not in group")
	}

	err = validators.ValidateListConversationRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	conversations, err := server.repo.List(ctx, &models.ManyConversationsFilter{GroupID: in.GroupId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	var groupConversations []*accountsv1.Conversation

	for _, conversation := range conversations {
		groupConversation := &accountsv1.Conversation{Id: conversation.ID, GroupId: conversation.GroupID, Title: conversation.Title}
		groupConversations = append(groupConversations, groupConversation)
	}
	return &accountsv1.ListConversationsResponse{
		Conversations: groupConversations,
	}, nil
}

func (server *conversationsAPI) SendConversationMessage(ctx context.Context, in *accountsv1.SendConversationMessageRequest) (*accountsv1.SendConversationMessageResponse, error) {
	return nil, nil
}

func (server *conversationsAPI) DeleteConversationMessage(ctx context.Context, in *accountsv1.DeleteConversationMessageRequest) (*accountsv1.DeleteConversationMessageResponse, error) {
	return nil, nil
}

func (server *conversationsAPI) GetConversationMessage(ctx context.Context, in *accountsv1.GetConversationMessageRequest) (*accountsv1.GetConversationMessageResponse, error) {
	return nil, nil
}

func (server *conversationsAPI) UpdateConversationMessage(ctx context.Context, in *accountsv1.UpdateConversationMessageRequest) (*accountsv1.UpdateConversationMessageResponse, error) {
	return nil, nil
}

func (server *conversationsAPI) ListConversationMessages(ctx context.Context, in *accountsv1.ListConversationMessagesRequest) (*accountsv1.ListConversationMessagesResponse, error) {
	return nil, nil
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
