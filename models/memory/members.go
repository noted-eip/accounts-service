package memory

import (
	"accounts-service/auth"
	"accounts-service/models"
	"errors"

	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type membersRepository struct {
	logger    *zap.Logger
	membersDB *Database
}

func NewMembersRepository(membersDB *Database, logger *zap.Logger) models.MembersRepository {
	return &membersRepository{
		logger:    logger.Named("memory").Named("members"),
		membersDB: membersDB,
	}
}

func (srv *membersRepository) Create(ctx context.Context, payload *models.MemberPayload) (*models.Member, error) {
	txn := srv.membersDB.DB.Txn(true)
	defer txn.Abort()

	id, err := uuid.NewRandom()
	if err != nil {
		srv.logger.Error("failed to generate new random uuid", zap.Error(err))
		return nil, err
	}

	member := models.Member{
		ID:        id.String(),
		GroupID:   payload.GroupID,
		AccountID: payload.AccountID,
		Role:      payload.Role,
	}

	err = txn.Insert("member", &member)
	if err != nil {
		srv.logger.Error("insert member failed", zap.Error(err), zap.String("account_id", *member.AccountID))
		return nil, err
	}

	txn.Commit()
	return &member, nil
}

func (srv *membersRepository) DeleteOne(ctx context.Context, filter *models.MemberFilter) (*models.Member, error) {
	txn := srv.membersDB.DB.Txn(true)
	defer txn.Abort()

	it, err := txn.Get("member", "group_id", *filter.GroupID)

	if err != nil {
		if errors.Is(err, memdb.ErrNotFound) {
			return nil, models.ErrNotFound
		}
		srv.logger.Error("unable to get member", zap.Error(err))
		return nil, err
	}

	memberDel := &models.Member{}
	for obj := it.Next(); obj != nil; obj = it.Next() {
		memberDel = obj.(*models.Member)
		if *memberDel.AccountID == *filter.AccountID {
			err = txn.Delete("member", models.Member{ID: memberDel.ID})
			if err != nil {
				srv.logger.Error("unable to delete member from object", zap.Error(err))
				return nil, err
			}
			break
		}
	}

	txn.Commit()
	return memberDel, nil
}

func (srv *membersRepository) DeleteMany(ctx context.Context, filter *models.MemberFilter) error {
	txn := srv.membersDB.DB.Txn(true)
	defer txn.Abort()

	it, err := txn.Get("member", "group_id", *filter.GroupID)

	if err != nil {
		if errors.Is(err, memdb.ErrNotFound) {
			return models.ErrNotFound
		}
		srv.logger.Error("unable to get members", zap.Error(err))
		return err
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		err = txn.Delete("member", models.Member{ID: obj.(*models.Member).ID})
		if err != nil {
			srv.logger.Error("unable to delete member from object", zap.Error(err))
			return err
		}
	}
	txn.Commit()
	return nil
}

func (srv *membersRepository) Get(ctx context.Context, filter *models.MemberFilter) (*models.Member, error) {
	txn := srv.membersDB.DB.Txn(false)
	defer txn.Abort()

	it, err := txn.Get("member", "account_id", *filter.AccountID)
	if err != nil {
		if errors.Is(err, memdb.ErrNotFound) {
			return nil, err
		}
		srv.logger.Error("unable to query member", zap.Error(err))
		return nil, err
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		member := obj.(*models.Member)
		if *member.GroupID == *filter.GroupID {
			return member, nil
		}
	}
	return nil, models.ErrNotFound
}

func (srv *membersRepository) Update(ctx context.Context, filter *models.MemberFilter, member *models.MemberPayload) (*models.Member, error) {
	return nil, errors.New("not implemented")
}

func (srv *membersRepository) List(ctx context.Context, filter *models.MemberFilter) ([]models.Member, error) {
	var members []models.Member
	var err error
	var it memdb.ResultIterator

	txn := srv.membersDB.DB.Txn(false)

	if filter.GroupID == nil || *filter.GroupID == "" {
		it, err = txn.Get("member", "account_id", *filter.AccountID)
	} else {
		it, err = txn.Get("member", "group_id", *filter.GroupID)
	}
	if err != nil {
		return nil, err
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		newMember := models.Member{ID: obj.(*models.Member).ID, GroupID: obj.(*models.Member).GroupID, AccountID: obj.(*models.Member).AccountID, Role: obj.(*models.Member).Role}
		members = append(members, newMember)
	}

	return members, nil
}

func (srv *membersRepository) SetAdmin(ctx context.Context, filter *models.MemberFilter) error {
	txn := srv.membersDB.DB.Txn(true)
	defer txn.Abort()

	it, err := txn.Get("member", "group_id", *filter.GroupID)

	if err != nil {
		if errors.Is(err, memdb.ErrNotFound) {
			return models.ErrNotFound
		}
		srv.logger.Error("unable to get members", zap.Error(err))
		return err
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		newAdmin := obj.(*models.Member)
		err = txn.Insert("member", &models.Member{ID: newAdmin.ID, AccountID: newAdmin.AccountID, GroupID: newAdmin.GroupID, Role: auth.RoleAdmin})
		if err != nil {
			srv.logger.Error("unable to update member from object", zap.Error(err))
			return err
		} else {
			break
		}
	}
	txn.Commit()
	return nil
}
