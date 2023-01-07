package main

import (
	"accounts-service/auth"
	"accounts-service/models"
	"accounts-service/models/memory"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type ConversationsAPISuite struct {
	suite.Suite
	account *accountsAPI
	srv     *conversationsAPI
}

func TestConversationsService(t *testing.T) {
	suite.Run(t, &ConversationsAPISuite{})
}

func (s *ConversationsAPISuite) SetupSuite() {
	logger := newLoggerOrFail(s.T())
	db := newDatabaseOrFail(s.T(), logger)

	s.srv = &conversationsAPI{
		auth:   auth.NewService(genKeyOrFail(s.T())),
		logger: logger,
		repo:   memory.NewConversationsRepository(db, logger),
	}

	s.account = &accountsAPI{
		auth:   s.srv.auth,
		logger: s.srv.logger,
		repo:   memory.NewAccountsRepository(db, logger),
	}

	s.srv.groupService = &groupsAPI{
		auth:       s.srv.auth,
		logger:     s.srv.logger,
		groupRepo:  memory.NewGroupsRepository(db, logger),
		memberRepo: memory.NewMembersRepository(db, logger),
		noteRepo:   nil,
	}
}

func ConvCreateUser(s *ConversationsAPISuite, name string, email string, hash []byte) (*models.Account, error) {
	return s.account.repo.Create(context.TODO(), &models.AccountPayload{Name: &name, Email: &email, Hash: &hash})
}

func createRandomAccount(s *ConversationsAPISuite) *models.Account {
	first, err := ConvCreateUser(s, "first", fmt.Sprint(randomString(), randomString(), "@email.fr"), randomString())
	s.Require().NoError(err)
	s.Require().NotNil(first)
	return first
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
	account := createRandomAccount(s)

	uid := uuid.MustParse(account.ID)
	ctx, err := s.srv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	group := ConvCreateDefaultGroup(s, ctx)
	listedConv, err := s.srv.ListConversations(ctx, &accountsv1.ListConversationsRequest{GroupId: group.Id})
	s.Require().NoError(err)
	s.Require().Equal(listedConv.Conversations[0].Title, "General conversation")
}

func (s *ConversationsAPISuite) TestCreateConversation() {
	account := createRandomAccount(s)

	uid := uuid.MustParse(account.ID)
	ctx, err := s.srv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	group := ConvCreateDefaultGroup(s, ctx)
	nConv, err := s.srv.CreateConversation(ctx, &accountsv1.CreateConversationRequest{GroupId: group.Id, Title: "Test"})
	s.Require().NoError(err)
	s.Require().Equal(nConv.Conversation.Title, "Test")

	listedConv, err := s.srv.ListConversations(ctx, &accountsv1.ListConversationsRequest{GroupId: group.Id})
	s.Require().NoError(err)
	s.Require().Equal(len(listedConv.Conversations), 1)
}

func (s *ConversationsAPISuite) TestDeleteConversation() {
	account := createRandomAccount(s)

	uid := uuid.MustParse(account.ID)
	ctx, err := s.srv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
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
