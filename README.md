# Accounts service

Service responsible for managing accounts and authentication.

## Build

To run the service you only need to have [golang](https://go.dev) and [docker](https://docs.docker.com/get-docker/) installed.

Upon cloning the repository run:
```
make init
```

To generate the protobuf server and stubs:
```
make codegen
```

You can then build the project using the go toolchain.
