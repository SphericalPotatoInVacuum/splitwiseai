package db

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"go.uber.org/zap"
)

type Config struct {
	Endpoint     string `env:"DB_ENDPOINT"`
	AwsKeyId     string `env:"DB_AWS_KEY_ID"`
	AwsSecretKey string `env:"DB_AWS_SECRET_KEY"`
}

func NewClient(cfg Config) *dynamodb.Client {
	zap.S().Debug("Creating DynamoDB client")

	db := dynamodb.NewFromConfig(
		aws.Config{
			Region:      "ru-central1",
			Credentials: credentials.NewStaticCredentialsProvider(cfg.AwsKeyId, cfg.AwsSecretKey, ""),
		},
		func(o *dynamodb.Options) {
			o.BaseEndpoint = &cfg.Endpoint
		},
	)

	return db
}
