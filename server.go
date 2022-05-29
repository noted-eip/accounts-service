package main

import (
	"accounts-service/auth"
	"accounts-service/grpc/accountspb"
	"accounts-service/models"
	"accounts-service/models/mongo"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct {
	logger  *zap.Logger
	slogger *zap.SugaredLogger

	authService auth.Service

	mongoDB *mongo.Database

	accountsRepository models.AccountsRepository
	accountsService    accountspb.AccountsServiceServer

	grpcServer *grpc.Server
}

// Init initializes the dependencies of the server and panics on error.
func (s *server) Init(opt ...grpc.ServerOption) {
	s.initLogger()
	s.initAuthService()
	s.initRepositories()
	s.initAccountsService()
	s.initGrpcServer(opt...)
}

func (s *server) Run() {
	lis, err := net.Listen("tcp", fmt.Sprint(":", *port))
	if err != nil {
		panic(err)
	}
	reflection.Register(s.grpcServer)
	s.slogger.Infof("running service on :%d", *port)
	if err := s.grpcServer.Serve(lis); err != nil {
		panic(err)
	}
}

func (s *server) Close() {
	s.logger.Info("graceful shutdown")
	s.logger.Sync()
	s.mongoDB.Disconnect(context.Background())
}

func (s *server) LoggerUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	res, err := handler(ctx, req)
	end := time.Now()
	if err != nil {
		s.logger.Warn("failed rpc",
			zap.String("code", status.Code(err).String()),
			zap.String("method", info.FullMethod),
			zap.Duration("duration", end.Sub(start)),
			zap.Error(err),
		)
		return res, err
	}
	s.logger.Info("rpc",
		zap.String("code", status.Code(err).String()),
		zap.String("method", info.FullMethod),
		zap.Duration("duration", end.Sub(start)),
	)
	return res, nil
}

func (s *server) initLogger() {
	var err error
	if *environment == envIsProd {
		s.logger, err = zap.NewProduction(zap.AddStacktrace(zapcore.FatalLevel), zap.WithCaller(false))
	} else {
		s.logger, err = zap.NewDevelopment(zap.AddStacktrace(zapcore.FatalLevel), zap.WithCaller(false))
	}
	must(err, "unable to instantiate zap.Logger")
	s.slogger = s.logger.Sugar()
}

func (s *server) initAuthService() {
	rawKey, err := base64.StdEncoding.DecodeString(*jwtPrivateKey)
	must(err, "could not decode jwt private key")
	s.authService = auth.NewService(ed25519.PrivateKey(rawKey))
}

func (s *server) initRepositories() {
	var err error
	s.mongoDB, err = mongo.NewDatabase(context.Background(), *mongoUri, *mongoDbName, s.logger)
	must(err, "could not instantiate mongo database")
	s.accountsRepository = mongo.NewAccountsRepository(s.mongoDB.DB, s.logger)
}

func (s *server) initAccountsService() {
	s.accountsService = &accountsService{
		auth:   s.authService,
		logger: s.slogger,
		repo:   s.accountsRepository,
	}
}

func (s *server) initGrpcServer(opt ...grpc.ServerOption) {
	s.grpcServer = grpc.NewServer(opt...)
	accountspb.RegisterAccountsServiceServer(s.grpcServer, s.accountsService)
}

func must(err error, msg string) {
	if err != nil {
		panic(fmt.Errorf("%s: %v", msg, err))
	}
}