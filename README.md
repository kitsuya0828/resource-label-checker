# resource-label-checker

**resource-label-checker** is a tool that periodically checks whether cloud resources are given a common label and notifies the results to Slack. 

The purpose is to effortlessly manage and analyse costs by labelling resources consistently.

## Setup
```bash
git clone https://github.com/Kitsuya0828/resource-label-checker.git
cd resource-label-checker
go mod download
```

## Run Locally
Set the environment variables indicated in `gcp.env.template` or `aws.env.template` appropriately before running the program. 

### Google Cloud
By default, it checks for the presence of labels listed in the `required-labels:` field in `gcp/config.yml`. 

In addition, because of the [Cloud Asset API](https://cloud.google.com/asset-inventory/docs/apis?hl=ja), the resource types to be excluded from the search can be selected from [Supported resource types](https://cloud.google.com/asset-inventory/docs/supported-asset-types?hl=ja#supported_resource_types) and appended to the `excluded-resources:` field.

Note that you must be granted the `cloudasset.assets.searchAllResources` permission on the desired scope (`PROJECT_ID`).

```bash
cd gcp
go run .
```

### AWS
By default, it checks for the presence of tags listed in the `required-labels:` field in `aws/config.yml`. 

In addition, because of the [AWS Config](https://docs.aws.amazon.com/config/latest/developerguide/WhatIsConfig.html), the resource types that can be included in the search are listed in [resourceType](https://docs.aws.amazon.com/config/latest/APIReference/API_ListDiscoveredResources.html#config-ListDiscoveredResources-request-resourceType) and you can edit the `included-resources:` field.

Note that you must be able to run `config:ListDiscoveredResources` action and `config:BatchGetResourceConfig` action.

```bash
cd aws
go run .
```

## Deploy
Sample Terraform code can be found in [Kitsuya0828/resource\-label\-checker\-terraform](https://github.com/Kitsuya0828/resource-label-checker-terraform).

The GitHub Actions workflow builds and pushes Docker images to the target Artifact Registry or ECR, but the sample workflow under the `.github/` directory is currently commented out.

To run in AWS Lambda, change the comment-out in `aws/main.go` as follows.

```go
func main() {
	// Run()	// Run locally or on ECS Fargate
	lambda.Start(Run)	// Run on AWS Lambda
}
```