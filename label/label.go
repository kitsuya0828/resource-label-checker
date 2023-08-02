package label

import (
	"context"
	"fmt"
	"time"

	"github.com/Kitsuya0828/resource-label-checker/fileio"
	"github.com/Kitsuya0828/resource-label-checker/slack"
)

type Result struct {
	NoRequiredLabelResources map[string][]string
	BannedLabelResources     map[string][]string
}

type CloudService interface {
	FilterLabels(context.Context) (*Result, error)
	Close() error
}

func SearchLabelAndNotify(ctx context.Context, svc CloudService, cfg slack.Config, outputDir string, scope string) error {
	// Search resources and filter labels (tags)
	result, err := svc.FilterLabels(ctx)
	if err != nil {
		return err
	}

	// Write out results to csv files and compress them into a zip file
	if err = fileio.WriteResultToCsvFiles(result.NoRequiredLabelResources, outputDir); err != nil {
		return err
	}
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return err
	}
	date := time.Now().In(jst).Format("2006-01-02")
	zipFileName := fmt.Sprintf("%s/%s_%s.zip", outputDir, date, scope)
	if err = fileio.CreateZip(outputDir, zipFileName, ".csv"); err != nil {
		return err
	}

	// Notify Slack
	slackService := slack.New(cfg.SlackToken, cfg.SlackChannelID)
	mainMessage := GetMainMessageText(result, scope, date)
	ts, err := slackService.SendText(mainMessage, "", true)
	if err != nil {
		return err
	}
	if err := slackService.SendFile(zipFileName, ts, ZipFileTitle); err != nil {
		return err
	}
	threadMessage := GetThreadMessageText(result)
	if _, err = slackService.SendText(threadMessage, ts, false); err != nil {
		return err
	}
	return nil
}

func CheckRequiredLabels(requiredLabels []string, gotLabels map[string]string) bool {
	if len(gotLabels) == 0 {
		return false
	}
	for _, requiredLabel := range requiredLabels {
		if _, ok := gotLabels[requiredLabel]; !ok {
			return false
		}
	}
	return true
}

func CheckBannedLabels(bannedLabels []string, gotLabels map[string]string) bool {
	if len(bannedLabels) == 0 {
		return true
	}
	for _, bannedLabel := range bannedLabels {
		if _, ok := gotLabels[bannedLabel]; ok {
			return false
		}
	}
	return true
}
