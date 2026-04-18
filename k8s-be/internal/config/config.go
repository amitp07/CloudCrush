package config

import "os"

type Config struct {
	NatsURL     string
	PostgresURL string
}

func Load() Config {
	natsUrl := os.Getenv("NATS_DSN")

	pgUrl := os.Getenv("POSTGRES_DSN")

	return Config{
		NatsURL:     natsUrl,
		PostgresURL: pgUrl,
	}
}
