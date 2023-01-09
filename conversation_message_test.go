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

type ConversationMessagesAPISuite struct {
	suite.Suite
	auth auth.TestService
	srv  *conversationsAPI
}

func TestConversationMessagesService(t *testing.T) {
	suite.Run(t, &ConversationMessagesAPISuite{})
}

func (s *ConversationMessagesAPISuite) SetupSuite() {
	logger := newLoggerOrFail(s.T())
	newMemoryTestDatabaseOrFail := newDatabaseOrFail(s.T(), logger)

	s.srv = &conversationsAPI{
		auth:        &s.auth,
		logger:      logger,
		messageRepo: memory.NewConversationMessagesRepository(newMemoryTestDatabaseOrFail, logger),
		repo:        memory.NewConversationsRepository(newMemoryTestDatabaseOrFail, logger),
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

func CreateDefaultGroup(s *ConversationMessagesAPISuite, ctx context.Context) *accountsv1.Group {
	groupRes, err := s.srv.groupService.CreateGroup(ctx, &accountsv1.CreateGroupRequest{
		Name:        "name",
		Description: "desc"},
	)
	s.Require().NoError(err)
	s.Require().NotNil(groupRes)
	s.Require().NotNil(groupRes.Group)
	return groupRes.Group
}

func CreateGroupAndGetConversation(s *ConversationMessagesAPISuite, ctx context.Context) *accountsv1.Conversation {
	group := CreateDefaultGroup(s, ctx)
	listedConv, err := s.srv.ListConversations(ctx, &accountsv1.ListConversationsRequest{GroupId: group.Id})
	s.Require().NoError(err)

	return listedConv.Conversations[0]
}

func CreateContext(s *ConversationMessagesAPISuite) *context.Context {
	uuid, err := uuid.NewRandom()
	s.Require().NoError(err)
	ctx, err := s.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uuid})
	s.Require().NoError(err)
	return &ctx
}

func (s *ConversationMessagesAPISuite) TestSendMessageNoError() {
	ctx := CreateContext(s)
	conversation := CreateGroupAndGetConversation(s, *ctx)

	message, err := s.srv.SendConversationMessage(*ctx, &accountsv1.SendConversationMessageRequest{ConversationId: conversation.Id, Content: "Bonjour ceci est un message test"})
	s.Require().NoError(err)
	s.Require().Equal(message.Message.Content, "Bonjour ceci est un message test")
}

func (s *ConversationMessagesAPISuite) TestSendMessageTooLong() {
	ctx := CreateContext(s)
	conversation := CreateGroupAndGetConversation(s, *ctx)

	_, err := s.srv.SendConversationMessage(*ctx, &accountsv1.SendConversationMessageRequest{ConversationId: conversation.Id,
		Content: "Bonjour ceci est un message test beaucoup trop long, il faut que je dépasse les 250 characters du coup j'vais raconter ma vie: Thomas, 23 and, developpeurs, voila voila sinon il fait froid et ce matin au 7/11 il y avait plus qu'un café ca ma un peu cassé les ****** parce que normalement j'en prend deux et pas un mais bon je me suis rabattu sur le jus d'orange. Bon inchalla ca fait plus de 250 characters maintenant"})
	s.Require().Error(err)
}

func (s *ConversationMessagesAPISuite) TestSendEmptyMessage() {
	ctx := CreateContext(s)
	conversation := CreateGroupAndGetConversation(s, *ctx)

	_, err := s.srv.SendConversationMessage(*ctx, &accountsv1.SendConversationMessageRequest{ConversationId: conversation.Id, Content: ""})
	s.Require().Error(err)
}

func (s *ConversationMessagesAPISuite) TestDeleteMessageNoError() {
	ctx := CreateContext(s)
	conversation := CreateGroupAndGetConversation(s, *ctx)

	message, err := s.srv.SendConversationMessage(*ctx, &accountsv1.SendConversationMessageRequest{ConversationId: conversation.Id, Content: "test"})
	s.Require().NoError(err)

	_, err = s.srv.DeleteConversationMessage(*ctx, &accountsv1.DeleteConversationMessageRequest{MessageId: message.Message.Id, ConversationId: conversation.Id})
	s.Require().NoError(err)
}

func (s *ConversationMessagesAPISuite) TestDeleteMessageTwoTimes() {
	ctx := CreateContext(s)
	conversation := CreateGroupAndGetConversation(s, *ctx)

	message, err := s.srv.SendConversationMessage(*ctx, &accountsv1.SendConversationMessageRequest{ConversationId: conversation.Id, Content: "test"})
	s.Require().NoError(err)

	_, err = s.srv.DeleteConversationMessage(*ctx, &accountsv1.DeleteConversationMessageRequest{MessageId: message.Message.Id, ConversationId: message.Message.ConversationId})
	s.Require().NoError(err)

	_, err = s.srv.DeleteConversationMessage(*ctx, &accountsv1.DeleteConversationMessageRequest{MessageId: message.Message.Id, ConversationId: message.Message.ConversationId})
	s.Require().Error(err)
}

func (s *ConversationMessagesAPISuite) TestGetMessageNoError() {
	ctx := CreateContext(s)
	conversation := CreateGroupAndGetConversation(s, *ctx)

	message, err := s.srv.SendConversationMessage(*ctx, &accountsv1.SendConversationMessageRequest{ConversationId: conversation.Id, Content: "test"})
	s.Require().NoError(err)

	getMessage, err := s.srv.GetConversationMessage(*ctx, &accountsv1.GetConversationMessageRequest{MessageId: message.Message.Id, ConversationId: message.Message.ConversationId})
	s.Require().NoError(err)
	s.Require().Equal(message.Message.Content, getMessage.Message.Content)
}

func (s *ConversationMessagesAPISuite) TestGetMessageBadUUID() {
	ctx := CreateContext(s)

	id, _ := uuid.NewRandom()

	_, err := s.srv.GetConversationMessage(*ctx, &accountsv1.GetConversationMessageRequest{MessageId: id.String()})
	s.Require().Error(err)
}

func (s *ConversationMessagesAPISuite) TestListMessagesNoError() {
	ctx := CreateContext(s)
	conversation := CreateGroupAndGetConversation(s, *ctx)

	_, err := s.srv.SendConversationMessage(*ctx, &accountsv1.SendConversationMessageRequest{ConversationId: conversation.Id, Content: "test_1"})
	s.Require().NoError(err)
	_, err = s.srv.SendConversationMessage(*ctx, &accountsv1.SendConversationMessageRequest{ConversationId: conversation.Id, Content: "test_2"})
	s.Require().NoError(err)
	_, err = s.srv.SendConversationMessage(*ctx, &accountsv1.SendConversationMessageRequest{ConversationId: conversation.Id, Content: "test_3"})
	s.Require().NoError(err)

	listedMessages, err := s.srv.ListConversationMessages(*ctx, &accountsv1.ListConversationMessagesRequest{ConversationId: conversation.Id, Limit: 0, Offset: 0})
	s.Require().NoError(err)
	s.Require().Equal(len(listedMessages.Messages), 3)
}

func (s *ConversationMessagesAPISuite) TestListMessagesFromNonExistingConversation() {
	ctx := CreateContext(s)
	// _ := CreateGroupAndGetConversation(s, *ctx)

	id, _ := uuid.NewRandom()

	_, err := s.srv.ListConversationMessages(*ctx, &accountsv1.ListConversationMessagesRequest{ConversationId: id.String(), Limit: 0, Offset: 0})
	s.Require().Error(err)
}

// 	s.Require().NoError(err)

// 	group := ConvCreateDefaultGroup(s, ctx)

// 	listedConv, err := s.srv.ListConversations(ctx, &accountsv1.ListConversationsRequest{GroupId: group.Id})
// 	s.Require().NoError(err)

// 	getConv, err := s.srv.GetConversation(ctx, &accountsv1.GetConversationRequest{ConversationId: listedConv.Conversations[0].Id})
// 	s.Require().NoError(err)
// 	s.Require().Equal(getConv.Conversation.Title, "General conversation")
// }

// func (s *ConversationMessagesAPISuite) TestGetConversationWithRandomUUID() {
// 	id, err := uuid.NewRandom()
// 	s.Require().NoError(err)
// 	ctx, err := s.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: id})
// 	s.Require().NoError(err)

// 	randomUUID, err := uuid.NewRandom()
// 	s.Require().NoError(err)

// 	_, err = s.srv.GetConversation(ctx, &accountsv1.GetConversationRequest{ConversationId: randomUUID.String()})
// 	s.Require().Error(err)
// }

// func (s *ConversationMessagesAPISuite) TestUpdateConversation() {
// 	uuid, err := uuid.NewRandom()
// 	s.Require().NoError(err)
// 	ctx, err := s.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uuid})
// 	s.Require().NoError(err)

// 	group := ConvCreateDefaultGroup(s, ctx)

// 	listedConv, err := s.srv.ListConversations(ctx, &accountsv1.ListConversationsRequest{GroupId: group.Id})
// 	s.Require().NoError(err)

// 	updatedConv, err := s.srv.UpdateConversation(ctx, &accountsv1.UpdateConversationRequest{ConversationId: listedConv.Conversations[0].Id, Title: "New title for test"})
// 	s.Require().NoError(err)
// 	s.Require().Equal(updatedConv.Conversation.Title, "New title for test")
// }

// func (s *ConversationMessagesAPISuite) TestUpdateConversationWithTooLongTitle() {
// 	uuid, err := uuid.NewRandom()
// 	s.Require().NoError(err)
// 	ctx, err := s.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uuid})
// 	s.Require().NoError(err)

// 	group := ConvCreateDefaultGroup(s, ctx)

// 	listedConv, err := s.srv.ListConversations(ctx, &accountsv1.ListConversationsRequest{GroupId: group.Id})
// 	s.Require().NoError(err)

// 	_, err = s.srv.UpdateConversation(ctx, &accountsv1.UpdateConversationRequest{ConversationId: listedConv.Conversations[0].Id, Title: "New title that should be too long to pass the test and if it work it is not normal (call 0646294625 to complain)"})
// 	s.Require().Error(err)
// }
