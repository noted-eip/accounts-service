package main

import (
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"context"
	"errors"
)

func (srv *groupsAPI) AddGroupMember(ctx context.Context, in *accountsv1.AddGroupMemberRequest) (*accountsv1.AddGroupMemberResponse, error) {
	return nil, errors.New("not implemented")
}

func (srv *groupsAPI) RemoveGroupMember(ctx context.Context, in *accountsv1.RemoveGroupMemberRequest) (*accountsv1.RemoveGroupMemberResponse, error) {
	return nil, errors.New("not implemented")
}

func (srv *groupsAPI) UpdateGroupMember(ctx context.Context, in *accountsv1.UpdateGroupMemberRequest) (*accountsv1.UpdateGroupMemberResponse, error) {
	return nil, errors.New("not implemented")
}

func (srv *groupsAPI) GetGroupMember(ctx context.Context, in *accountsv1.GetGroupMemberRequest) (*accountsv1.GetGroupMemberResponse, error) {
	return nil, errors.New("not implemented")
}

func (srv *groupsAPI) ListGroupMember(ctx context.Context, in *accountsv1.ListGroupMembersRequest) (*accountsv1.ListGroupMembersResponse, error) {
	return nil, errors.New("not implemented")
}
