package main

import (
	"accounts-service/models"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"accounts-service/validators"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *conversationsAPI) SendConversationMessage(ctx context.Context, in *accountsv1.SendConversationMessageRequest) (*accountsv1.SendConversationMessageResponse, error) {
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

	err = validators.ValidateSendConversationMessageRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	message, err := server.messageRepo.Create(ctx, &models.CreateConversationMessagePayload{ConversationID: in.ConversationId, SenderAccountID: accountId, Content: in.Content})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	return &accountsv1.SendConversationMessageResponse{
		Message: &accountsv1.ConversationMessage{
			Id:              message.ID,
			ConversationId:  message.ConversationID,
			SenderAccountId: message.SenderAccountID,
			Content:         message.Content,
			CreatedAt:       timestamppb.New(message.CreatedAt),
		},
	}, nil
}

func (server *conversationsAPI) DeleteConversationMessage(ctx context.Context, in *accountsv1.DeleteConversationMessageRequest) (*accountsv1.DeleteConversationMessageResponse, error) {
	token, err := server.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	err = validators.ValidateDeleteConversationMessageRequest(in)
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

	message, err := server.messageRepo.Get(ctx, &models.OneConversationMessageFilter{ID: in.MessageId, ConversationID: in.ConversationId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	if message.SenderAccountID != accountId {
		return nil, status.Error(codes.InvalidArgument, "sender is not the owner of the message")
	}

	err = server.messageRepo.Delete(ctx, &models.OneConversationMessageFilter{ID: in.MessageId, ConversationID: in.ConversationId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	return &accountsv1.DeleteConversationMessageResponse{}, nil
}

func (server *conversationsAPI) GetConversationMessage(ctx context.Context, in *accountsv1.GetConversationMessageRequest) (*accountsv1.GetConversationMessageResponse, error) {
	token, err := server.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	err = validators.ValidateGetConversationMessageRequest(in)
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

	message, err := server.messageRepo.Get(ctx, &models.OneConversationMessageFilter{ID: in.ConversationId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	return &accountsv1.GetConversationMessageResponse{
		Message: &accountsv1.ConversationMessage{
			Id:              message.ID,
			ConversationId:  message.ConversationID,
			SenderAccountId: message.SenderAccountID,
			Content:         message.Content,
			CreatedAt:       timestamppb.New(message.CreatedAt),
		},
	}, nil
}

func (server *conversationsAPI) UpdateConversationMessage(ctx context.Context, in *accountsv1.UpdateConversationMessageRequest) (*accountsv1.UpdateConversationMessageResponse, error) {
	token, err := server.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	err = validators.ValidateUpdateConversationMessageRequest(in)
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

	message, err := server.messageRepo.Get(ctx, &models.OneConversationMessageFilter{ID: in.MessageId, ConversationID: in.ConversationId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	if message.SenderAccountID != accountId {
		return nil, status.Error(codes.InvalidArgument, "sender is not the owner of the message")
	}

	filter := models.OneConversationMessageFilter{ID: in.MessageId, ConversationID: in.ConversationId}

	updatedMessage, err := server.messageRepo.Update(ctx, &filter, &models.UpdateConversationMessagePayload{Content: in.Content})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	return &accountsv1.UpdateConversationMessageResponse{
		Message: &accountsv1.ConversationMessage{
			Id:              updatedMessage.ID,
			ConversationId:  updatedMessage.ConversationID,
			SenderAccountId: updatedMessage.SenderAccountID,
			Content:         updatedMessage.Content,
			CreatedAt:       timestamppb.New(updatedMessage.CreatedAt),
		},
	}, nil
}

func (server *conversationsAPI) ListConversationMessages(ctx context.Context, in *accountsv1.ListConversationMessagesRequest) (*accountsv1.ListConversationMessagesResponse, error) {
	token, err := server.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	err = validators.ValidateListConversationMessageRequest(in)
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

	if in.Limit == 0 {
		in.Limit = 20
	}

	messages, err := server.messageRepo.List(ctx, &models.ManyConversationMessagesFilter{
		ConversationID: in.ConversationId,
	}, &models.Pagination{Offset: int64(in.Offset), Limit: int64(in.Limit)})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	var conversationMessages []*accountsv1.ConversationMessage

	for _, message := range messages {
		conversationMessage := &accountsv1.ConversationMessage{Id: message.ID, ConversationId: message.ConversationID, SenderAccountId: message.SenderAccountID, Content: message.Content, CreatedAt: timestamppb.New(message.CreatedAt)}
		conversationMessages = append(conversationMessages, conversationMessage)
	}

	return &accountsv1.ListConversationMessagesResponse{
		Messages: conversationMessages,
	}, nil
}
