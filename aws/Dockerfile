# syntax=docker/dockerfile:1

## Build
FROM golang:1.20-buster AS builder
WORKDIR /app
COPY . .
RUN go mod download
WORKDIR /app/aws
RUN GOOS=linux GOARCH=amd64 go build -o /main

## Deploy for ECS Fargate
FROM gcr.io/distroless/base-debian10
WORKDIR /
COPY --from=builder /main /main
ENTRYPOINT ["/main"]

## Deploy for ECS Fargate
# FROM public.ecr.aws/lambda/provided:al2
# WORKDIR /
# COPY --from=build /main /main
# ENTRYPOINT ["/main"]