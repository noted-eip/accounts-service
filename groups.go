package main

import (
	"accounts-service/grpc/groupspb"
	"accounts-service/models"
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type groupsService struct {
	groupspb.UnimplementedGroupServiceServer

	logger *zap.SugaredLogger
	repo   models.GroupsRepository
}

var _ groupspb.GroupServiceServer = &groupsService{}

func (srv *groupsService) CreateGroup(ctx context.Context, in *groupspb.CreateGroupRequest) (*emptypb.Empty, error) {
	id, err := uuid.Parse(in.OwnerId)
	if err != nil {
		srv.logger.Errorw("failed to convert uuid from string", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not get account")
	}

	srv.repo.Create(ctx, &models.GroupPayload{Name: &in.Name, Members: &[]models.Member{{ID: id}}})
	return &emptypb.Empty{}, nil
}

func (srv *groupsService) DeleteGroup(ctx context.Context, in *groupspb.GroupFilterRequest) (*emptypb.Empty, error) {
	id, err := uuid.Parse(in.Id)
	if err != nil {
		srv.logger.Errorw("failed to convert uuid from string", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not get account")
	}
	srv.repo.Delete(ctx, &models.OneGroupFilter{ID: id, Name: in.Name})
	return &emptypb.Empty{}, nil
}

func (srv *groupsService) UpdateGroup(ctx context.Context, in *groupspb.UpdateGroupRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (srv *groupsService) ListGroupMembers(ctx context.Context, in *groupspb.GroupFilterRequest) (groupspb.ListGroupMembersReply, error) {
	return groupspb.ListGroupMembersReply{}, nil
}
