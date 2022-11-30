package main

import (
	"accounts-service/auth"
	"accounts-service/models"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type invitesAPI struct {
	accountsv1.UnimplementedInvitesAPIServer

	auth       auth.Service
	logger     *zap.Logger
	groupRepo  models.GroupsRepository
	inviteRepo models.InvitesRepository
}

var _ accountsv1.InvitesAPIServer = &invitesAPI{}

func (srv *invitesAPI) SendInvite(ctx context.Context, in *accountsv1.SendInviteRequest) (*accountsv1.SendInviteResponse, error) {
	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	accountId := token.UserID.String()

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
