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
)

func (srv *groupsAPI) AddGroupNote(ctx context.Context, in *accountsv1.AddGroupNoteRequest) (*accountsv1.AddGroupNoteResponse, error) {
	err := validators.ValidateAddGroupNote(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	accountId := token.UserID.String()

	_, err = srv.memberRepo.Get(ctx, &models.MemberFilter{AccountID: &accountId, GroupID: &in.GroupId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	payload := models.GroupNotePayload{AuthorID: accountId, GroupID: in.GroupId, NoteID: in.NoteId, Title: in.Title}
	_, err = srv.noteRepo.Create(ctx, &payload)
	if err != nil {
		return nil, statusFromModelError(err)
	}

	return &accountsv1.AddGroupNoteResponse{Note: &accountsv1.GroupNote{AuthorAccountId: accountId, NoteId: in.NoteId, Title: in.Title}}, nil
}

func (srv *groupsAPI) RemoveGroupNote(ctx context.Context, in *accountsv1.RemoveGroupNoteRequest) (*accountsv1.RemoveGroupNoteResponse, error) {
	err := validators.ValidateRemoveGroupNote(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	accountId := token.UserID.String()

	_, err = srv.memberRepo.Get(ctx, &models.MemberFilter{AccountID: &accountId, GroupID: &in.GroupId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	filter := models.GroupNoteFilter{NoteID: in.NoteId, GroupID: in.GroupId}
	_, err = srv.noteRepo.DeleteOne(ctx, &filter)
	if err != nil {
		return nil, statusFromModelError(err)
	}

	return &accountsv1.RemoveGroupNoteResponse{}, nil
}

func (srv *groupsAPI) UpdateGroupNote(ctx context.Context, in *accountsv1.UpdateGroupNoteRequest) (*accountsv1.UpdateGroupNoteResponse, error) {

	return nil, errors.New("not implemented")
}

func (srv *groupsAPI) GetGroupNote(ctx context.Context, in *accountsv1.GetGroupNoteRequest) (*accountsv1.GetGroupNoteResponse, error) {
	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	accountId := token.UserID.String()

	err = validators.ValidateGetGroupNote(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, err = srv.memberRepo.Get(ctx, &models.MemberFilter{AccountID: &accountId, GroupID: &in.GroupId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	note, err := srv.noteRepo.Get(ctx, &models.GroupNoteFilter{NoteID: in.NoteId, GroupID: in.GroupId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	groupNote := accountsv1.GroupNote{AuthorAccountId: note.AuthorID, NoteId: note.NoteID, Title: note.Title}
	return &accountsv1.GetGroupNoteResponse{Note: &groupNote}, nil
}

func (srv *groupsAPI) ListGroupNotes(ctx context.Context, in *accountsv1.ListGroupNotesRequest) (*accountsv1.ListGroupNotesResponse, error) {
	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	err = validators.ValidateListGroupNote(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	accountId := token.UserID.String()

	_, err = srv.memberRepo.Get(ctx, &models.MemberFilter{AccountID: &accountId, GroupID: &in.GroupId})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	if in.Limit == 0 {
		in.Limit = 20
	}

	groupNotes, err := srv.noteRepo.List(ctx, &models.GroupNoteFilter{GroupID: in.GroupId, AuthorID: in.AuthorAccountId}, &models.Pagination{Offset: int64(in.Offset), Limit: int64(in.Limit)})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	groupsNotesResp := []*accountsv1.GroupNote{}
	for _, groupNote := range groupNotes {
		elem := &accountsv1.GroupNote{NoteId: groupNote.NoteID, Title: groupNote.Title, AuthorAccountId: groupNote.AuthorID}
		if err != nil {
			srv.logger.Error("failed to decode groupNote", zap.Error(err))
		}
		groupsNotesResp = append(groupsNotesResp, elem)
	}
	return &accountsv1.ListGroupNotesResponse{Notes: groupsNotesResp}, nil
}
