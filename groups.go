package main

import (
	"accounts-service/models"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type groupsService struct {
	accountsv1.UnimplementedGroupsAPIServer

	logger *zap.SugaredLogger
	repo   models.GroupsRepository
}

var _ accountsv1.GroupsAPIServer = &groupsService{}

func (srv *groupsService) CreateGroup(ctx context.Context, in *accountsv1.CreateGroupRequest) (*accountsv1.CreateGroupResponse, error) {
	id, err := uuid.Parse(in.OwnerId)
	if err != nil {
		srv.logger.Errorw("failed to convert uuid from string", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not get account")
	}
	fmt.Println("description => ", in)
	srv.repo.Create(ctx, &models.GroupPayload{Name: &in.Name, Members: &[]models.Member{{ID: id.String()}}, Description: &in.Description})
	return &accountsv1.CreateGroupResponse{}, nil
}

func (srv *groupsService) DeleteGroup(ctx context.Context, in *accountsv1.DeleteGroupRequest) (*accountsv1.DeleteGroupResponse, error) {
	id, err := uuid.Parse(in.Id)
	if err != nil {
		srv.logger.Errorw("failed to convert uuid from string", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not get account")
	}
	srv.repo.Delete(ctx, &models.OneGroupFilter{ID: id.String(), Name: &in.Name})
	return &accountsv1.DeleteGroupResponse{}, nil
}

func (srv *groupsService) UpdateGroup(ctx context.Context, in *accountsv1.UpdateGroupRequest) (*accountsv1.UpdateGroupResponse, error) {
	// 	id, err := uuid.Parse(in.Group.Id)
	// 	if err != nil {
	// 		srv.logger.Errorw("failed to convert uuid from string", "error", err.Error())
	// 		return nil, status.Errorf(codes.Internal, "could not update account")
	// 	}

	// 	fieldMask := in.GetUpdateMask()
	// 	fieldMask.Normalize()
	// 	if !fieldMask.IsValid(in.Group) {
	// 		return nil, status.Errorf(codes.InvalidArgument, "invalid field mask")
	// 	}
	// 	fmutils.Filter(in.GetGroup
	// (), fieldMask.GetPaths())

	// 	acc, err := srv.repo.Get(ctx, &models.OneGroupFilter{ID: id})
	// 	if err != nil {
	// 		srv.logger.Errorw("failed to get Group", "error", err.Error())
	// 		return nil, status.Errorf(codes.Internal, "could not update Group")
	// 	}

	// 	var protoGroup accountsv1.Account
	// 	err = copier.Copy(&protoGroup, &acc)
	// 	if err != nil {
	// 		srv.logger.Errorw("invalid account conversion", "error", err.Error())
	// 		return nil, status.Errorf(codes.Internal, "could not update account")
	// 	}
	// 	proto.Merge(&protoGroup, in.Group)

	// 	err = srv.repo.Update(ctx, &models.OneGroupFilter{ID: id}, &models.GroupPayload{Name: &protoAccount.Email, Name: &protoAccount.Name})
	// 	if err != nil {
	// 		srv.logger.Errorw("failed to update account", "error", err.Error())
	// 		return nil, status.Errorf(codes.Internal, "could not update account")
	// 	}
	// 	protoAccount.Id = id.String()
	return &accountsv1.UpdateGroupResponse{}, nil
}

func (srv *groupsService) GetGroup(ctx context.Context, in *accountsv1.GetGroupRequest) (*accountsv1.GetGroupResponse, error) {
	id, err := uuid.Parse(in.Id)
	if err != nil {
		srv.logger.Errorw("failed to convert uuid from string", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not get group")
	}
	accountModel, err := srv.repo.Get(ctx, &models.OneGroupFilter{ID: id.String(), Name: &in.Name})
	if err != nil {
		srv.logger.Errorw("failed to get group", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not get group")
	}

	var members []*accountsv1.GroupMember
	for _, m := range *accountModel.Members {
		memberId, err := uuid.Parse(m.ID)
		if err != nil {
			srv.logger.Errorw("failed to convert members uuid from string", "error", err.Error())
			return nil, status.Errorf(codes.Internal, "could not get group")
		}
		members = append(members, &accountsv1.GroupMember{AccountId: memberId.String()})
	}

	account := accountsv1.Group{Id: in.Id, Name: *accountModel.Name, Members: members, Description: *accountModel.Description}
	return &accountsv1.GetGroupResponse{Group: &account}, nil
}

func (srv *groupsService) JoinGroup(ctx context.Context, in *accountsv1.JoinGroupRequest) (*accountsv1.JoinGroupResponse, error) {
	id, err := uuid.Parse(in.Id)
	if err != nil {
		srv.logger.Errorw("failed to convert uuid from string", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not join group")
	}

	memberId, err := uuid.Parse(in.MemberId)
	if err != nil {
		srv.logger.Errorw("failed to convert uuid from string", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not join group")
	}

	acc, err := srv.repo.Get(ctx, &models.OneGroupFilter{ID: id.String()})
	if err != nil {
		srv.logger.Errorw("failed to get group", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not join group")
	}

	newMember := *acc.Members
	newMember = append(newMember, models.Member{ID: memberId.String()})

	err = srv.repo.Update(ctx, &models.OneGroupFilter{ID: id.String()}, &models.GroupPayload{Members: &newMember})
	if err != nil {
		srv.logger.Errorw("failed to update group", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not join group")
	}
	return &accountsv1.JoinGroupResponse{}, nil
}

func (srv *groupsService) AddNoteToGroup(ctx context.Context, in *accountsv1.AddNoteToGroupRequest) (*accountsv1.AddNoteToGroupResponse, error) {
	id, err := uuid.Parse(in.Id)
	if err != nil {
		srv.logger.Errorw("failed to convert uuid from string", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not join group")
	}

	noteId, err := uuid.Parse(in.NoteId)
	if err != nil {
		srv.logger.Errorw("failed to convert uuid from string", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not join group")
	}

	acc, err := srv.repo.Get(ctx, &models.OneGroupFilter{ID: id.String()})
	if err != nil {
		srv.logger.Errorw("failed to get group", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not join group")
	}

	newNote := *acc.Notes
	newNote = append(newNote, models.Note{ID: noteId.String()})

	err = srv.repo.Update(ctx, &models.OneGroupFilter{ID: id.String()}, &models.GroupPayload{Notes: &newNote})
	if err != nil {
		srv.logger.Errorw("failed to update group", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "could not join group")
	}
	return &accountsv1.AddNoteToGroupResponse{}, nil
}
