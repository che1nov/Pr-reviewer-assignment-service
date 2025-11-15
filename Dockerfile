# syntax=docker/dockerfile:1

ARG GO_VERSION=1.24

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /src

RUN apk add --no-cache build-base

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /build/service ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates postgresql-client

WORKDIR /app

COPY --from=builder /build/service ./service
COPY --from=builder /src/openapi.yml ./openapi.yml

EXPOSE 8080

CMD ["/app/service"]

