# Run the protoc compiler to generate the Golang server code.
codegen:
	protoc --go_out=. --go-grpc_out=. api/protos/accounts/*.proto
		

# Fetch the latest version of the protos submodule.
update-submodules:
	git submodule update --remote
