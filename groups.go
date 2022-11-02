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

	auth       auth.Service
	logger     *zap.Logger
	groupRepo  models.GroupsRepository
	memberRepo models.MembersRepository
}

var _ accountsv1.GroupsAPIServer = &groupsAPI{}

func (srv *groupsAPI) CreateGroup(ctx context.Context, in *accountsv1.CreateGroupRequest) (*accountsv1.CreateGroupResponse, error) {
	// token, err := srv.authenticate(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	// id := token.UserID.String()

	group, err := srv.groupRepo.Create(ctx, &models.GroupPayload{Name: &in.Name, Description: &in.Description})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	return &accountsv1.CreateGroupResponse{
		Group: &accountsv1.Group{
			Id:          group.ID,
			Name:        *group.Name,
			Description: *group.Description,
		},
	}, nil
}

func (srv *groupsAPI) DeleteGroup(ctx context.Context, in *accountsv1.DeleteGroupRequest) (*accountsv1.DeleteGroupResponse, error) {
	err := validators.ValidateDeleteGroupRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// token, err := srv.authenticate(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	// id := token.UserID.String()

	err = srv.groupRepo.Delete(ctx, &models.OneGroupFilter{ID: in.GroupId})
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

	acc, err := srv.groupRepo.Get(ctx, &models.OneGroupFilter{ID: in.Group.Id})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	fmutils.Filter(in.GetGroup(), fieldMask.GetPaths())

	var protoGroup accountsv1.Group
	err = copier.Copy(&protoGroup, &acc)
	if err != nil {
		srv.logger.Error("invalid group conversion", zap.Error(err))
		return nil, status.Error(codes.Internal, "could not update group")
	}
	proto.Merge(&protoGroup, in.Group)

	updatedGroup, err := srv.groupRepo.Update(ctx, &models.OneGroupFilter{ID: acc.ID}, &models.GroupPayload{Name: &protoGroup.Name, Description: &protoGroup.Description})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	returnedGroup := accountsv1.Group{Id: updatedGroup.ID, Name: *updatedGroup.Name, Description: *updatedGroup.Description}
	return &accountsv1.UpdateGroupResponse{Group: &returnedGroup}, nil
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
