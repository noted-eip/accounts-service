# Run the protoc compiler to generate the Golang server code.
codegen: update-submodules
	docker run --rm -v `pwd`/grpc:/app/grpc -v `pwd`/misc:/app/misc -w /app noted-go-protoc /bin/sh -c misc/gen_proto.sh

# Run MongoDB database as a docker container
run-db:
	docker run --detach --publish 27017:27017 mongo

# Fetch the latest version of the protos submodule.
update-submodules:
	git submodule update --remote

# After cloning the repo, run init
init:
	git submodule init
	docker build -t noted-go-protoc -f misc/Dockerfile .
