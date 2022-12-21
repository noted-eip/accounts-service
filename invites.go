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

type invitesAPI struct {
	accountsv1.UnimplementedInvitesAPIServer

	auth         auth.Service
	logger       *zap.Logger
	groupService accountsv1.GroupsAPIServer
	groupRepo    models.GroupsRepository
	accountRepo  models.AccountsRepository
	inviteRepo   models.InvitesRepository
}

var _ accountsv1.InvitesAPIServer = &invitesAPI{}

func (srv *invitesAPI) SendInvite(ctx context.Context, in *accountsv1.SendInviteRequest) (*accountsv1.SendInviteResponse, error) {
	err := validators.ValidateSendInvite(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	accountId := token.UserID.String()

	_, err = srv.groupService.GetGroupMember(ctx, &accountsv1.GetGroupMemberRequest{GroupId: in.GroupId, AccountId: accountId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "sender not in group")
	}

	_, err = srv.accountRepo.Get(ctx, &models.OneAccountFilter{ID: in.RecipientAccountId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "recipient does not exist")
	}

	_, err = srv.groupService.GetGroupMember(ctx, &accountsv1.GetGroupMemberRequest{GroupId: in.GroupId, AccountId: in.RecipientAccountId})
	if err == nil {
		return nil, status.Error(codes.InvalidArgument, "recipient already in group")
	}

	invite, err := srv.inviteRepo.Create(ctx, &models.InvitePayload{SenderAccountID: &accountId, RecipientAccountID: &in.RecipientAccountId, GroupID: &in.GroupId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	// TODO: Invite bevahior ? E-mail etc

	return &accountsv1.SendInviteResponse{
		Invite: &accountsv1.Invite{
			Id:                 invite.ID,
			GroupId:            *invite.GroupID,
			SenderAccountId:    *invite.SenderAccountID,
			RecipientAccountId: *invite.RecipientAccountID,
		},
	}, nil
}

func (srv *invitesAPI) GetInvite(ctx context.Context, in *accountsv1.GetInviteRequest) (*accountsv1.GetInviteResponse, error) {
	err := validators.ValidateGetInvite(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	invite, err := srv.inviteRepo.Get(ctx, &models.OneInviteFilter{ID: in.InviteId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	accountId := token.UserID.String()

	if invite == nil || (*invite.RecipientAccountID != accountId && *invite.SenderAccountID != accountId) {
		return nil, status.Error(codes.NotFound, "invite not found")
	}

	return &accountsv1.GetInviteResponse{Invite: &accountsv1.Invite{
		Id:                 invite.ID,
		GroupId:            *invite.GroupID,
		SenderAccountId:    *invite.SenderAccountID,
		RecipientAccountId: *invite.RecipientAccountID,
	}}, nil
}

func (srv *invitesAPI) ListInvites(ctx context.Context, in *accountsv1.ListInvitesRequest) (*accountsv1.ListInvitesResponse, error) {
	err := validators.ValidateListInvites(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, err = srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	if in.Limit == 0 {
		in.Limit = 20
	}

	invites, err := srv.inviteRepo.List(ctx, &models.ManyInvitesFilter{
		SenderAccountID:    &in.SenderAccountId,
		RecipientAccountID: &in.RecipientAccountId,
		GroupID:            &in.GroupId,
	}, &models.Pagination{Offset: int64(in.Offset), Limit: int64(in.Limit)})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	// NOTE: As every account are apparently accessible to list and get from every account as long as you are authentificated, I did the same with invites

	var inviteResp []*accountsv1.Invite
	for _, invite := range invites {
		elem := &accountsv1.Invite{
			Id:                 invite.ID,
			GroupId:            *invite.GroupID,
			SenderAccountId:    *invite.SenderAccountID,
			RecipientAccountId: in.RecipientAccountId,
		}
		inviteResp = append(inviteResp, elem)
	}
	return &accountsv1.ListInvitesResponse{Invites: inviteResp}, nil
}

func (srv *invitesAPI) AcceptInvite(ctx context.Context, in *accountsv1.AcceptInviteRequest) (*accountsv1.AcceptInviteResponse, error) {
	err := validators.ValidateAcceptInvite(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	accountId := token.UserID.String()

	res, err := srv.GetInvite(ctx, &accountsv1.GetInviteRequest{InviteId: in.InviteId})
	if err != nil {
		return nil, status.Error(codes.NotFound, "invite not found")
	}

	invite := res.Invite
	if invite == nil || (invite.RecipientAccountId != accountId) {
		return nil, status.Error(codes.NotFound, "invite not found")
	}

	_, err = srv.groupService.AddGroupMember(ctx, &accountsv1.AddGroupMemberRequest{GroupId: invite.GroupId, AccountId: invite.RecipientAccountId})
	if err != nil {
		return nil, err
	}

	err = srv.inviteRepo.Delete(ctx, &models.ManyInvitesFilter{RecipientAccountID: &invite.RecipientAccountId, GroupID: &invite.GroupId})
	if err != nil {
		return nil, err
	}

	return &accountsv1.AcceptInviteResponse{}, nil
}

func (srv *invitesAPI) DenyInvite(ctx context.Context, in *accountsv1.DenyInviteRequest) (*accountsv1.DenyInviteResponse, error) {
	err := validators.ValidateDenyInvite(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	accountId := token.UserID.String()

	res, err := srv.GetInvite(ctx, &accountsv1.GetInviteRequest{InviteId: in.InviteId})
	if err != nil {
		return nil, status.Error(codes.NotFound, "invite not found")
	}

	invite := res.Invite
	if invite == nil || (invite.RecipientAccountId != accountId) {
		return nil, status.Error(codes.NotFound, "invite not found")
	}

	err = srv.inviteRepo.Delete(ctx, &models.ManyInvitesFilter{
		RecipientAccountID: &invite.RecipientAccountId,
		GroupID:            &invite.GroupId,
		SenderAccountID:    &invite.SenderAccountId,
	})
	if err != nil {
		return nil, err
	}

	return &accountsv1.DenyInviteResponse{}, nil
}

// TODO: This function is duplicated from accountsService.authenticate().
// Find a way to extract this into a separate function or use a base class
// to share common behaviour.
func (srv *invitesAPI) authenticate(ctx context.Context) (*auth.Token, error) {
	token, err := srv.auth.TokenFromContext(ctx)
	if err != nil {
		srv.logger.Debug("could not authenticate request", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	return token, nil
}
