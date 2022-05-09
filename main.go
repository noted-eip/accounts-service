package main

import (
	"net"

	"accounts-service/auth"
	"accounts-service/grpc/accountspb"

	"accounts-service/models"

	_ "github.com/joho/godotenv/autoload"

	"google.golang.org/grpc"

	"google.golang.org/grpc/reflection"
)

func main() {
	models.Init()
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(auth.ForwardAuthMetadatathUnaryInterceptor),
	)

	accSrv := accountsService{}
	accountspb.RegisterAccountsServiceServer(srv, &accSrv)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		panic(err)
	}

	reflection.Register(srv)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
