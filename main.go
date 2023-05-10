package main

import (
	"os"

	"accounts-service/auth"

	"google.golang.org/grpc"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("accounts-service", "Accounts service for the Noted backend").DefaultEnvars()

	environment      = app.Flag("env", "either development or production").Default(envIsProd).Enum(envIsProd, envIsDev)
	port             = app.Flag("port", "grpc server port").Default("3000").Int16()
	noteServiceUrl   = app.Flag("note-service-url", "address of the note-service url to communicate with").Default("notes.noted.koyeb:3000").String()
	mongoUri         = app.Flag("mongo-uri", "address of the mongodb server").Default("mongodb://localhost:27017").String()
	mongoDbName      = app.Flag("mongo-db-name", "name of the mongo database").Default("accounts-service").String()
	jwtPrivateKey    = app.Flag("jwt-private-key", "base64 encoded ed25519 private key").Default("SGfCQAb05CtmhEesWxcrfXSQR6JjmEMeyjR7Mo21S60ZDW9VVTUuCvEMlGjlqiw4I/z8T11KqAXexvGIPiuffA==").String()
	gmailSuperSecret = app.Flag("gmail-super-secret", "token to authenticate accounts service with noted gmail account").Default("").String()
	googleAuthSecret = app.Flag("google-auth-secret", "google auth secret").Default("").String()
)

var (
	envIsProd = "production"
	envIsDev  = "development"
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	s := &server{}
	s.Init(grpc.ChainUnaryInterceptor(s.LoggerUnaryInterceptor, auth.ForwardAuthMetadatathUnaryInterceptor))
	s.Run()
	defer s.Close()
}
