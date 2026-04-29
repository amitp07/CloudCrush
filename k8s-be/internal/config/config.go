package config

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type Config struct {
	NatsURL     string
	PostgresURL string
}

const (
	NATS_URL     = "/cloudcrush/sandbox/nats/url"
	POSTGRES_URL = "/cloudcrush/sandbox/postgres/url"
)

func Load() Config {
	natsUrl := os.Getenv("NATS_DSN")
	pgUrl := os.Getenv("POSTGRES_DSN")

	if natsUrl == "" && pgUrl == "" {

		ctx := context.Background()

		awsConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion("ap-south-1"))

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
	}

	return Config{
		NatsURL:     natsUrl,
		PostgresURL: pgUrl,
	}
}
