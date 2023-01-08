package models

import (
	"context"
	"time"
)

type GroupNote struct {
	ID        string    `json:"_id" bson:"_id,omitempty"`
	NoteID    string    `json:"note_id" bson:"note_id,omitempty"`
	GroupID   string    `json:"group_id" bson:"group_id,omitempty"`
	Title     string    `json:"title" bson:"title,omitempty"`
	AuthorID  string    `json:"author_id" bson:"author_id,omitempty"`
	CreatedAt time.Time `json:"created_at" bson:"created_at,omitempty"`
}

type GroupNotePayload struct {
	NoteID   string `json:"note_id" bson:"note_id,omitempty"`
	GroupID  string `json:"group_id" bson:"group_id,omitempty"`
	Title    string `json:"title" bson:"title,omitempty"`
	AuthorID string `json:"author_id" bson:"author_id,omitempty"`
}

type GroupNoteFilter struct {
	GroupID  string `json:"group_id" bson:"group_id,omitempty"`
	AuthorID string `json:"author_id" bson:"author_id,omitempty"`
	NoteID   string `json:"note_id" bson:"note_id,omitempty"`
}

type GroupNotesRepository interface {
	Create(ctx context.Context, note *GroupNotePayload) (*GroupNote, error)

	DeleteOne(ctx context.Context, filter *GroupNoteFilter) (*GroupNote, error)

	DeleteMany(ctx context.Context, filter *GroupNoteFilter) error

	Get(ctx context.Context, filter *GroupNoteFilter) (*GroupNote, error)

	Update(ctx context.Context, filter *GroupNoteFilter, GroupNote *GroupNotePayload) (*GroupNote, error)

	List(ctx context.Context, filter *GroupNoteFilter, pagination *Pagination) ([]GroupNote, error)
}
