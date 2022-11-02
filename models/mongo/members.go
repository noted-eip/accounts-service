package mongo

import (
	"accounts-service/models"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type membersRepository struct {
	logger *zap.Logger
	db     *mongo.Database
	coll   *mongo.Collection
}

func NewMembersRepository(db *mongo.Database, logger *zap.Logger) models.MembersRepository {
	return &membersRepository{
		logger: logger.Named("mongo").Named("members"),
		db:     db,
		coll:   db.Collection("members"),
	}
}

func (srv *membersRepository) Create(ctx context.Context, member *models.MemberPayload) (*models.Member, error) {
	return &models.Member{}, nil
}

func (srv *membersRepository) Delete(ctx context.Context, filter *models.MemberFilter) error {
	return nil
}

func (srv *membersRepository) Get(ctx context.Context, filter *models.MemberFilter) (*models.Member, error) {
	return &models.Member{}, nil
}

func (srv *membersRepository) Update(ctx context.Context, filter *models.MemberFilter, member *models.MemberPayload) (*models.Member, error) {
	return &models.Member{}, nil
}

func (srv *membersRepository) List(ctx context.Context, filter *models.MemberFilter, pagination *models.Pagination) ([]models.Member, error) {
	return []models.Member{}, nil
}
