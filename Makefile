# Run the protoc compiler to generate the Golang server code.
codegen:
	protoc --go_out=. \
		--go_opt=Mapi/protos/accounts/accounts.proto=api/accountspb api/protos/accounts/accounts.proto \
		--go_opt=Mapi/protos/accounts/auth.proto=api/accountspb api/protos/accounts/auth.proto \
		--go-grpc_out=.
		

# Fetch the latest version of the protos submodule.
update-submodules:
	git submodule update --remote
