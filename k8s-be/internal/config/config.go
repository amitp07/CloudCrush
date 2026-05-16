package config

import (
	"os"
)

type Config struct {
	NatsURL       string
	PostgresURL   string
	AwsS3Bucket   string
	AwsS3Endpoint string
	AwsRegion     string
}

const (
	DefaultNatsURL  = "nats://localhost:4222"
	DefaultPgURL    = "postgres://postgres:postgres@localhost:5432/k8s-db?sslmode=disable"
	DefaultBucket   = "cloudcrush-bucket"
	DefaultRegion   = "us-east-1"
	DefaultEndpoint = "http://minio-service.default.svc.cluster.local:9000"
)

func Load() Config {
	natsUrl := os.Getenv("NATS_DSN")
	if natsUrl == "" {
		natsUrl = DefaultNatsURL
	}

	pgUrl := os.Getenv("POSTGRES_DSN")
	if pgUrl == "" {
		pgUrl = DefaultPgURL
	}

	bucket := os.Getenv("AWS_IMAGE_JOB_BUCKET")
	if bucket == "" {
		bucket = DefaultBucket
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = DefaultRegion
	}

	s3Endpoint := os.Getenv("AWS_S3_ENDPOINT")
	if s3Endpoint == "" {
		s3Endpoint = DefaultEndpoint
	}

	return Config{
		NatsURL:       natsUrl,
		PostgresURL:   pgUrl,
		AwsS3Bucket:   bucket,
		AwsS3Endpoint: s3Endpoint,
		AwsRegion:     region,
	}
}
