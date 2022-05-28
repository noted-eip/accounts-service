package main

import (
	"os"

	"accounts-service/auth"
	"accounts-service/models"

	_ "github.com/joho/godotenv/autoload"

	"google.golang.org/grpc"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("accounts-service", "Accounts service for the Noted backend").DefaultEnvars()

	environment   = app.Flag("env", "either development or production").Default(envIsProd).Enum(envIsProd, envIsDev)
	port          = app.Flag("port", "grpc server port").Default("3000").Int16()
	databaseUri   = app.Flag("database-uri", "uri of the database").Default("mongodb://localhost:27017").String()
	jwtPrivateKey = app.Flag("jwt-private-key", "base64 encoded ed25519 private key").Default("SGfCQAb05CtmhEesWxcrfXSQR6JjmEMeyjR7Mo21S60ZDW9VVTUuCvEMlGjlqiw4I/z8T11KqAXexvGIPiuffA==").String()
)

var (
	envIsProd = "production"
	envIsDev  = "development"
)

func main() {
	app.Parse(os.Args[1:])

	// We should remove this once the revamp of the models package is merged.
	models.Init(*databaseUri)

	s := &server{}
	s.Init(grpc.ChainUnaryInterceptor(s.LoggerUnaryInterceptor, auth.ForwardAuthMetadatathUnaryInterceptor))
	s.Run()
	defer s.Close()
}
