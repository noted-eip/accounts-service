package mongo

import (
	"accounts-service/models"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
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

	// _, err := rep.coll.Indexes().CreateOne(
	// 	context.Background(),
	// 	mongo.IndexModel{
	// 		Keys:    bson.D{{Key: "account_id", Value: 1}, {Key: "group_id", Value: 1}},
	// 		Options: options.Index().SetUnique(true),
	// 	},
	// )
	// if err != nil {
	// 	rep.logger.Error("index creation failed", zap.Error(err))
	// }
	return rep
}

func (srv *notesRepository) Create(ctx context.Context, payload *models.NotePayload) (*models.Note, error) {
	return nil, errors.New("not implemented")
}

func (srv *notesRepository) DeleteOne(ctx context.Context, filter *models.NoteFilter) (*models.Note, error) {
	return nil, errors.New("not implemented")
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
