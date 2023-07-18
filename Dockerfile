FROM golang:1.20.4-alpine as build
WORKDIR /app
COPY . .
# Required for VCS stamping.
RUN apk add --no-cache git
RUN go build .

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/accounts-service .
COPY --from=build /app/mail.html .
ENTRYPOINT [ "./accounts-service" ]
