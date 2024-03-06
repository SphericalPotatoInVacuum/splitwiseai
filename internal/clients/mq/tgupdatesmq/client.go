package tgupdatesmq

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"go.uber.org/zap"
)

type client struct {
	svc      *sqs.Client
	queueUrl string
}

type Config struct {
	Endpoint     string `env:"TG_UPDATES_MQ_ENDPOINT"`
	QueueUrl     string `env:"TG_UPDATES_MQ_QUEUE_URL"`
	AwsKeyId     string `env:"DB_AWS_KEY_ID"`
	AwsSecretKey string `env:"DB_AWS_SECRET_KEY"`
}

func NewClient(cfg Config) (*client, error) {
	zap.S().Debugw("Creating SQS client", "config", cfg)

	awsCfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("ru-central1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AwsKeyId, cfg.AwsSecretKey, "")),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           cfg.Endpoint,
					SigningRegion: "ru-central1",
				}, nil
			},
		)),
		config.WithClientLogMode(aws.LogRequestWithBody|aws.LogResponseWithBody|aws.LogRetries),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	svc := sqs.NewFromConfig(awsCfg)

	return &client{
		svc:      svc,
		queueUrl: cfg.QueueUrl,
	}, nil
}

func (c *client) PublishMessage(ctx context.Context, message string) error {
	output, err := c.svc.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody: &message,
		QueueUrl:    &c.queueUrl,
	})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	zap.S().Debugw("Message sent", "output", output)

	return nil
}
