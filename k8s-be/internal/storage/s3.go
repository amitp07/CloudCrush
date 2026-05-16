package storage

import (
	"bytes"
	"context"
	"io"

	"github.com/amitp07/CloudCrush/k8s-be/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
	Client *s3.Client
	Bucket string
}

// create new S3 client
func NewS3Client(ctx context.Context, appCfg *config.Config) *S3Client {

	cfg, err := awsCfg.LoadDefaultConfig(
		ctx,
		awsCfg.WithRegion(appCfg.AwsRegion),
	)

	if err != nil {
		panic(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if appCfg.AwsS3Endpoint != "" {
			o.BaseEndpoint = aws.String(appCfg.AwsS3Endpoint)
			o.UsePathStyle = true
		}
	})

	return &S3Client{
		Client: client,
		Bucket: appCfg.AwsS3Bucket,
	}

}

// upload file to s3
func (s *S3Client) UploadFile(ctx context.Context, key string, file []byte) {
	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(file),
	})

	if err != nil {
		panic(err)
	}
}

// download s3 file
func (s *S3Client) DownloadFile(ctx context.Context, key string) []byte {
	file, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
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
