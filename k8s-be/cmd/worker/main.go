package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/amitp07/CloudCrush/k8s-be/internal/broker"
	"github.com/amitp07/CloudCrush/k8s-be/internal/config"
	"github.com/amitp07/CloudCrush/k8s-be/internal/dto"
	"github.com/amitp07/CloudCrush/k8s-be/internal/processor"
	"github.com/amitp07/CloudCrush/k8s-be/internal/storage"
	"github.com/amitp07/CloudCrush/k8s-be/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found loading env vars from system environment")
	}

	cfg := config.Load()

	//db connection
	pg, err := pgxpool.New(context.Background(), cfg.PostgresURL)
	if err != nil {
		panic(err)
	}
	if err = pg.Ping(context.Background()); err != nil {
		panic(err)
	}

	store := store.NewStore(pg)

	s3Client := storage.NewS3Client(context.Background(), &cfg)

	// nats connection
	nc, err := broker.ConnectNats(cfg.NatsURL)
	if err != nil {
		panic(err)
	}

	// get nats broker
	nb, err := broker.NewNatsBroker(nc)
	if err != nil {
		panic(err)
	}

	// fetch all subscribers
	sub, err := nb.Js.PullSubscribe("IMAGE.created", "worker-group", nats.PullMaxWaiting(128))
	if err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msgs, err := sub.Fetch(1, nats.Context(ctx))
			if errors.Is(err, nats.ErrTimeout) || errors.Is(err, context.DeadlineExceeded) {
				continue
			}
			if err != nil {
				log.Printf("Fetch error %v\n", err.Error())
				time.Sleep(1 * time.Second)
				continue
			}

			for _, msg := range msgs {
				var job dto.ImageJob
				if err := json.Unmarshal(msg.Data, &job); err != nil {
					fmt.Printf("Rejected bad message: %v, Raw data %s\n", err, string(msg.Data))
					msg.Term()
					continue
				}

				jobId := job.Id
				fmt.Printf("Processing Job %s\n", jobId)

				file := s3Client.DownloadFile(context.Background(), job.Key)

				outputPath := fmt.Sprintf("compressed%s", strings.TrimPrefix(job.Key, "raw"))

				imageBytes, err := processor.CompressImage(file)
				if err != nil {
					msg.Ack()
					panic(err)
				}

				s3Client.UploadFile(context.Background(), outputPath, imageBytes)

				store.UpdateJobStatus("complete", outputPath, jobId)

				msg.Ack()

				fmt.Printf("Processiong Completed for Job %s\n", jobId)
			}
		}
	}
}
