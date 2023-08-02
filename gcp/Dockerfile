# syntax=docker/dockerfile:1

## Build
FROM golang:1.20-buster AS builder
WORKDIR /app
COPY . .
RUN go mod download
WORKDIR /app/gcp
RUN GOOS=linux GOARCH=amd64 go build -o /main

## Deploy
FROM gcr.io/distroless/base-debian10
WORKDIR /
COPY --from=builder /main /main
ENTRYPOINT ["/main"]