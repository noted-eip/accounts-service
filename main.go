package main

import (
	"accounts-service/grpc/accountspb"
	"net"

	"google.golang.org/grpc"
)

func main() {
	srv := grpc.NewServer()
	accSrv := accountsService{}
	accountspb.RegisterAccountsServiceServer(srv, &accSrv)

	lis, err := net.Listen("tcp", ":3000")
	if err != nil {
		panic(err)
	}
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
