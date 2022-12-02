package mongo

import (
	"accounts-service/models"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type notesRepository struct {
	logger *zap.Logger
	db     *mongo.Database
	coll   *mongo.Collection
}

func NewNotesRepository(db *mongo.Database, logger *zap.Logger) models.NotesRepository {
	rep := &notesRepository{
		logger: logger.Named("mongo").Named("notes"),
		db:     db,
		coll:   db.Collection("notes"),
	}

	_, err := rep.coll.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "note_id", Value: 1}, {Key: "group_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		rep.logger.Error("index creation failed", zap.Error(err))
	}
	return rep
}

func (srv *notesRepository) Create(ctx context.Context, payload *models.NotePayload) (*models.Note, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		srv.logger.Error("failed to generate new random uuid", zap.Error(err))
		return nil, err
	}

	note := models.Note{ID: id.String(), AuthorID: payload.AuthorID, GroupID: payload.GroupID, NoteID: payload.NoteID, Title: payload.Title, CreatedAt: time.Now().UTC()}

	_, err = srv.coll.InsertOne(ctx, note)
	if err != nil {
		srv.logger.Error("insert failed", zap.Error(err), zap.String("_id", note.ID))
		return nil, err
	}

	return &note, nil
}

func (srv *notesRepository) DeleteOne(ctx context.Context, filter *models.NoteFilter) (*models.Note, error) {
	note := models.Note{}
	err := srv.coll.FindOneAndDelete(ctx, filter).Decode(&note)
	if err != nil {
		srv.logger.Error("delete one failed", zap.Error(err))
		return nil, err
	}

	if note.ID == "" {
		srv.logger.Error("delete one no document found", zap.Error(err))
	}

	return &note, nil
}

func (srv *notesRepository) DeleteMany(ctx context.Context, filter *models.NoteFilter) error {
	return errors.New("not implemented")
}

func (srv *notesRepository) Get(ctx context.Context, filter *models.NoteFilter) (*models.Note, error) {
	return nil, errors.New("not implemented")
}

func (srv *notesRepository) Update(ctx context.Context, filter *models.NoteFilter, Note *models.NotePayload) (*models.Note, error) {
	return nil, errors.New("not implemented")
}

func (srv *notesRepository) List(ctx context.Context, filter *models.NoteFilter) ([]models.Note, error) {
	return nil, errors.New("not implemented")
}
