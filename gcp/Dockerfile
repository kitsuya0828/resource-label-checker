# syntax=docker/dockerfile:1

## Build
FROM golang:1.20-buster AS builder
WORKDIR /app
COPY . .
RUN go mod download
WORKDIR /app/gcp
RUN go build -o /main

## Deploy
FROM alpine:latest
WORKDIR /
COPY --from=builder /main /main
ENTRYPOINT ["/main"]