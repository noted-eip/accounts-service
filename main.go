package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"net"
	"os"

	"accounts-service/auth"
	"accounts-service/grpc/accountspb"

	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("accounts-service", "Accounts service for the Noted backend").DefaultEnvars()

	// This is temporary. In the future we will want to have a KMS and eventually read the keys from
	// a file or a mounted secret.
	jwtPrivateKey = app.Flag("jwt-private-key", "base64 encoded ed25519 private key").Default("SGfCQAb05CtmhEesWxcrfXSQR6JjmEMeyjR7Mo21S60ZDW9VVTUuCvEMlGjlqiw4I/z8T11KqAXexvGIPiuffA==").String()
	port          = app.Flag("port", "grpc server port").Default("3000").Int16()
	environment   = app.Flag("environment", "either development or production").Default("development").String()
)

func main() {
	app.Parse(os.Args[1:])

	authService := newAuthService()
	logger := newLogger()
	defer logger.Sync()

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(auth.ForwardAuthMetadatathUnaryInterceptor),
	)

	accSrv := accountsService{
		logger: logger,
		auth:   authService,
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
