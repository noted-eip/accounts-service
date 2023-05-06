package main

import (
	"accounts-service/auth"
	"accounts-service/communication"
	"accounts-service/models"
	"accounts-service/models/mongo"
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct {
	logger *zap.Logger

	authService auth.Service

	mongoDB *mongo.Database

	accountsRepository models.AccountsRepository

	accountsService accountsv1.AccountsAPIServer
	noteService     *communication.NoteServiceClient

	grpcServer *grpc.Server

	googleOauthConfig *oauth2.Config
}

// Init initializes the dependencies of the server and panics on error.
func (s *server) Init(opt ...grpc.ServerOption) {
	s.initLogger()
	s.InitEnv()
	s.initAuthService()
	s.initRepositories()
	s.initNoteServiceClient()
	s.initAccountsAPI()
	s.initGrpcServer(opt...)
}

func (s *server) Run() {
	lis, err := net.Listen("tcp", fmt.Sprint(":", *port))
	must(err, "failed to create tcp listener")
	reflection.Register(s.grpcServer)
	s.logger.Info(fmt.Sprint("service running on :", *port))
	err = s.grpcServer.Serve(lis)
	must(err, "failed to run grpc server")
}

func (s *server) Close() {
	s.logger.Info("graceful shutdown")
	s.mongoDB.Disconnect(context.Background())
	s.noteService.Close()
	s.logger.Sync()
}

func (s *server) LoggerUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
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
		zap.String("will", method),
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
}

func (s *server) initNoteServiceClient() {
	noteService, err := communication.NewNoteServiceClient(*noteServiceUrl)
	if *environment == envIsDev && err != nil {
		s.logger.Warn(fmt.Sprintf("could not instantiate note service connection: %v", err))
		noteService = nil
	} else {
		must(err, "could not instantiate note service connection")
	}
	s.noteService = noteService
}

func (s *server) InitEnv() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}
	var GOOGLE_APP_ID string = os.Getenv("GOOGLE_APP_ID")
	var GOOGLE_APP_SECRET string = os.Getenv("GOOGLE_APP_SECRET")
	var GOOGLE_REDIRECT_URI string = os.Getenv("GOOGLE_REDIRECT_URI")
	s.googleOauthConfig = &oauth2.Config{
		RedirectURL:  GOOGLE_REDIRECT_URI,
		ClientID:     GOOGLE_APP_ID,
		ClientSecret: GOOGLE_APP_SECRET,
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint: google.Endpoint,
	}
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

func (s *server) initAccountsAPI() {

	mailService := mailingAPI{
		logger: s.logger,
		repo:   s.accountsRepository,
		secret: *gmailSuperSecret,
	}

	s.accountsService = &accountsAPI{
		noteService: s.noteService,
		mailService: mailService,
		auth:        s.authService,
		logger:      s.logger,
		repo:        s.accountsRepository,
		googleOAuth: s.googleOauthConfig,
	}
}

func (s *server) initGrpcServer(opt ...grpc.ServerOption) {
	s.grpcServer = grpc.NewServer(opt...)
	accountsv1.RegisterAccountsAPIServer(s.grpcServer, s.accountsService)
}

func must(err error, msg string) {
	if err != nil {
		panic(fmt.Errorf("%s: %v", msg, err))
	}
}
