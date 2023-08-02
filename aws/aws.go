package main

import (
	"context"
	"encoding/json"
	"log"

	_ "embed"

	"github.com/Kitsuya0828/resource-label-checker/fileio"
	"github.com/Kitsuya0828/resource-label-checker/label"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/caarlos0/env/v9"
)

//go:embed config.yml
var data []byte

type AwsService struct {
	client      *configservice.ConfigService
	accountName string
}

type config struct {
	Region      string `env:"REGION" envDefault:"ap-northeast-1"`
	AccountName string `env:"ACCOUNT_NAME,notEmpty"`
}

func NewService(ctx context.Context) (*AwsService, error) {
	// Parse environment variables
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Region: aws.String(cfg.Region)},
	}))
	// Create a ConfigService client with additional configuration
	c := configservice.New(sess)
	svc := &AwsService{
		client:      c,
		accountName: cfg.AccountName,
	}
	return svc, nil
}

type ResourceInfo struct {
	Type string
	Arn  string
}

type Configuration struct {
	Tags []map[string]string `json:"tags"`
}

func (svc *AwsService) FilterLabels(ctx context.Context) (*label.Result, error) {

	includedResourceTypes, err := fileio.GetStringCollection(data, "included-resources")
	if err != nil {
		return nil, err
	}

	requiredLabels, err := fileio.GetStringCollection(data, "required-labels")
	if err != nil {
		return nil, err
	}
	bannedLabels, err := fileio.GetStringCollection(data, "banned-labels")
	if err != nil {
		return nil, err
	}

	noRequiredLabelResources := make(map[string][]string)
	bannedLabelResources := make(map[string][]string)

	// This process cannot be parallelized by goroutine
	// Causes "ThrottlingException: Rate exceeded" error
	for _, t := range includedResourceTypes {
		log.Printf("# Searching for resources of type: %s\n", t)
		input := &configservice.ListDiscoveredResourcesInput{
			ResourceType: aws.String(t),
		}

		for {
			output, err := svc.client.ListDiscoveredResources(input)
			if err != nil {
				log.Printf("##### API Request Error ######\nerror: %v\n", err)
				continue
			}
			for _, r := range output.ResourceIdentifiers {
				resourceKeys := []*configservice.ResourceKey{
					{
						ResourceId:   r.ResourceId,
						ResourceType: r.ResourceType,
					},
				}
				configInput := &configservice.BatchGetResourceConfigInput{
					ResourceKeys: resourceKeys,
				}
				configOutput, err := svc.client.BatchGetResourceConfig(configInput)
				if err != nil {
					log.Printf("##### API Request Error ######\nerror: %v\n", err)
					continue
				}

				items := configOutput.BaseConfigurationItems
				if len(items) != 0 {
					item := items[0]
					config := item.Configuration
					var cfg Configuration

					// TODO: Deal with json.Unmarshal errors (Mainly because `tags` field is not always `[]map[string]string` type)
					if err := json.Unmarshal([]byte(*config), &cfg); err != nil {
						log.Printf("##### Unmarshal Error ######\nconfig: %v\nerror: %v\n", *config, err)
						continue
					}

					tags := make(map[string]string)
					for _, tag := range cfg.Tags {
						var tagKey, tagValue string
						for prefix, v := range tag {
							if prefix == "key" {
								tagKey = v
							} else if prefix == "value" {
								tagValue = v
							}
						}
						if tagKey != "" && tagValue != "" {
							tags[tagKey] = tagValue
						}
					}

					// TODO: Some resources do not have `Arn` field
					if item.Arn == nil || item.ResourceType == nil {
						log.Printf("##### No ARN Error ######\nconfig: %v\n", *config)
						continue
					}

					if ok := label.CheckRequiredLabels(requiredLabels, tags); !ok {
						noRequiredLabelResources[*r.ResourceType] = append(noRequiredLabelResources[*r.ResourceType], *item.Arn)
					}
					if ok := label.CheckBannedLabels(bannedLabels, tags); !ok {
						bannedLabelResources[*r.ResourceType] = append(bannedLabelResources[*r.ResourceType], *item.Arn)
					}
				}
			}
			if output.NextToken != nil {
				input.SetNextToken(*output.NextToken)
			} else {
				break
			}
		}
	}

	result := &label.Result{
		NoRequiredLabelResources: noRequiredLabelResources,
		BannedLabelResources:     bannedLabelResources,
	}
	return result, nil
}

func (svc *AwsService) Close() error {
	return nil
}
