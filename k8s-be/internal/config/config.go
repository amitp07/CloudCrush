package config

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type Config struct {
	NatsURL       string
	PostgresURL   string
	AwsS3Bucket   string
	AwsS3Endpoint string
	AwsRegion     string
}

const (
	NATS_URL       = "/cloudcrush/sandbox/nats/url"
	POSTGRES_URL   = "/cloudcrush/sandbox/postgres/url"
	S3_BUCKET_NAME = "/cloudcrush/sandbox/s3/bucketName"
	AWS_REGION     = "ap-south-1"
)

func Load() Config {
	natsUrl := os.Getenv("NATS_DSN")
	pgUrl := os.Getenv("POSTGRES_DSN")
	bucket := os.Getenv("AWS_IMAGE_JOB_BUCKET")
	region := os.Getenv("AWS_REGION")

	if region == "" {
		region = AWS_REGION
	}

	if natsUrl == "" && pgUrl == "" {

		ctx := context.Background()

		awsConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))

		if err != nil {
			panic(err)
		}

		client := ssm.NewFromConfig(awsConfig)

		result, err := client.GetParametersByPath(ctx, &ssm.GetParametersByPathInput{
			Path:           aws.String("/cloudcrush"),
			Recursive:      aws.Bool(true),
			WithDecryption: aws.Bool(true),
		})

		if err != nil {
			panic(err)
		}

		m := map[string]string{}
		for _, param := range result.Parameters {
			m[*param.Name] = *param.Value
		}

		natsUrl = m[NATS_URL]
		pgUrl = m[POSTGRES_URL]
		bucket = m[S3_BUCKET_NAME]
	}

	return Config{
		NatsURL:     natsUrl,
		PostgresURL: pgUrl,
		AwsS3Bucket: bucket,
	}
}
