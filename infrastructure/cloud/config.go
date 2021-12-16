package cloud

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var defaultAwsCfgMu sync.Once
var DefaultAwsConfig aws.Config
var DefaultAmazonDynamoDbClient *dynamodb.Client

func init() {
	defaultAwsCfgMu.Do(func() {
		DefaultAwsConfig, _ = config.LoadDefaultConfig(context.TODO())
		DefaultAmazonDynamoDbClient = dynamodb.NewFromConfig(DefaultAwsConfig)
	})
}
