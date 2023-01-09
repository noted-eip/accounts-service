package main

import (
	"accounts-service/models"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"accounts-service/validators"
	"context"
	"errors"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TODO: internal token clarification with invitation API

func (srv *groupsAPI) AddGroupMember(ctx context.Context, in *accountsv1.AddGroupMemberRequest) (*accountsv1.AddGroupMemberResponse, error) {
	err := validators.ValidateAddGroupMember(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, err = srv.groupRepo.Get(ctx, &models.OneGroupFilter{ID: in.GroupId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	payload := models.MemberPayload{AccountID: &in.AccountId, GroupID: &in.GroupId, Role: "user"}
	_, err = srv.memberRepo.Create(ctx, &payload)
	if err != nil {
		return nil, statusFromModelError(err)
	}

	return &accountsv1.AddGroupMemberResponse{}, nil
}

func (srv *groupsAPI) RemoveGroupMember(ctx context.Context, in *accountsv1.RemoveGroupMemberRequest) (*accountsv1.RemoveGroupMemberResponse, error) {
	err := validators.ValidateRemoveGroupMember(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	filter := models.MemberFilter{AccountID: &in.AccountId, GroupID: &in.GroupId}
	member, err := srv.memberRepo.Get(ctx, &filter)
	if err != nil {
		return nil, statusFromModelError(err)
	}

	id := token.UserID.String()
	memberRequestDeletionFilter := models.MemberFilter{AccountID: &id, GroupID: &in.GroupId}
	memberRequestDeletion, err := srv.memberRepo.Get(ctx, &memberRequestDeletionFilter)
	if err != nil {
		return nil, statusFromModelError(err)
	}

	if memberRequestDeletion.Role == "user" && *member.AccountID != token.UserID.String() {
		return nil, status.Error(codes.PermissionDenied, "user must be admin or delete himself")
	}

	memberDel, err := srv.memberRepo.DeleteOne(ctx, &filter)
	if err != nil {
		return nil, statusFromModelError(err)
	}

	if memberDel.Role == "admin" {
		_, err = srv.memberRepo.Update(ctx, &models.MemberFilter{GroupID: memberDel.GroupID}, &models.MemberPayload{GroupID: member.GroupID, Role: "admin"})
		if err != nil {
			return nil, statusFromModelError(err)
		}
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

	filter := models.MemberFilter{AccountID: &in.AccountId, GroupID: &in.GroupId}
	member, err := srv.memberRepo.Get(ctx, &filter)
	if err != nil {
		return nil, statusFromModelError(err)
	}

	if member == nil {
		return nil, status.Error(codes.NotFound, "member not found")
	}

	groupMember := accountsv1.GroupMember{AccountId: *member.AccountID, Role: member.Role, CreatedAt: timestamppb.New(member.CreatedAt)}
	return &accountsv1.GetGroupMemberResponse{Member: &groupMember}, nil
}

func (srv *groupsAPI) ListGroupMembers(ctx context.Context, in *accountsv1.ListGroupMembersRequest) (*accountsv1.ListGroupMembersResponse, error) {

	err := validators.ValidateListGroupMember(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	_, err = srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	if in.Limit == 0 {
		in.Limit = 10
	}

	filter := models.MemberFilter{GroupID: &in.GroupId}
	members, err := srv.memberRepo.List(ctx, &filter)
	if err != nil {
		return nil, statusFromModelError(err)
	}

	var groupMembers []*accountsv1.GroupMember
	for i := int(in.Offset); i < len(members); i++ {
		if i == int(in.Limit) {
			break
		}
		groupMember := &accountsv1.GroupMember{AccountId: *members[i].AccountID, Role: members[i].Role, CreatedAt: timestamppb.New(members[i].CreatedAt)}
		if err != nil {
			srv.logger.Error("failed to decode member", zap.Error(err))
		}
		groupMembers = append(groupMembers, groupMember)
	}

	return &accountsv1.ListGroupMembersResponse{Members: groupMembers}, nil
}
