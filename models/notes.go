package models

import (
	"context"
	"time"
)

type Note struct {
	ID        string    `json:"_id" bson:"_id,omitempty"`
	NoteID    string    `json:"note_id" bson:"note_id,omitempty"`
	GroupID   string    `json:"group_id" bson:"group_id,omitempty"`
	Title     string    `json:"title" bson:"title,omitempty"`
	AuthorID  string    `json:"author_id" bson:"author_id,omitempty"`
	CreatedAt time.Time `json:"created_at" bson:"created_at,omitempty"`
}

type NotePayload struct {
	NoteID   string `json:"note_id" bson:"note_id,omitempty"`
	GroupID  string `json:"group_id" bson:"group_id,omitempty"`
	Title    string `json:"title" bson:"title,omitempty"`
	AuthorID string `json:"author_id" bson:"author_id,omitempty"`
}

type NoteFilter struct {
	GroupID  string `json:"group_id" bson:"group_id,omitempty"`
	AuthorID string `json:"author_id" bson:"author_id,omitempty"`
	NoteID   string `json:"note_id" bson:"note_id,omitempty"`
}

type NotesRepository interface {
	Create(ctx context.Context, note *NotePayload) (*Note, error)

	DeleteOne(ctx context.Context, filter *NoteFilter) (*Note, error)

	DeleteMany(ctx context.Context, filter *NoteFilter) error

	Get(ctx context.Context, filter *NoteFilter) (*Note, error)

	Update(ctx context.Context, filter *NoteFilter, note *NotePayload) (*Note, error)

	List(ctx context.Context, filter *NoteFilter) ([]Note, error)
}
