package main

import (
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"context"
	"errors"
)

func (srv *groupsAPI) AddGroupNote(ctx context.Context, in *accountsv1.AddGroupNoteRequest) (*accountsv1.AddGroupNoteResponse, error) {
	return nil, errors.New("not implemented")
}

func (srv *groupsAPI) RemoveGroupNote(ctx context.Context, in *accountsv1.RemoveGroupNoteRequest) (*accountsv1.RemoveGroupNoteResponse, error) {
	return nil, errors.New("not implemented")
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
