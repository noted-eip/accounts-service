package main

// import (
// 	"accounts-service/grpc/accountspb"
// 	"context"

// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// )

// type authService struct {
// 	accountspb.UnimplementedAuthServiceServer
// }

// var _ accountspb.AuthServiceServer = &authService{}

// func (srv authService) Authenticate(context.Context, *accountspb.AuthenticateRequest) (*accountspb.AuthenticateReply, error) {
// 	return nil, status.Errorf(codes.Unimplemented, "not implemented")
// }
