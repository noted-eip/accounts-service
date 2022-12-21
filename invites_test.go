package main

import (
	"accounts-service/auth"
	"accounts-service/models"
	"accounts-service/models/memory"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type InvitesAPISuite struct {
	suite.Suite
	srv *invitesAPI
}

func TestInvitesService(t *testing.T) {
	suite.Run(t, &InvitesAPISuite{})
}

func (s *InvitesAPISuite) SetupSuite() {
	logger := newLoggerOrFail(s.T())
	db := newDatabaseOrFail(s.T(), logger)

	s.srv = &invitesAPI{
		auth:        auth.NewService(genKeyOrFail(s.T())),
		logger:      logger,
		inviteRepo:  memory.NewInvitesRepository(db, logger),
		accountRepo: memory.NewAccountsRepository(db, logger),
		groupRepo:   memory.NewGroupsRepository(db, logger),
	}
	s.srv.groupService = &groupsAPI{
		auth:       s.srv.auth,
		logger:     s.srv.logger,
		groupRepo:  s.srv.groupRepo,
		memberRepo: memory.NewMembersRepository(db, logger),
		noteRepo:   nil,
	}
}

func createUser(s *InvitesAPISuite, name string, email string, hash []byte) (*models.Account, error) {
	return s.srv.accountRepo.Create(context.TODO(), &models.AccountPayload{Name: &name, Email: &email, Hash: &hash})
}

func randomString() []byte {
	return []byte(fmt.Sprint(rand.Int()))
}

// NOTE: Yes I could have made a function that takes a int and gives me back an array with x accounts, but is it worth it and is life even worth it anyway ?
func createTwoRandomAccounts(s *InvitesAPISuite) (*models.Account, *models.Account) {
	first, err := createUser(s, "first", fmt.Sprint(randomString(), randomString(), "@email.fr"), randomString())
	s.Require().NoError(err)
	s.Require().NotNil(first)
	second, err := createUser(s, "second", fmt.Sprint(randomString(), randomString(), "@email.com"), randomString())
	s.Require().NoError(err)
	s.Require().NotNil(second)
	return first, second
}

func createDefaultGroup(s *InvitesAPISuite, ctx context.Context) *accountsv1.Group {
	groupRes, err := s.srv.groupService.CreateGroup(ctx, &accountsv1.CreateGroupRequest{
		Name:        "name",
		Description: "desc"},
	)
	s.Require().NoError(err)
	s.Require().NotNil(groupRes)
	s.Require().NotNil(groupRes.Group)
	return groupRes.Group
}

func (s *InvitesAPISuite) TestValidSendInvite() {
	sender, recipient := createTwoRandomAccounts(s)

	uid := uuid.MustParse(sender.ID)
	ctx, err := s.srv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	group := createDefaultGroup(s, ctx)

	invite, err := s.srv.SendInvite(ctx, &accountsv1.SendInviteRequest{GroupId: group.Id, RecipientAccountId: recipient.ID})

	s.Require().NoError(err)
	s.NotNil(invite)
	s.Equal(invite.Invite.GroupId, group.Id)
	s.Equal(invite.Invite.RecipientAccountId, recipient.ID)
	s.Equal(invite.Invite.SenderAccountId, sender.ID)

	getInviteRes, err := s.srv.GetInvite(ctx, &accountsv1.GetInviteRequest{InviteId: invite.Invite.Id})
	s.Require().NoError(err)
	s.Equal(getInviteRes.Invite.GroupId, group.Id)
	s.Equal(getInviteRes.Invite.RecipientAccountId, recipient.ID)
	s.Equal(getInviteRes.Invite.SenderAccountId, sender.ID)
}

func (s *InvitesAPISuite) TestSendInviteErrorSenderNotPartOfGroup() {
	sender, recipient := createTwoRandomAccounts(s)

	uid := uuid.MustParse(sender.ID)
	ctx, err := s.srv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	group := createDefaultGroup(s, ctx)

	uid = uuid.MustParse(recipient.ID)
	ctx, err = s.srv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	invite, err := s.srv.SendInvite(ctx, &accountsv1.SendInviteRequest{GroupId: group.Id, RecipientAccountId: recipient.ID})

	s.Require().Error(err)
	s.Require().Nil(invite)
}

func (s *InvitesAPISuite) TestSendInviteErrorGroupDoesNotExist() {
	sender, recipient := createTwoRandomAccounts(s)

	uid := uuid.MustParse(sender.ID)
	ctx, err := s.srv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	invite, err := s.srv.SendInvite(ctx, &accountsv1.SendInviteRequest{GroupId: uuid.New().String(), RecipientAccountId: recipient.ID})

	s.Require().Error(err)
	s.Require().Nil(invite)
}

func (s *InvitesAPISuite) TestSendInviteErrorRecipientDoesNotExist() {
	sender, _ := createTwoRandomAccounts(s)

	uid := uuid.MustParse(sender.ID)
	ctx, err := s.srv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	group := createDefaultGroup(s, ctx)

	invite, err := s.srv.SendInvite(ctx, &accountsv1.SendInviteRequest{GroupId: group.Id, RecipientAccountId: uuid.New().String()})
	s.Require().Error(err)
	s.Require().Nil(invite)
}

func (s *InvitesAPISuite) TestAcceptInvite() {
	sender, recipient := createTwoRandomAccounts(s)

	uid := uuid.MustParse(sender.ID)
	ctx, err := s.srv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	group := createDefaultGroup(s, ctx)

	invite, err := s.srv.SendInvite(ctx, &accountsv1.SendInviteRequest{GroupId: group.Id, RecipientAccountId: recipient.ID})
	s.Require().NoError(err)

	uid = uuid.MustParse(recipient.ID)
	ctx, err = s.srv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	acceptInviteRes, err := s.srv.AcceptInvite(ctx, &accountsv1.AcceptInviteRequest{InviteId: invite.Invite.Id})
	s.Require().NoError(err)
	s.Require().NotNil(acceptInviteRes)

	_, err = s.srv.GetInvite(ctx, &accountsv1.GetInviteRequest{InviteId: invite.Invite.Id})
	s.Require().Error(err)

	member, err := s.srv.groupService.GetGroupMember(ctx, &accountsv1.GetGroupMemberRequest{GroupId: group.Id, AccountId: recipient.ID})
	s.Require().NoError(err)
	s.Require().Equal(member.Member.AccountId, recipient.ID)
}

func (s *InvitesAPISuite) TestDenyInvite() {
	sender, recipient := createTwoRandomAccounts(s)

	uid := uuid.MustParse(sender.ID)
	ctx, err := s.srv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	group := createDefaultGroup(s, ctx)

	invite, err := s.srv.SendInvite(ctx, &accountsv1.SendInviteRequest{GroupId: group.Id, RecipientAccountId: recipient.ID})
	s.Require().NoError(err)

	uid = uuid.MustParse(recipient.ID)
	ctx, err = s.srv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	denyRequestRes, err := s.srv.DenyInvite(ctx, &accountsv1.DenyInviteRequest{InviteId: invite.Invite.Id})
	s.Require().NoError(err)
	s.Require().NotNil(denyRequestRes)

	_, err = s.srv.GetInvite(ctx, &accountsv1.GetInviteRequest{InviteId: invite.Invite.Id})
	s.Require().Error(err)

	member, err := s.srv.groupService.GetGroupMember(ctx, &accountsv1.GetGroupMemberRequest{GroupId: group.Id, AccountId: recipient.ID})
	s.Require().Error(err)
	s.Require().Nil(member)
}

func (s *InvitesAPISuite) TestAcceptInviteErrorNotYourInvite() {
	sender, recipient := createTwoRandomAccounts(s)

	uid := uuid.MustParse(sender.ID)
	ctx, err := s.srv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	group := createDefaultGroup(s, ctx)

	invite, err := s.srv.SendInvite(ctx, &accountsv1.SendInviteRequest{GroupId: group.Id, RecipientAccountId: recipient.ID})
	s.Require().NoError(err)

	acceptInviteRes, err := s.srv.AcceptInvite(ctx, &accountsv1.AcceptInviteRequest{InviteId: invite.Invite.Id})
	s.Require().Error(err)
	s.Require().Nil(acceptInviteRes)
}

func (s *InvitesAPISuite) TestDenyInviteErrorNotYourInvite() {
	sender, recipient := createTwoRandomAccounts(s)

	uid := uuid.MustParse(sender.ID)
	ctx, err := s.srv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	group := createDefaultGroup(s, ctx)

	invite, err := s.srv.SendInvite(ctx, &accountsv1.SendInviteRequest{GroupId: group.Id, RecipientAccountId: recipient.ID})
	s.Require().NoError(err)

	denyInviteRes, err := s.srv.DenyInvite(ctx, &accountsv1.DenyInviteRequest{InviteId: invite.Invite.Id})
	s.Require().Error(err)
	s.Require().Nil(denyInviteRes)
}

func (s *InvitesAPISuite) TestAcceptNonExistingInvite() {
	recipient, _ := createTwoRandomAccounts(s)

	uid := uuid.MustParse(recipient.ID)
	ctx, err := s.srv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	acceptInviteRes, err := s.srv.AcceptInvite(ctx, &accountsv1.AcceptInviteRequest{InviteId: uuid.New().String()})
	s.Require().Error(err)
	s.Require().Nil(acceptInviteRes)
}

func (s *InvitesAPISuite) TestDenyNonExistingInvite() {
	recipient, _ := createTwoRandomAccounts(s)

	uid := uuid.MustParse(recipient.ID)
	ctx, err := s.srv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	denyInviteRes, err := s.srv.DenyInvite(ctx, &accountsv1.DenyInviteRequest{InviteId: uuid.New().String()})
	s.Require().Error(err)
	s.Require().Nil(denyInviteRes)
}
