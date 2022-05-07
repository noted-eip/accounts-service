package main

import (
	"fmt"
	"net"
	"os"

	"accounts-service/auth"
	"accounts-service/grpc/accountspb"

	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("accounts-service", "Accounts service for the Noted backend").DefaultEnvars()

	port        = app.Flag("port", "grpc server port").Default("3000").Int16()
	environment = app.Flag("environment", "either development or production").Default("development").String()
)

func main() {
	app.Parse(os.Args[1:])

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(auth.ForwardAuthMetadatathUnaryInterceptor),
	)

	accSrv := accountsService{}
	accountspb.RegisterAccountsServiceServer(srv, &accSrv)

	startServer(srv)
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
