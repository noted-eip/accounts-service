package main

import (
	"accounts-service/auth"
	"accounts-service/models/memory"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type ConversationsAPISuite struct {
	suite.Suite
	auth auth.TestService
	srv  *conversationsAPI
}

func TestConversationsService(t *testing.T) {
	suite.Run(t, &ConversationsAPISuite{})
}

func (s *ConversationsAPISuite) SetupSuite() {
	logger := newLoggerOrFail(s.T())
	newMemoryTestDatabaseOrFail := newDatabaseOrFail(s.T(), logger)

	s.srv = &conversationsAPI{
		auth:   &s.auth,
		logger: logger,
		repo:   memory.NewConversationsRepository(newMemoryTestDatabaseOrFail, logger),
	}

	s.srv.groupService = &groupsAPI{
		auth:             s.srv.auth,
		logger:           s.srv.logger,
		groupRepo:        memory.NewGroupsRepository(newMemoryTestDatabaseOrFail, logger),
		memberRepo:       memory.NewMembersRepository(newMemoryTestDatabaseOrFail, logger),
		conversationRepo: s.srv.repo,
		noteRepo:         nil,
	}
}

func ConvCreateDefaultGroup(s *ConversationsAPISuite, ctx context.Context) *accountsv1.Group {
	groupRes, err := s.srv.groupService.CreateGroup(ctx, &accountsv1.CreateGroupRequest{
		Name:        "name",
		Description: "desc"},
	)
	s.Require().NoError(err)
	s.Require().NotNil(groupRes)
	s.Require().NotNil(groupRes.Group)
	return groupRes.Group
}

func (s *ConversationsAPISuite) TestCreateGroupWithDefaultConversation() {
	uuid, err := uuid.NewRandom()
	s.Require().NoError(err)
	ctx, err := s.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uuid})
	s.Require().NoError(err)

	group := ConvCreateDefaultGroup(s, ctx)
	listedConv, err := s.srv.ListConversations(ctx, &accountsv1.ListConversationsRequest{GroupId: group.Id})
	s.Require().NoError(err)
	s.Require().Equal(listedConv.Conversations[0].Title, "General conversation")
}

func (s *ConversationsAPISuite) TestCreateConversation() {
	uuid, err := uuid.NewRandom()
	s.Require().NoError(err)
	ctx, err := s.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uuid})
	s.Require().NoError(err)

	group := ConvCreateDefaultGroup(s, ctx)
	nConv, err := s.srv.CreateConversation(ctx, &accountsv1.CreateConversationRequest{GroupId: group.Id, Title: "Test"})
	s.Require().NoError(err)
	s.Require().Equal(nConv.Conversation.Title, "Test")

	listedConv, err := s.srv.ListConversations(ctx, &accountsv1.ListConversationsRequest{GroupId: group.Id})
	s.Require().NoError(err)
	s.Require().Equal(len(listedConv.Conversations), 2)
}

func (s *ConversationsAPISuite) TestDeleteConversation() {
	uuid, err := uuid.NewRandom()
	s.Require().NoError(err)
	ctx, err := s.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uuid})
	s.Require().NoError(err)

	group := ConvCreateDefaultGroup(s, ctx)

	listedConv, err := s.srv.ListConversations(ctx, &accountsv1.ListConversationsRequest{GroupId: group.Id})
	s.Require().NoError(err)

	_, err = s.srv.DeleteConversation(ctx, &accountsv1.DeleteConversationRequest{ConversationId: listedConv.Conversations[0].Id})
	s.Require().NoError(err)

	updatedListedConv, err := s.srv.ListConversations(ctx, &accountsv1.ListConversationsRequest{GroupId: group.Id})
	s.Require().NoError(err)
	s.Require().Equal(len(updatedListedConv.Conversations), 0)
}

func (s *ConversationsAPISuite) TestGetConversation() {
	uuid, err := uuid.NewRandom()
	s.Require().NoError(err)
	ctx, err := s.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uuid})
	s.Require().NoError(err)

	group := ConvCreateDefaultGroup(s, ctx)

	listedConv, err := s.srv.ListConversations(ctx, &accountsv1.ListConversationsRequest{GroupId: group.Id})
	s.Require().NoError(err)

	getConv, err := s.srv.GetConversation(ctx, &accountsv1.GetConversationRequest{ConversationId: listedConv.Conversations[0].Id})
	s.Require().NoError(err)
	s.Require().Equal(getConv.Conversation.Title, "General conversation")
}

func (s *ConversationsAPISuite) TestGetConversationWithRandomUUID() {
	id, err := uuid.NewRandom()
	s.Require().NoError(err)
	ctx, err := s.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: id})
	s.Require().NoError(err)

	randomUUID, err := uuid.NewRandom()
	s.Require().NoError(err)

	_, err = s.srv.GetConversation(ctx, &accountsv1.GetConversationRequest{ConversationId: randomUUID.String()})
	s.Require().Error(err)
}

func (s *ConversationsAPISuite) TestUpdateConversation() {
	uuid, err := uuid.NewRandom()
	s.Require().NoError(err)
	ctx, err := s.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uuid})
	s.Require().NoError(err)

	group := ConvCreateDefaultGroup(s, ctx)

	listedConv, err := s.srv.ListConversations(ctx, &accountsv1.ListConversationsRequest{GroupId: group.Id})
	s.Require().NoError(err)

	updatedConv, err := s.srv.UpdateConversation(ctx, &accountsv1.UpdateConversationRequest{ConversationId: listedConv.Conversations[0].Id, Title: "New title for test"})
	s.Require().NoError(err)
	s.Require().Equal(updatedConv.Conversation.Title, "New title for test")
}

func (s *ConversationsAPISuite) TestUpdateConversationWithTooLongTitle() {
	uuid, err := uuid.NewRandom()
	s.Require().NoError(err)
	ctx, err := s.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uuid})
	s.Require().NoError(err)

	group := ConvCreateDefaultGroup(s, ctx)

	listedConv, err := s.srv.ListConversations(ctx, &accountsv1.ListConversationsRequest{GroupId: group.Id})
	s.Require().NoError(err)

	_, err = s.srv.UpdateConversation(ctx, &accountsv1.UpdateConversationRequest{ConversationId: listedConv.Conversations[0].Id, Title: "New title that should be too long to pass the test and if it work it is not normal (call 0646294625 to complain)"})
	s.Require().Error(err)
}
