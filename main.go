package main

import (
	"accounts-service/grpc/accounts"
	"net"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":3000")
	if err != nil {
		panic(err)
	}
	srv := grpc.NewServer()
	accSrv := accountsService{}
	accounts.RegisterAccountsServiceServer(srv, &accSrv)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
