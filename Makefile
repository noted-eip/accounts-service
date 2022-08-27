# Run this command upon cloning the repository or when wanting to
# fetch the latest version of the protorepo.
update-submodules:
	git submodule update --init --remote

# Run the golangci-lint linter.
lint:
	docker run -w /app -v `pwd`:/app:ro golangci/golangci-lint golangci-lint run

# Run MongoDB database as a docker container.
run-db:
	docker run --name accounts-mongo --detach --publish 27017:27017 mongo

# Stop MongoDB database.
stop-db:
	docker kill accounts-mongo
	docker rm accounts-mongo
