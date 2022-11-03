package main

import (
	"accounts-service/auth"
	"accounts-service/models"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"accounts-service/validators"
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TODO: internal token clarification with invitation API

func (srv *groupsAPI) AddGroupMember(ctx context.Context, in *accountsv1.AddGroupMemberRequest) (*accountsv1.AddGroupMemberResponse, error) {
	err := validators.ValidateAddGroupMember(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to validate add member request")
	}

	_, err = srv.groupRepo.Get(ctx, &models.OneGroupFilter{ID: in.GroupId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to get group from group_id")
	}

	payload := models.MemberPayload{Account: &in.AccountId, Group: &in.GroupId, Role: auth.RoleUser, CreatedAt: time.Now().UTC()}
	_, err = srv.memberRepo.Create(ctx, &payload)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to add member to group")
	}

	return &accountsv1.AddGroupMemberResponse{}, nil
}

func (srv *groupsAPI) RemoveGroupMember(ctx context.Context, in *accountsv1.RemoveGroupMemberRequest) (*accountsv1.RemoveGroupMemberResponse, error) {
	err := validators.ValidateRemoveGroupMember(in)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to validate remove member request")
	}

	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	filter := models.MemberFilter{Account: &in.AccountId, Group: &in.GroupId}
	member, err := srv.memberRepo.Get(ctx, &filter)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user from requested group")
	}

	if member.Role == auth.RoleUser && *member.Account != token.UserID.String() {
		return nil, status.Error(codes.PermissionDenied, "user must be admin or delete himself")
	}

	err = srv.memberRepo.DeleteOne(ctx, &filter)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to remove member to group")
	}

	return &accountsv1.RemoveGroupMemberResponse{}, nil
}

func (srv *groupsAPI) UpdateGroupMember(ctx context.Context, in *accountsv1.UpdateGroupMemberRequest) (*accountsv1.UpdateGroupMemberResponse, error) {

	return nil, errors.New("not implemented")
}

func (srv *groupsAPI) GetGroupMember(ctx context.Context, in *accountsv1.GetGroupMemberRequest) (*accountsv1.GetGroupMemberResponse, error) {
	err := validators.ValidateGetGroupMember(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, err = srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	filter := models.MemberFilter{Account: &in.AccountId, Group: &in.GroupId}
	member, err := srv.memberRepo.Get(ctx, &filter)
	if err != nil {
		return nil, statusFromModelError(err)
	}

	if member == nil {
		return nil, status.Error(codes.NotFound, "member not found")
	}

	groupMember := accountsv1.GroupMember{AccountId: *member.Account, Role: member.Role, CreatedAt: timestamppb.New(member.CreatedAt)}
	return &accountsv1.GetGroupMemberResponse{Member: &groupMember}, nil
}

func (srv *groupsAPI) ListGroupMembers(ctx context.Context, in *accountsv1.ListGroupMembersRequest) (*accountsv1.ListGroupMembersResponse, error) {
	err := validators.ValidateListGroupMember(in)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to remove member to group")
	}

	_, err = srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	filter := models.MemberFilter{Group: &in.GroupId}
	members, err := srv.memberRepo.List(ctx, &filter)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to remove member to group")
	}

	var groupMembers []*accountsv1.GroupMember
	for _, member := range members {
		groupMember := &accountsv1.GroupMember{AccountId: *member.Account, Role: member.Role, CreatedAt: timestamppb.New(member.CreatedAt)}
		if err != nil {
			srv.logger.Error("failed to decode member", zap.Error(err))
		}
		groupMembers = append(groupMembers, groupMember)
	}

	return &accountsv1.ListGroupMembersResponse{Members: groupMembers}, nil
}
