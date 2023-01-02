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

type GroupsAPISuite struct {
	suite.Suite
	groupSrv   *groupsAPI
	accountSrv *accountsAPI
}

func TestGroupsService(t *testing.T) {
	suite.Run(t, new(GroupsAPISuite))
}

func (s *GroupsAPISuite) SetupSuite() {
	logger := newLoggerOrFail(s.T())
	db := newDatabaseOrFail(s.T(), logger)

	s.accountSrv = &accountsAPI{
		auth:   auth.NewService(genKeyOrFail(s.T())),
		logger: logger,
		repo:   memory.NewAccountsRepository(db, logger),
	}

	s.groupSrv = &groupsAPI{
		auth:       auth.NewService(genKeyOrFail(s.T())),
		logger:     logger,
		groupRepo:  memory.NewGroupsRepository(db, logger),
		memberRepo: memory.NewMembersRepository(db, logger),
	}
}

func (s *GroupsAPISuite) TestCreateGroup() {
	createAccountRes, err := s.accountSrv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "maxime.dodin@gmail.com", Password: "1234", Name: "Maxime"})
	s.Require().NoError(err)

	uid := uuid.MustParse(createAccountRes.Account.Id)

	ctx, err := s.groupSrv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	createGroupRes, err := s.groupSrv.CreateGroup(ctx, &accountsv1.CreateGroupRequest{Description: "description", Name: "EIP"})
	s.Require().NoError(err)

	getGroupMemberRes, err := s.groupSrv.GetGroupMember(ctx, &accountsv1.GetGroupMemberRequest{GroupId: createGroupRes.Group.Id, AccountId: uid.String()})
	s.Require().NoError(err)
	s.Equal("admin", getGroupMemberRes.Member.Role)
	s.Equal(uid.String(), getGroupMemberRes.Member.AccountId)
}

func (s *GroupsAPISuite) TestAddMembersToGroup() {
	createAccountMaxRes, err := s.accountSrv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "maxime@gmail.com", Password: "1234", Name: "Maxime"})
	s.Require().NoError(err)

	createAccountGabiRes, err := s.accountSrv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "gabriel@gmail.com", Password: "1234", Name: "Gabriel"})
	s.Require().NoError(err)

	createAccountBalthiRes, err := s.accountSrv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "balthazard@gmail.com", Password: "1234", Name: "Balthazard"})
	s.Require().NoError(err)

	uid := uuid.MustParse(createAccountMaxRes.Account.Id)

	ctx, err := s.groupSrv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	createGroupRes, err := s.groupSrv.CreateGroup(ctx, &accountsv1.CreateGroupRequest{Description: "description", Name: "EIP"})
	s.Require().NoError(err)

	_, err = s.groupSrv.AddGroupMember(ctx, &accountsv1.AddGroupMemberRequest{GroupId: createGroupRes.Group.Id, AccountId: createAccountBalthiRes.Account.Id})
	s.Require().NoError(err)

	_, err = s.groupSrv.AddGroupMember(ctx, &accountsv1.AddGroupMemberRequest{GroupId: createGroupRes.Group.Id, AccountId: createAccountGabiRes.Account.Id})
	s.Require().NoError(err)

	listGroupMemberResp, err := s.groupSrv.ListGroupMembers(ctx, &accountsv1.ListGroupMembersRequest{GroupId: createGroupRes.Group.Id, Limit: 10, Offset: 0})
	s.Require().NoError(err)

	s.Require().Equal(3, len(listGroupMemberResp.Members))
}

func (s *GroupsAPISuite) TestDeleteGroup() {
	createAccountMaxRes, err := s.accountSrv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "maxime@gmail.com", Password: "1234", Name: "Maxime"})
	s.Require().NoError(err)

	createAccountGabiRes, err := s.accountSrv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "gabriel@gmail.com", Password: "1234", Name: "Gabriel"})
	s.Require().NoError(err)

	createAccountBalthiRes, err := s.accountSrv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "balthazard@gmail.com", Password: "1234", Name: "Balthazard"})
	s.Require().NoError(err)

	uid := uuid.MustParse(createAccountMaxRes.Account.Id)

	ctx, err := s.groupSrv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	createGroupRes, err := s.groupSrv.CreateGroup(ctx, &accountsv1.CreateGroupRequest{Description: "description", Name: "EIP"})
	s.Require().NoError(err)

	_, err = s.groupSrv.AddGroupMember(ctx, &accountsv1.AddGroupMemberRequest{GroupId: createGroupRes.Group.Id, AccountId: createAccountBalthiRes.Account.Id})
	s.Require().NoError(err)

	_, err = s.groupSrv.AddGroupMember(ctx, &accountsv1.AddGroupMemberRequest{GroupId: createGroupRes.Group.Id, AccountId: createAccountGabiRes.Account.Id})
	s.Require().NoError(err)

	_, err = s.groupSrv.DeleteGroup(ctx, &accountsv1.DeleteGroupRequest{GroupId: createGroupRes.Group.Id})
	s.Require().NoError(err)

	listGroupMemberResp, err := s.groupSrv.ListGroupMembers(ctx, &accountsv1.ListGroupMembersRequest{GroupId: createGroupRes.Group.Id, Limit: 10, Offset: 0})
	s.Require().NoError(err)

	s.Require().Equal(0, len(listGroupMemberResp.Members))
}

func (s *GroupsAPISuite) TestLeaveGroupAsAdmin() {
	createAccountMaxRes, err := s.accountSrv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "maxime@gmail.com", Password: "1234", Name: "Maxime"})
	s.Require().NoError(err)

	createAccountGabiRes, err := s.accountSrv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "gabriel@gmail.com", Password: "1234", Name: "Gabriel"})
	s.Require().NoError(err)

	createAccountBalthiRes, err := s.accountSrv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "balthazard@gmail.com", Password: "1234", Name: "Balthazard"})
	s.Require().NoError(err)

	uid := uuid.MustParse(createAccountMaxRes.Account.Id)

	ctx, err := s.groupSrv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	createGroupRes, err := s.groupSrv.CreateGroup(ctx, &accountsv1.CreateGroupRequest{Description: "description", Name: "EIP"})
	s.Require().NoError(err)

	_, err = s.groupSrv.AddGroupMember(ctx, &accountsv1.AddGroupMemberRequest{GroupId: createGroupRes.Group.Id, AccountId: createAccountBalthiRes.Account.Id})
	s.Require().NoError(err)

	_, err = s.groupSrv.AddGroupMember(ctx, &accountsv1.AddGroupMemberRequest{GroupId: createGroupRes.Group.Id, AccountId: createAccountGabiRes.Account.Id})
	s.Require().NoError(err)

	_, err = s.groupSrv.RemoveGroupMember(ctx, &accountsv1.RemoveGroupMemberRequest{GroupId: createGroupRes.Group.Id, AccountId: createAccountMaxRes.Account.Id})
	s.Require().NoError(err)

	listGroupMemberResp, err := s.groupSrv.ListGroupMembers(ctx, &accountsv1.ListGroupMembersRequest{GroupId: createGroupRes.Group.Id, Limit: 10, Offset: 0})
	s.Require().NoError(err)

	isAdmin := false
	s.Require().Equal(2, len(listGroupMemberResp.Members))
	for _, gm := range listGroupMemberResp.Members {
		if gm.Role == auth.RoleAdmin {
			isAdmin = true
		}
	}
	s.Require().True(isAdmin)
}

func (s *GroupsAPISuite) TestRemoveAdminMemberAsUser() {
	createAccountMaxRes, err := s.accountSrv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "maxime@gmail.com", Password: "1234", Name: "Maxime"})
	s.Require().NoError(err)

	createAccountGabiRes, err := s.accountSrv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "gabriel@gmail.com", Password: "1234", Name: "Gabriel"})
	s.Require().NoError(err)

	createAccountBalthiRes, err := s.accountSrv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "balthazard@gmail.com", Password: "1234", Name: "Balthazard"})
	s.Require().NoError(err)

	uid := uuid.MustParse(createAccountMaxRes.Account.Id)

	ctx, err := s.groupSrv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	createGroupRes, err := s.groupSrv.CreateGroup(ctx, &accountsv1.CreateGroupRequest{Description: "description", Name: "EIP"})
	s.Require().NoError(err)

	_, err = s.groupSrv.AddGroupMember(ctx, &accountsv1.AddGroupMemberRequest{GroupId: createGroupRes.Group.Id, AccountId: createAccountBalthiRes.Account.Id})
	s.Require().NoError(err)

	_, err = s.groupSrv.AddGroupMember(ctx, &accountsv1.AddGroupMemberRequest{GroupId: createGroupRes.Group.Id, AccountId: createAccountGabiRes.Account.Id})
	s.Require().NoError(err)

	uidBalthi := uuid.MustParse(createAccountBalthiRes.Account.Id)

	ctxBalthi, err := s.groupSrv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uidBalthi})
	s.Require().NoError(err)

	_, err = s.groupSrv.RemoveGroupMember(ctxBalthi, &accountsv1.RemoveGroupMemberRequest{GroupId: createGroupRes.Group.Id, AccountId: createAccountMaxRes.Account.Id})
	s.Require().Error(err)

	listGroupMemberResp, err := s.groupSrv.ListGroupMembers(ctx, &accountsv1.ListGroupMembersRequest{GroupId: createGroupRes.Group.Id, Limit: 10, Offset: 0})
	s.Require().NoError(err)

	isAdmin := false
	s.Require().Equal(3, len(listGroupMemberResp.Members))
	for _, gm := range listGroupMemberResp.Members {
		if gm.Role == auth.RoleAdmin {
			isAdmin = true
		}
	}
	s.Require().True(isAdmin)
}

func (s *GroupsAPISuite) TestLeaveGroupAsUser() {
	createAccountMaxRes, err := s.accountSrv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "maxime@gmail.com", Password: "1234", Name: "Maxime"})
	s.Require().NoError(err)

	createAccountGabiRes, err := s.accountSrv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "gabriel@gmail.com", Password: "1234", Name: "Gabriel"})
	s.Require().NoError(err)

	createAccountBalthiRes, err := s.accountSrv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "balthazard@gmail.com", Password: "1234", Name: "Balthazard"})
	s.Require().NoError(err)

	uid := uuid.MustParse(createAccountMaxRes.Account.Id)

	ctx, err := s.groupSrv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	createGroupRes, err := s.groupSrv.CreateGroup(ctx, &accountsv1.CreateGroupRequest{Description: "description", Name: "EIP"})
	s.Require().NoError(err)

	_, err = s.groupSrv.AddGroupMember(ctx, &accountsv1.AddGroupMemberRequest{GroupId: createGroupRes.Group.Id, AccountId: createAccountBalthiRes.Account.Id})
	s.Require().NoError(err)

	_, err = s.groupSrv.AddGroupMember(ctx, &accountsv1.AddGroupMemberRequest{GroupId: createGroupRes.Group.Id, AccountId: createAccountGabiRes.Account.Id})
	s.Require().NoError(err)

	uidBalthi := uuid.MustParse(createAccountBalthiRes.Account.Id)

	ctxBalthi, err := s.groupSrv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uidBalthi})
	s.Require().NoError(err)

	_, err = s.groupSrv.RemoveGroupMember(ctxBalthi, &accountsv1.RemoveGroupMemberRequest{GroupId: createGroupRes.Group.Id, AccountId: createAccountBalthiRes.Account.Id})
	s.Require().NoError(err)

	listGroupMemberResp, err := s.groupSrv.ListGroupMembers(ctx, &accountsv1.ListGroupMembersRequest{GroupId: createGroupRes.Group.Id, Limit: 10, Offset: 0})
	s.Require().NoError(err)

	isAdmin := false
	s.Require().Equal(2, len(listGroupMemberResp.Members))
	for _, gm := range listGroupMemberResp.Members {
		if gm.Role == auth.RoleAdmin {
			isAdmin = true
		}
	}
	s.Require().True(isAdmin)
}

func (s *GroupsAPISuite) TestListGroupIBelongTo() {
	createAccountMaxRes, err := s.accountSrv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "maxime@gmail.com", Password: "1234", Name: "Maxime"})
	s.Require().NoError(err)

	createAccountGabiRes, err := s.accountSrv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "gabriel@gmail.com", Password: "1234", Name: "Gabriel"})
	s.Require().NoError(err)

	createAccountBalthiRes, err := s.accountSrv.CreateAccount(context.TODO(), &accountsv1.CreateAccountRequest{Email: "balthazard@gmail.com", Password: "1234", Name: "Balthazard"})
	s.Require().NoError(err)

	uid := uuid.MustParse(createAccountMaxRes.Account.Id)

	uidGabi := uuid.MustParse(createAccountGabiRes.Account.Id)

	ctxGabi, err := s.groupSrv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uidGabi})
	s.Require().NoError(err)

	ctx, err := s.groupSrv.auth.ContextWithToken(context.TODO(), &auth.Token{UserID: uid})
	s.Require().NoError(err)

	createGroupRes1, err := s.groupSrv.CreateGroup(ctx, &accountsv1.CreateGroupRequest{Description: "description", Name: "Group1"})
	s.Require().NoError(err)

	_, err = s.groupSrv.AddGroupMember(ctx, &accountsv1.AddGroupMemberRequest{GroupId: createGroupRes1.Group.Id, AccountId: createAccountGabiRes.Account.Id})
	s.Require().NoError(err)

	_, err = s.groupSrv.AddGroupMember(ctx, &accountsv1.AddGroupMemberRequest{GroupId: createGroupRes1.Group.Id, AccountId: createAccountBalthiRes.Account.Id})
	s.Require().NoError(err)

	s.Require().NoError(err)

	createGroupRes2, err := s.groupSrv.CreateGroup(ctx, &accountsv1.CreateGroupRequest{Description: "description", Name: "Group2"})
	s.Require().NoError(err)

	_, err = s.groupSrv.AddGroupMember(ctx, &accountsv1.AddGroupMemberRequest{GroupId: createGroupRes2.Group.Id, AccountId: createAccountGabiRes.Account.Id})
	s.Require().NoError(err)

	_, err = s.groupSrv.AddGroupMember(ctx, &accountsv1.AddGroupMemberRequest{GroupId: createGroupRes2.Group.Id, AccountId: createAccountBalthiRes.Account.Id})
	s.Require().NoError(err)

	_, err = s.groupSrv.CreateGroup(ctxGabi, &accountsv1.CreateGroupRequest{Description: "description", Name: "Group3"})
	s.Require().NoError(err)

	_, err = s.groupSrv.CreateGroup(ctxGabi, &accountsv1.CreateGroupRequest{Description: "description", Name: "Group4"})
	s.Require().NoError(err)

	_, err = s.groupSrv.CreateGroup(ctx, &accountsv1.CreateGroupRequest{Description: "description", Name: "Group5"})
	s.Require().NoError(err)

	listGroupRespGabi, err := s.groupSrv.ListGroups(ctx, &accountsv1.ListGroupsRequest{AccountId: createAccountGabiRes.Account.Id, Limit: 10, Offset: 0})
	s.Require().NoError(err)
	s.Require().Equal(4, len(listGroupRespGabi.Groups))

	listGroupRespBalthi, err := s.groupSrv.ListGroups(ctx, &accountsv1.ListGroupsRequest{AccountId: createAccountBalthiRes.Account.Id})
	s.Require().NoError(err)
	s.Require().Equal(2, len(listGroupRespBalthi.Groups))
}
