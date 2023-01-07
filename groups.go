package main

import (
	"accounts-service/auth"
	"accounts-service/models"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"accounts-service/validators"
	"context"

	"github.com/mennanov/fmutils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type groupsAPI struct {
	accountsv1.UnimplementedGroupsAPIServer

	auth             auth.Service
	logger           *zap.Logger
	groupRepo        models.GroupsRepository
	memberRepo       models.MembersRepository
	conversationRepo models.ConversationsRepository
	noteRepo         models.GroupNotesRepository
}

var _ accountsv1.GroupsAPIServer = &groupsAPI{}

func (srv *groupsAPI) CreateGroup(ctx context.Context, in *accountsv1.CreateGroupRequest) (*accountsv1.CreateGroupResponse, error) {
	token, err := srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	accountId := token.UserID.String()

	group, err := srv.groupRepo.Create(ctx, &models.GroupPayload{Name: &in.Name, Description: &in.Description})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	member := models.MemberPayload{AccountID: &accountId, GroupID: &group.ID, Role: auth.RoleAdmin}
	_, err = srv.memberRepo.Create(ctx, &member)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = srv.conversationRepo.Create(ctx, &models.CreateConversationPayload{Title: "General conversation", GroupID: group.ID})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &accountsv1.CreateGroupResponse{
		Group: &accountsv1.Group{
			Id:          group.ID,
			Name:        *group.Name,
			Description: *group.Description,
			CreatedAt:   timestamppb.New(group.CreatedAt),
		},
	}, nil
}

func (srv *groupsAPI) DeleteGroup(ctx context.Context, in *accountsv1.DeleteGroupRequest) (*accountsv1.DeleteGroupResponse, error) {
	err := validators.ValidateDeleteGroupRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, err = srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	err = srv.groupRepo.Delete(ctx, &models.OneGroupFilter{ID: in.GroupId})
	if err != nil {
		return nil, statusFromModelError(err)
	}
	memberFilter := models.MemberFilter{GroupID: &in.GroupId}
	err = srv.memberRepo.DeleteMany(ctx, &memberFilter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &accountsv1.DeleteGroupResponse{}, nil
}

func (srv *groupsAPI) UpdateGroup(ctx context.Context, in *accountsv1.UpdateGroupRequest) (*accountsv1.UpdateGroupResponse, error) {

	err := validators.ValidateUpdatedGroupRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not update Group")
	}

	groupID := in.Group.Id
	fieldMask := in.GetUpdateMask()
	fieldMask.Normalize()
	if !fieldMask.IsValid(in.Group) {
		return nil, status.Error(codes.InvalidArgument, "invalid field mask")
	}

	allowList := []string{"description", "name"}
	fmutils.Filter(in.GetGroup(), fieldMask.GetPaths())
	fmutils.Filter(in.GetGroup(), allowList)

	_, err = srv.groupRepo.Update(ctx, &models.OneGroupFilter{ID: groupID}, &models.GroupPayload{Name: &in.GetGroup().Name, Description: &in.GetGroup().Description})
	if err != nil {
		return nil, statusFromModelError(err)
	}

	return &accountsv1.UpdateGroupResponse{Group: in.GetGroup()}, nil
}

func (srv *groupsAPI) GetGroup(ctx context.Context, in *accountsv1.GetGroupRequest) (*accountsv1.GetGroupResponse, error) {
	_, err := srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	err = validators.ValidateGetGroup(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not get Group")
	}

	group, err := srv.groupRepo.Get(ctx, &models.OneGroupFilter{ID: in.GroupId})
	if err != nil {
		srv.logger.Error("failed get group from group id", zap.Error(err))
		return nil, statusFromModelError(err)
	}

	return &accountsv1.GetGroupResponse{
		Group: &accountsv1.Group{
			Id:          group.ID,
			Description: *group.Description,
			Name:        *group.Name,
			CreatedAt:   timestamppb.New(group.CreatedAt),
		},
	}, nil
}

func (srv *groupsAPI) ListGroups(ctx context.Context, in *accountsv1.ListGroupsRequest) (*accountsv1.ListGroupsResponse, error) {
	err := validators.ValidateListGroups(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not validate list groups request")
	}

	_, err = srv.authenticate(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	if in.Limit == 0 {
		in.Limit = 10
	}

	memberFromGroups, err := srv.memberRepo.List(ctx, &models.MemberFilter{AccountID: &in.AccountId})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "could not list groups member from groups Id")
	}

	groups := []*accountsv1.Group{}
	for _, member := range memberFromGroups {
		group, err := srv.groupRepo.Get(ctx, &models.OneGroupFilter{ID: *member.GroupID})

		if err != nil {
			srv.logger.Error("failed get group from member id", zap.Error(err))
		}
		groups = append(groups, &accountsv1.Group{Id: group.ID, Description: *group.Description, Name: *group.Name, CreatedAt: timestamppb.New(group.CreatedAt)})
	}

	return &accountsv1.ListGroupsResponse{Groups: groups}, nil
}

// TODO: This function is duplicated from accountsService.authenticate().
// Find a way to extract this into a separate function or use a base class
// to share common behaviour.
func (srv *groupsAPI) authenticate(ctx context.Context) (*auth.Token, error) {
	token, err := srv.auth.TokenFromContext(ctx)
	if err != nil {
		srv.logger.Debug("could not authenticate request", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	return token, nil
}
