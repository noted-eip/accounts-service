rm -rf protorepo/noted/accounts/v1/*pb/
protoc --go_out=. --go-grpc_out=. protorepo/noted/accounts/v1/*.proto