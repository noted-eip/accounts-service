# Accounts service

Service responsible for managing accounts and authentication.

## Build

To run the service you only need to have [golang](https://go.dev) or [docker](https://docs.docker.com/get-docker/) installed.

After cloning the repository run:

```
make update-submodules
```
This will update the git submodules referencing our gRPC models and gRPC API definition.


You can then build the project by running the following command :

```
make re
```

Or by building and running the Dockerfile.

## Configuration
### CLI configuration

| Env Name                           | Flag Name           | Default                     | Description                               |
|------------------------------------|---------------------|-----------------------------|-------------------------------------------|
| `ACCOUNTS_SERVICE_PORT`            | `--port`            | `3000`                      | The port the application shall listen on. |
| `ACCOUNTS_SERVICE_ENV`             | `--env`             | `production`                | Either `production` or `development`.     |
| `ACCOUNTS_SERVICE_MONGO_URI`       | `--mongo-uri`       | `mongodb://localhost:27017` | Address of the MongoDB server.            |
| `ACCOUNTS_SERVICE_MONGO_DB_NAME`   | `--mongo-db-name`   | `accounts-service`          | Name of the Mongo database.               |
| `ACCOUNTS_SERVICE_JWT_PRIVATE_KEY` | `--jwt-private-key` | -                           | Base64 encoded ed25519 private key.       |
| `ACCOUNTS_SERVICE_GMAIL_SUPER_SECRET`   | `--gmail-super-secret`   |         | Gmail secret to send emails.               |
| `ACCOUNTS_SERVICE_ACCOUNT_SERVICE_URL`   | `--account-service-url`   | `notes.noted.koyeb:3000`          | Notes service's address               |

### Other env variables

| Env Name                           | Description                               |
|------------------------------------|-------------------------------------------|
| `JSON_FIREBASE_CREDS_B64`            | Firebase credentials used to connect to converted in base 64  |
| `GOOGLE_SECRET_AUTH`            | Google secret used to use for OAuth2|
| `FIREBASE_PROJECT_NB`            | Firebase project number identifier|


## Authentication

The accounts service expects the `Authorization` header to be set to a Bearer JWT on requests that require authentication.

```
Authorization: Bearer <token>
```
