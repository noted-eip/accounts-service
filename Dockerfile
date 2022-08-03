FROM golang:1.18.1-alpine as build
WORKDIR /app
COPY . .
RUN go build .

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/accounts-service .
ENTRYPOINT [ "./accounts-service" ]
