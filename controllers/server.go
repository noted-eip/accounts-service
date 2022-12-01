package controllers

import (
	"accounts-service/auth"
	"accounts-service/config"
	"accounts-service/models"
	"accounts-service/models/mongo"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type Server struct {
	logger *zap.Logger

	authService auth.Service

	mongoDB *mongo.Database

	accountsRepository models.AccountsRepository
	groupsRepository   models.GroupsRepository
	membersRepository  models.MembersRepository

	accountsService accountsv1.AccountsAPIServer
	groupsService   accountsv1.GroupsAPIServer

	grpcServer *grpc.Server
}

// Init initializes the dependencies of the server and panics on error.
func (s *Server) Init(opt ...grpc.ServerOption) {
	s.initLogger()
	s.initAuthService()
	s.initRepositories()
	s.initAccountsService()
	s.initGroupsService()
	s.initGrpcServer(opt...)
}

func (s *Server) Run(port *int16) {
	lis, err := net.Listen("tcp", fmt.Sprint(":", *port))
	must(err, "failed to create tcp listener")
	reflection.Register(s.grpcServer)
	s.logger.Info(fmt.Sprint("service running on :", *port))
	err = s.grpcServer.Serve(lis)
	must(err, "failed to run grpc server")
}

func (s *Server) Close() {
	s.logger.Info("graceful shutdown")
	s.mongoDB.Disconnect(context.Background())
	s.logger.Sync()
}

func (s *Server) LoggerUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	res, err := handler(ctx, req)
	end := time.Now()

	method := info.FullMethod[strings.LastIndexByte(info.FullMethod, '/')+1:]

	if err != nil {
		var displayErr = err
		st, ok := status.FromError(err)
		if ok {
			displayErr = errors.New(st.Message())
		}
		s.logger.Warn("failed rpc",
			zap.String("code", status.Code(err).String()),
			zap.String("method", method),
			zap.Duration("duration", end.Sub(start)),
			zap.Error(displayErr),
		)
		return res, err
	}

	s.logger.Info("rpc",
		zap.String("code", status.Code(err).String()),
		zap.String("method", method),
		zap.Duration("duration", end.Sub(start)),
	)

	return res, nil
}

func (s *Server) initLogger() {
	var err error
	if *config.Conf.Environment == config.EnvIsProd {
		s.logger, err = zap.NewProduction(zap.AddStacktrace(zapcore.FatalLevel), zap.WithCaller(false))
	} else {
		s.logger, err = zap.NewDevelopment(zap.AddStacktrace(zapcore.FatalLevel), zap.WithCaller(false))
	}
	must(err, "unable to instantiate zap.Logger")
}

func (s *Server) initAuthService() {
	rawKey, err := base64.StdEncoding.DecodeString(*config.Conf.JwtPrivateKey)
	must(err, "could not decode jwt private key")
	s.authService = auth.NewService(ed25519.PrivateKey(rawKey))
}

func (s *Server) initRepositories() {
	var err error
	s.mongoDB, err = mongo.NewDatabase(context.Background(), *config.Conf.MongoUri, *config.Conf.MongoDbName, s.logger)
	must(err, "could not instantiate mongo database")
	s.accountsRepository = mongo.NewAccountsRepository(s.mongoDB.DB, s.logger)
	s.groupsRepository = mongo.NewGroupsRepository(s.mongoDB.DB, s.logger)
	s.membersRepository = mongo.NewMembersRepository(s.mongoDB.DB, s.logger)
}

func (s *Server) initAccountsService() {
	s.accountsService = &AccountsAPI{
		Auth:   s.authService,
		Logger: s.logger,
		Repo:   s.accountsRepository,
	}
}

func (s *Server) initGroupsService() {
	s.groupsService = &GroupsAPI{
		Auth:       s.authService,
		Logger:     s.logger,
		GroupRepo:  s.groupsRepository,
		MemberRepo: s.membersRepository,
	}
}

func (s *Server) initGrpcServer(opt ...grpc.ServerOption) {
	s.grpcServer = grpc.NewServer(opt...)
	accountsv1.RegisterAccountsAPIServer(s.grpcServer, s.accountsService)
	accountsv1.RegisterGroupsAPIServer(s.grpcServer, s.groupsService)
}

func must(err error, msg string) {
	if err != nil {
		panic(fmt.Errorf("%s: %v", msg, err))
	}
}