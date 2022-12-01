package main

import (
	"os"

	"accounts-service/auth"
	"accounts-service/config"
	"accounts-service/controllers"

	"google.golang.org/grpc"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("accounts-service", "Accounts service for the Noted backend").DefaultEnvars()

	Environment   = app.Flag("env", "either development or production").Default(config.EnvIsProd).Enum(config.EnvIsProd, config.EnvIsDev)
	port          = app.Flag("port", "grpc server port").Default("3000").Int16()
	MongoUri      = app.Flag("mongo-uri", "address of the mongodb server").Default("mongodb://localhost:27017").String()
	MongoDbName   = app.Flag("mongo-db-name", "name of the mongo database").Default("accounts-service").String()
	JwtPrivateKey = app.Flag("jwt-private-key", "base64 encoded ed25519 private key").Default("SGfCQAb05CtmhEesWxcrfXSQR6JjmEMeyjR7Mo21S60ZDW9VVTUuCvEMlGjlqiw4I/z8T11KqAXexvGIPiuffA==").String()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	s := &controllers.Server{}
	config.Conf = &config.Config{
		Environment:   Environment,
		MongoUri:      MongoUri,
		MongoDbName:   MongoDbName,
		JwtPrivateKey: JwtPrivateKey,
	}
	s.Init(grpc.ChainUnaryInterceptor(s.LoggerUnaryInterceptor, auth.ForwardAuthMetadatathUnaryInterceptor))
	s.Run(port)
	defer s.Close()
}
