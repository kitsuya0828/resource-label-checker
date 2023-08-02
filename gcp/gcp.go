package main

import (
	"context"
	"fmt"

	_ "embed"

	asset "cloud.google.com/go/asset/apiv1"
	"cloud.google.com/go/asset/apiv1/assetpb"
	"github.com/Kitsuya0828/resource-label-checker/fileio"
	"github.com/Kitsuya0828/resource-label-checker/label"
	"github.com/caarlos0/env/v9"
	"google.golang.org/api/iterator"
)

//go:embed config.yml
var data []byte

type config struct {
	ProjectID string `env:"PROJECT_ID,notEmpty"`
}

type GcpService struct {
	client    *asset.Client
	projectId string
}

func NewService(ctx context.Context) (*GcpService, error) {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	c, err := asset.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	svc := &GcpService{
		client:    c,
		projectId: cfg.ProjectID,
	}
	return svc, nil
}

func (svc *GcpService) FilterLabels(ctx context.Context) (*label.Result, error) {
	excludedAssetTypes, err := fileio.GetStringCollection(data, "excluded-resources")
	if err != nil {
		return nil, err
	}
	excludedAssetTypesMap := make(map[string]bool)
	for _, t := range excludedAssetTypes {
		excludedAssetTypesMap[t] = true
	}

	c := svc.client
	req := &assetpb.SearchAllResourcesRequest{
		Scope: fmt.Sprintf("projects/%s", svc.projectId),
	}
	it := c.SearchAllResources(ctx, req)

	requiredLabels, err := fileio.GetStringCollection(data, "required-labels")
	if err != nil {
		return nil, err
	}
	bannedLabels, err := fileio.GetStringCollection(data, "required-labels")
	if err != nil {
		return nil, err
	}
	noRequiredLabelResources := make(map[string][]string)
	bannedLabelResources := make(map[string][]string)

	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		if excludedAssetTypesMap[resp.AssetType] {
			continue
		}

		if ok := label.CheckRequiredLabels(requiredLabels, resp.Labels); !ok {
			noRequiredLabelResources[resp.AssetType] = append(noRequiredLabelResources[resp.AssetType], resp.Name)
		}
		if ok := label.CheckBannedLabels(bannedLabels, resp.Labels); !ok {
			noRequiredLabelResources[resp.AssetType] = append(noRequiredLabelResources[resp.AssetType], resp.Name)
		}
	}

	result := &label.Result{
		NoRequiredLabelResources: noRequiredLabelResources,
		BannedLabelResources:     bannedLabelResources,
	}
	return result, nil
}

func (svc *GcpService) Close() error {
	if err := svc.client.Close(); err != nil {
		return err
	}
	return nil
}
