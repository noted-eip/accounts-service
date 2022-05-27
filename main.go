package main

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"time"

	"accounts-service/auth"
	"accounts-service/grpc/accountspb"
	"accounts-service/models"
	"accounts-service/models/mongo"

	_ "github.com/joho/godotenv/autoload"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("accounts-service", "Accounts service for the Noted backend").DefaultEnvars()

	// This is temporary. In the future we will want to have a KMS and eventually read the keys from
	// a file or a mounted secret.
	jwtPrivateKey = app.Flag("jwt-private-key", "base64 encoded ed25519 private key").Default("SGfCQAb05CtmhEesWxcrfXSQR6JjmEMeyjR7Mo21S60ZDW9VVTUuCvEMlGjlqiw4I/z8T11KqAXexvGIPiuffA==").String()
	port          = app.Flag("port", "grpc server port").Default("3000").Int16()
	databaseUri   = app.Flag("database-uri", "uri of the database").Default("mongodb://localhost:27017").String()
	environment   = app.Flag("environment", "either development or production").Default("development").String()
)

func main() {
	app.Parse(os.Args[1:])

	models.Init(*databaseUri)
	authService := newAuthService()
	logger := newLogger()
	defer logger.Sync()

	repo := mongo.NewAccountsRepository(logger)

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(loggerUnaryInterceptor(logger)),
	)

	accSrv := accountsService{
		logger: logger,
		auth:   authService,
		repo:   repo,
	}
	accountspb.RegisterAccountsServiceServer(srv, &accSrv)

	startServer(srv)
}

func newAuthService() auth.Service {
	rawKey, err := base64.StdEncoding.DecodeString(*jwtPrivateKey)
	if err != nil {
		panic("could not decode private key: " + err.Error())
	}
	return auth.NewService(ed25519.PrivateKey(rawKey))
}

func newLogger() *zap.SugaredLogger {
	var logger *zap.Logger
	var err error
	if *environment == "development" {
		logger, err = zap.NewDevelopment(zap.AddCaller(), zap.AddStacktrace(zapcore.FatalLevel))
	} else {
		logger, err = zap.NewProduction(zap.AddCaller(), zap.AddStacktrace(zapcore.FatalLevel))
	}
	if err != nil {
		panic("unable to create logger: " + err.Error())
	}
	return logger.Sugar()
}

func startServer(srv *grpc.Server) {
	lis, err := net.Listen("tcp", fmt.Sprint(":", *port))
	if err != nil {
		panic(err)
	}
	reflection.Register(srv)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}

// Later this will be moved in an observability package.
func loggerUnaryInterceptor(logger *zap.SugaredLogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		res, err := auth.ForwardAuthMetadatathUnaryInterceptor(ctx, req, info, handler)
		end := time.Now()

		fields := []zapcore.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("time", end.Sub(start)),
		}
		if peer, ok := peer.FromContext(ctx); ok {
			fields = append(fields, zap.String("peer", peer.Addr.String()))
		}
		if err != nil {
			fields = append(fields, zap.String("code", status.Code(err).String()), zap.String("error", err.Error()))
			logger.Desugar().Warn("failed rpc", fields...)
			return res, err
		}
		logger.Desugar().Info("rpc", fields...)
		return res, err
	}
}
