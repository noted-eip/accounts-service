# Accounts service

Service responsible for managing accounts and authentication.

## Build

To run the service you only need to have [golang](https://go.dev) and [docker](https://docs.docker.com/get-docker/) installed.

Upon cloning the repository run:

```
make update-submodules
```

You can then build the project using the go toolchain.

## Configuration

| Env Name                           | Flag Name           | Default                     | Description                               |
|------------------------------------|---------------------|-----------------------------|-------------------------------------------|
| `ACCOUNTS_SERVICE_PORT`            | `--port`            | `3000`                      | The port the application shall listen on. |
| `ACCOUNTS_SERVICE_ENV`             | `--env`             | `production`                | Either `production` or `development`.     |
| `ACCOUNTS_SERVICE_MONGO_URI`       | `--mongo-uri`       | `mongodb://localhost:27017` | Address of the MongoDB server.            |
| `ACCOUNTS_SERVICE_MONGO_DB_NAME`   | `--mongo-db-name`   | `accounts-service`          | Name of the Mongo database.               |
| `ACCOUNTS_SERVICE_JWT_PRIVATE_KEY` | `--jwt-private-key` | -                           | Base64 encoded ed25519 private key.       |

## Authentication

The accounts service expects the `Authorization` header to be set to a Bearer JWT on requests that require authentication.

```
Authorization: Bearer <token>
```
