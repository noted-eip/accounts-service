package main

import (
	"accounts-service/models"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"accounts-service/validators"
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (srv *groupsAPI) AddGroupNote(ctx context.Context, in *accountsv1.AddGroupNoteRequest) (*accountsv1.AddGroupNoteResponse, error) {
	err := validators.ValidateAddGroupNote(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to validate add groupNote request")
	}

	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	accountId := token.UserID.String()

	_, err = srv.memberRepo.Get(ctx, &models.MemberFilter{AccountID: &accountId, GroupID: &in.GroupId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to get groupNote from group_id")
	}

	payload := models.GroupNotePayload{AuthorID: accountId, GroupID: in.GroupId, NoteID: in.NoteId, Title: in.Title}
	_, err = srv.noteRepo.Create(ctx, &payload)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to add note to group")
	}

	return &accountsv1.AddGroupNoteResponse{Note: &accountsv1.GroupNote{AuthorAccountId: accountId, NoteId: in.NoteId, Title: in.Title}}, nil
}

func (srv *groupsAPI) RemoveGroupNote(ctx context.Context, in *accountsv1.RemoveGroupNoteRequest) (*accountsv1.RemoveGroupNoteResponse, error) {
	err := validators.ValidateRemoveGroupNote(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to validate add member request")
	}

	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	accountId := token.UserID.String()

	_, err = srv.memberRepo.Get(ctx, &models.MemberFilter{AccountID: &accountId, GroupID: &in.GroupId})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to get member from group_id")
	}

	filter := models.GroupNoteFilter{NoteID: in.NoteId, GroupID: in.GroupId}
	_, err = srv.noteRepo.DeleteOne(ctx, &filter)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete note from group")
	}

	return &accountsv1.RemoveGroupNoteResponse{}, nil
}

func (srv *groupsAPI) UpdateGroupNote(ctx context.Context, in *accountsv1.UpdateGroupNoteRequest) (*accountsv1.UpdateGroupNoteResponse, error) {

	return nil, errors.New("not implemented")
}

func (srv *groupsAPI) GetGroupNote(ctx context.Context, in *accountsv1.GetGroupNoteRequest) (*accountsv1.GetGroupNoteResponse, error) {
	return nil, errors.New("not implemented")
}

func (srv *groupsAPI) ListGroupNotes(ctx context.Context, in *accountsv1.ListGroupNotesRequest) (*accountsv1.ListGroupNotesResponse, error) {
	return nil, errors.New("not implemented")
}
