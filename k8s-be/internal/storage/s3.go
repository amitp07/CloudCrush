package storage

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
	Client *s3.Client
	Bucket string
}

// create new S3 client
func NewS3Client(ctx context.Context) *S3Client {
	bucket := os.Getenv("AWS_IMAGE_JOB_BUCKET")
	endpoint := os.Getenv("AWS_ENDPOINT")
	region := os.Getenv("AWS_REGION")
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)

	if err != nil {
		panic(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		}
	})

	return &S3Client{
		Client: client,
		Bucket: bucket,
	}

}

// upload file to s3
func (s *S3Client) UploadFile(ctx context.Context, bucket string, key string, file []byte) {
	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(file),
	})

	if err != nil {
		panic(err)
	}
}

// download s3 file
func (s *S3Client) DownloadFile(ctx context.Context, bucket string, key string) []byte {
	file, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		panic(err)
	}

	bytes, err := io.ReadAll(file.Body)
	if err != nil {
		panic(err)
	}

	defer file.Body.Close()
	return bytes
}
