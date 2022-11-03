package mongo

import (
	"accounts-service/models"
	"errors"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type membersRepository struct {
	logger *zap.Logger
	db     *mongo.Database
	coll   *mongo.Collection
}

func NewMembersRepository(db *mongo.Database, logger *zap.Logger) models.MembersRepository {
	rep := &membersRepository{
		logger: logger.Named("mongo").Named("members"),
		db:     db,
		coll:   db.Collection("members"),
	}

	_, err := rep.coll.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "account_id", Value: 1}, {Key: "group_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		rep.logger.Error("index creation failed", zap.Error(err))
	}
	return rep
}

func (srv *membersRepository) Create(ctx context.Context, payload *models.MemberPayload) (*models.Member, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		srv.logger.Error("failed to generate new random uuid", zap.Error(err))
		return nil, err
	}

	member := models.Member{ID: id.String(), Account: payload.Account, Group: payload.Group, Role: payload.Role}

	_, err = srv.coll.InsertOne(ctx, member)
	if err != nil {
		srv.logger.Error("insert failed", zap.Error(err), zap.String("group_id", *member.Group))
		return nil, err
	}

	return &member, nil
}

func (srv *membersRepository) DeleteOne(ctx context.Context, filter *models.MemberFilter) error {
	delete, err := srv.coll.DeleteOne(ctx, filter)
	if err != nil {
		srv.logger.Error("delete one failed", zap.Error(err))
		return err
	}
	if delete.DeletedCount == 0 {
		return models.ErrNotFound
	}

	return nil
}

func (srv *membersRepository) DeleteMany(ctx context.Context, filter *models.MemberFilter) error {
	delete, err := srv.coll.DeleteMany(ctx, filter)
	if err != nil {
		srv.logger.Error("delete many failed", zap.Error(err))
		return err
	}
	if delete.DeletedCount == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (srv *membersRepository) Get(ctx context.Context, filter *models.MemberFilter) (*models.Member, error) {
	var member models.Member

	err := srv.coll.FindOne(ctx, filter).Decode(&member)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, models.ErrNotFound
		}
		srv.logger.Error("query failed", zap.Error(err))
		return nil, err
	}

	return &member, nil
}

func (srv *membersRepository) Update(ctx context.Context, filter *models.MemberFilter, member *models.MemberPayload) (*models.Member, error) {
	return &models.Member{}, nil
}

func (srv *membersRepository) List(ctx context.Context, filter *models.MemberFilter) ([]models.Member, error) {
	var members []models.Member

	// opt := options.FindOptions{
	// 	Limit: &pagination.Limit,
	// 	Skip:  &pagination.Offset,
	// }

	cursor, err := srv.coll.Find(ctx, &filter)
	if err != nil {
		srv.logger.Error("mongo find members query failed", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var member models.Member
		err := cursor.Decode(&member)
		if err != nil {
			srv.logger.Error("failed to decode mongo cursor result", zap.Error(err))
		}
		members = append(members, member)
	}

	return members, nil
}
