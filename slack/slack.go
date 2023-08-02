package slack

import (
	"github.com/slack-go/slack"
)

type Service struct {
	cli       *slack.Client
	channelID string
}

type Config struct {
	SlackToken     string `env:"SLACK_TOKEN,notEmpty"`
	SlackChannelID string `env:"SLACK_CHANNEL_ID,notEmpty"`
}

func New(token string, channelID string) *Service {
	cli := slack.New(token)
	return &Service{
		cli:       cli,
		channelID: channelID,
	}
}

func (s *Service) SendText(text, ts string, emoji bool) (string, error) {
	params := slack.PostMessageParameters{}
	if emoji {
		params.LinkNames = 0
	}
	options := []slack.MsgOption{
		slack.MsgOptionText(text, false),
		slack.MsgOptionTS(ts),
		slack.MsgOptionPostMessageParameters(params),
	}
	_, ts, _, err := s.cli.SendMessage(s.channelID, options...)
	if err != nil {
		return "", err
	}
	return ts, nil
}

func (s *Service) SendFile(filePath string, ts string, title string) error {
	fileUploadParams := slack.FileUploadParameters{
		Channels:        []string{s.channelID},
		File:            filePath,
		ThreadTimestamp: ts,
		Title:           title,
	}
	_, err := s.cli.UploadFile(fileUploadParams)
	if err != nil {
		return err
	}
	return nil
}
