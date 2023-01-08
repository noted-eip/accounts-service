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

type groupNotesRepository struct {
	logger *zap.Logger
	db     *mongo.Database
	coll   *mongo.Collection
}

func NewNotesRepository(db *mongo.Database, logger *zap.Logger) models.GroupNotesRepository {
	rep := &groupNotesRepository{
		logger: logger.Named("mongo").Named("notes"),
		db:     db,
		coll:   db.Collection("notes"),
	}

	_, err := rep.coll.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "note_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		rep.logger.Error("index creation failed", zap.Error(err))
	}
	return rep
}

func (srv *groupNotesRepository) Create(ctx context.Context, payload *models.GroupNotePayload) (*models.GroupNote, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		srv.logger.Error("failed to generate new random uuid", zap.Error(err))
		return nil, err
	}

	note := models.GroupNote{ID: id.String(), AuthorID: payload.AuthorID, GroupID: payload.GroupID, NoteID: payload.NoteID, Title: payload.Title, CreatedAt: time.Now().UTC()}

	_, err = srv.coll.InsertOne(ctx, note)
	if err != nil {
		srv.logger.Error("insert failed", zap.Error(err), zap.String("_id", note.ID))
		return nil, err
	}

	return &note, nil
}

func (srv *groupNotesRepository) DeleteOne(ctx context.Context, filter *models.GroupNoteFilter) (*models.GroupNote, error) {
	note := models.GroupNote{}
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

func (srv *groupNotesRepository) DeleteMany(ctx context.Context, filter *models.GroupNoteFilter) error {
	return errors.New("not implemented")
}

func (srv *groupNotesRepository) Get(ctx context.Context, filter *models.GroupNoteFilter) (*models.GroupNote, error) {
	var note models.GroupNote

	err := srv.coll.FindOne(ctx, filter).Decode(&note)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, models.ErrNotFound
		}
		srv.logger.Error("query failed", zap.Error(err))
		return nil, err
	}

	return &note, nil
}

func (srv *groupNotesRepository) Update(ctx context.Context, filter *models.GroupNoteFilter, GroupNote *models.GroupNotePayload) (*models.GroupNote, error) {
	return nil, errors.New("not implemented")
}

func (srv *groupNotesRepository) List(ctx context.Context, filter *models.GroupNoteFilter, pagination *models.Pagination) ([]models.GroupNote, error) {
	var groupNotes []models.GroupNote

	opt := options.FindOptions{
		Limit: &pagination.Limit,
		Skip:  &pagination.Offset,
	}

	cursor, err := srv.coll.Find(ctx, filter, &opt)
	if err != nil {
		srv.logger.Error("mongo find groupNotes query failed", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var elem models.GroupNote
		err := cursor.Decode(&elem)
		if err != nil {
			srv.logger.Error("failed to decode mongo cursor result", zap.Error(err))
		}
		groupNotes = append(groupNotes, elem)
	}

	return groupNotes, nil
}
