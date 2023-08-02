package main

import (
	"context"

	"log"

	"github.com/Kitsuya0828/resource-label-checker/label"
	"github.com/Kitsuya0828/resource-label-checker/slack"
	"github.com/caarlos0/env/v9"
)

func main() {
	ctx := context.Background()

	// Parse Slack environment variables
	slackCfg := slack.Config{}
	if err := env.Parse(&slackCfg); err != nil {
		log.Fatal(err)
	}

	// Create cloud service
	svc, err := NewService(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer svc.Close()

	// Search resources tags and notify Slack
	outputDir := "result"
	if err := label.SearchLabelAndNotify(ctx, svc, slackCfg, outputDir, svc.projectId); err != nil {
		log.Fatal(err)
	}
	log.Println("Successfully notified Slack")
}
