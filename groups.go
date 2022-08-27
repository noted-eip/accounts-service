package main

import (
	"accounts-service/auth"
	"accounts-service/models"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"accounts-service/validators"
	"context"

	"github.com/jinzhu/copier"
	"github.com/mennanov/fmutils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type groupsAPI struct {
	accountsv1.UnimplementedGroupsAPIServer

	auth   auth.Service
	logger *zap.Logger
	repo   models.GroupsRepository
}

var _ accountsv1.GroupsAPIServer = &groupsAPI{}

func (srv *groupsAPI) CreateGroup(ctx context.Context, in *accountsv1.CreateGroupRequest) (*accountsv1.CreateGroupResponse, error) {
	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	id := token.UserID.String()

	group, err := srv.repo.Create(ctx, &models.GroupPayload{Name: &in.Name, OwnerID: &id, Description: &in.Description, Members: &[]models.GroupMember{{ID: token.UserID.String()}}})
	if err != nil {
		return nil, statusFromModelError(err)
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

func (srv *groupsAPI) DeleteGroup(ctx context.Context, in *accountsv1.DeleteGroupRequest) (*accountsv1.DeleteGroupResponse, error) {
	err := validators.ValidateDeleteGroupRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}
	id := token.UserID.String()

	err = srv.repo.Delete(ctx, &models.OneGroupFilter{ID: in.Id, OwnerID: &id})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	return &accountsv1.DeleteGroupResponse{}, nil
}

func (srv *groupsAPI) UpdateGroup(ctx context.Context, in *accountsv1.UpdateGroupRequest) (*accountsv1.UpdateGroupResponse, error) {

	err := validators.ValidateUpdatedGroupRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not update Group")
	}

	fieldMask := in.GetUpdateMask()
	fieldMask.Normalize()
	if !fieldMask.IsValid(in.Group) {
		return nil, status.Error(codes.InvalidArgument, "invalid field mask")
	}
	fmutils.Filter(in.GetGroup(), fieldMask.GetPaths())

	acc, err := srv.repo.Get(ctx, &models.OneGroupFilter{ID: in.Group.Id})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	var protoGroup accountsv1.Group
	err = copier.Copy(&protoGroup, &acc)
	if err != nil {
		srv.logger.Error("invalid account conversion", zap.Error(err))
		return nil, status.Error(codes.Internal, "could not update account")
	}
	proto.Merge(&protoGroup, in.Group)

	updatedGroup, err := srv.repo.Update(ctx, &models.OneGroupFilter{ID: in.Group.Id}, &models.GroupPayload{OwnerID: &protoGroup.OwnerId, Name: &protoGroup.Name, Description: &protoGroup.Description})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	var groupMembers []*accountsv1.GroupMember
	for _, members := range *updatedGroup.Members {
		elem := &accountsv1.GroupMember{AccountId: members.ID}
		if err != nil {
			srv.logger.Error("failed to decode account", zap.Error(err))
		}
		groupMembers = append(groupMembers, elem)
	}
	returnedGroup := accountsv1.Group{Id: updatedGroup.ID, OwnerId: *updatedGroup.OwnerID, Name: *updatedGroup.Name, Description: *updatedGroup.Description, Members: groupMembers}
	return &accountsv1.UpdateGroupResponse{Group: &returnedGroup}, nil
}

func (srv *groupsAPI) JoinGroup(ctx context.Context, in *accountsv1.JoinGroupRequest) (*accountsv1.JoinGroupResponse, error) {
	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	acc, err := srv.repo.Get(ctx, &models.OneGroupFilter{ID: in.Id})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	newMember := *acc.Members
	newMember = append(newMember, models.GroupMember{ID: token.UserID.String()})

	_, err = srv.repo.Update(ctx, &models.OneGroupFilter{ID: in.Id}, &models.GroupPayload{Members: &newMember})
	if err != nil {
		return nil, statusFromModelError(err)
	}
	return &accountsv1.JoinGroupResponse{}, nil
}

func (srv *groupsAPI) AddNoteToGroup(ctx context.Context, in *accountsv1.AddNoteToGroupRequest) (*accountsv1.AddNoteToGroupResponse, error) {
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
func (srv *groupsAPI) authenticate(ctx context.Context) (*auth.Token, error) {
	token, err := srv.auth.TokenFromContext(ctx)
	if err != nil {
		srv.logger.Debug("could not authenticate request", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	return token, nil
}
