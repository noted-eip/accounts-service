FROM golang:1.18.1-alpine as build

WORKDIR /app

COPY . .

# Install proto-dependencies and generate
RUN apk update && apk add --no-cache make git protobuf-dev=3.18.1-r1
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
RUN ./misc/gen_proto.sh

RUN go build .

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/accounts-service .

ENTRYPOINT [ "./accounts-service" ]
