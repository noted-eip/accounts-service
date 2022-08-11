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

type groupsService struct {
	accountsv1.UnimplementedGroupsAPIServer

	auth   auth.Service
	logger *zap.Logger
	repo   models.GroupsRepository
}

var _ accountsv1.GroupsAPIServer = &groupsService{}

func (srv *groupsService) CreateGroup(ctx context.Context, in *accountsv1.CreateGroupRequest) (*accountsv1.CreateGroupResponse, error) {
	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	group, err := srv.repo.Create(ctx, &models.GroupPayload{Name: &in.Name, Members: &[]models.GroupMember{{ID: token.UserID.String()}}, Description: &in.Description})
	if err != nil {
		// TODO: Translate error from models.Err to gRPC.
		return nil, status.Error(codes.Internal, "could not create group")
	}

	return &accountsv1.CreateGroupResponse{
		Group: &accountsv1.Group{
			Id:          group.ID,
			Name:        *group.Name,
			Description: *group.Description,
			OwnerId:     *group.OwnerID,
		},
	}, nil
}

func (srv *groupsService) DeleteGroup(ctx context.Context, in *accountsv1.DeleteGroupRequest) (*accountsv1.DeleteGroupResponse, error) {
	// TODO: validation.
	// TODO: Cannot a delete a group which I do not own.

	err := srv.repo.Delete(ctx, &models.OneGroupFilter{ID: in.Id})
	if err != nil {
		return nil, status.Error(codes.Internal, "unable to delete group")
	}

	return &accountsv1.DeleteGroupResponse{}, nil
}

func (srv *groupsService) UpdateGroup(ctx context.Context, in *accountsv1.UpdateGroupRequest) (*accountsv1.UpdateGroupResponse, error) {
	// 	id, err := uuid.Parse(in.Group.Id)
	// 	if err != nil {
	// 		srv.logger.Error("failed to convert uuid from string", zap.Error(err))
	// 		return nil, status.Error(codes.Internal, "could not update account")
	// 	}

	// 	fieldMask := in.GetUpdateMask()
	// 	fieldMask.Normalize()
	// 	if !fieldMask.IsValid(in.Group) {
	// 		return nil, status.Error(codes.InvalidArgument, "invalid field mask")
	// 	}
	// 	fmutils.Filter(in.GetGroup
	// (), fieldMask.GetPaths())

	// 	acc, err := srv.repo.Get(ctx, &models.OneGroupFilter{ID: id})
	// 	if err != nil {
	// 		srv.logger.Error("failed to get Group", zap.Error(err))
	// 		return nil, status.Error(codes.Internal, "could not update Group")
	// 	}

	// 	var protoGroup accountsv1.Account
	// 	err = copier.Copy(&protoGroup, &acc)
	// 	if err != nil {
	// 		srv.logger.Error("invalid account conversion", zap.Error(err))
	// 		return nil, status.Error(codes.Internal, "could not update account")
	// 	}
	// 	proto.Merge(&protoGroup, in.Group)

	// 	err = srv.repo.Update(ctx, &models.OneGroupFilter{ID: id}, &models.GroupPayload{Name: &protoAccount.Email, Name: &protoAccount.Name})
	// 	if err != nil {
	// 		srv.logger.Error("failed to update account", zap.Error(err))
	// 		return nil, status.Error(codes.Internal, "could not update account")
	// 	}
	// 	protoAccount.Id = id.String()

	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (srv *groupsService) JoinGroup(ctx context.Context, in *accountsv1.JoinGroupRequest) (*accountsv1.JoinGroupResponse, error) {
	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	acc, err := srv.repo.Get(ctx, &models.OneGroupFilter{ID: in.Id})
	if err != nil {
		srv.logger.Error("failed to get group", zap.Error(err))
		return nil, status.Error(codes.Internal, "could not join group")
	}

	newMember := *acc.Members
	newMember = append(newMember, models.GroupMember{ID: token.UserID.String()})

	_, err = srv.repo.Update(ctx, &models.OneGroupFilter{ID: in.Id}, &models.GroupPayload{Members: &newMember})
	if err != nil {
		srv.logger.Error("failed to update group", zap.Error(err))
		return nil, status.Error(codes.Internal, "could not join group")
	}
	return &accountsv1.JoinGroupResponse{}, nil
}

func (srv *groupsService) AddNoteToGroup(ctx context.Context, in *accountsv1.AddNoteToGroupRequest) (*accountsv1.AddNoteToGroupResponse, error) {
	// id, err := uuid.Parse(in.Id)
	// if err != nil {
	// 	srv.logger.Error("failed to convert uuid from string", zap.Error(err))
	// 	return nil, status.Error(codes.Internal, "could not join group")
	// }

	// noteId, err := uuid.Parse(in.NoteId)
	// if err != nil {
	// 	srv.logger.Error("failed to convert uuid from string", zap.Error(err))
	// 	return nil, status.Error(codes.Internal, "could not join group")
	// }

	// acc, err := srv.repo.Get(ctx, &models.OneGroupFilter{ID: id.String()})
	// if err != nil {
	// 	srv.logger.Error("failed to get group", zap.Error(err))
	// 	return nil, status.Error(codes.Internal, "could not join group")
	// }

	// newNote := *acc.Notes
	// newNote = append(newNote, models.Note{ID: noteId.String()})

	// err = srv.repo.Update(ctx, &models.OneGroupFilter{ID: id.String()}, &models.GroupPayload{Notes: &newNote})
	// if err != nil {
	// 	srv.logger.Error("failed to update group", zap.Error(err))
	// 	return nil, status.Error(codes.Internal, "could not join group")
	// }

	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// TODO: This function is duplicated from accountsService.authenticate().
// Find a way to extract this into a separate function or use a base class
// to share common behaviour.
func (srv *groupsService) authenticate(ctx context.Context) (*auth.Token, error) {
	token, err := srv.auth.TokenFromContext(ctx)
	if err != nil {
		srv.logger.Debug("could not authenticate request", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	return token, nil
}
