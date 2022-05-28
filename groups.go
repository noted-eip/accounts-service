package main

import (
	"accounts-service/grpc/groupspb"
	"accounts-service/models"
	"context"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

type groupsService struct {
	groupspb.UnimplementedGroupServiceServer

	logger *zap.SugaredLogger
	repo   models.AccountsRepository
}

var _ groupspb.GroupServiceServer = &groupsService{}

func (srv *groupsService) CreateGroup(ctx context.Context, in *groupspb.CreateGroupRequest) (*emptypb.Empty, error) {

	return &emptypb.Empty{}, nil
}

func (srv *groupsService) DeleteGroup(ctx context.Context, in *groupspb.GroupFilterRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (srv *groupsService) UpdateGroup(ctx context.Context, in *groupspb.UpdateGroupRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (srv *groupsService) ListGroupMembers(ctx context.Context, in *groupspb.GroupFilterRequest) (groupspb.ListGroupMembersReply, error) {
	return groupspb.ListGroupMembersReply{}, nil
}
