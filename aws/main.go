package main

import (
	"context"

	"log"

	"github.com/Kitsuya0828/resource-label-checker/label"
	"github.com/Kitsuya0828/resource-label-checker/slack"
	// "github.com/aws/aws-lambda-go/lambda"
	"github.com/caarlos0/env/v9"
)

func Run() {
	ctx := context.Background()

	// Parse Slack environment variables
	cfg := slack.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	// Create cloud service
	svc, err := NewService(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer svc.Close()

	// Search resources tags and notify Slack
	outputDir := "/tmp/resource-label-checker"
	if err := label.SearchLabelAndNotify(ctx, svc, cfg, outputDir, svc.accountName); err != nil {
		log.Fatal(err)
	}
	log.Println("Successfully notified Slack")
}

func main() {
	Run()	// Run locally or on ECS Fargate
	// lambda.Start(Run)	// Run on AWS Lambda
}
