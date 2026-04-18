package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/amitp07/CloudCrush/k8s-be/internal/broker"
	"github.com/amitp07/CloudCrush/k8s-be/internal/config"
	"github.com/amitp07/CloudCrush/k8s-be/internal/processor"
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

	pgStore := store.NewStore(pg)

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
			fmt.Printf("msg %v\n", msgs)
			if errors.Is(err, nats.ErrTimeout) || errors.Is(err, context.DeadlineExceeded) {
				continue
			}
			if err != nil {
				log.Printf("Fetch error %v\n", err.Error())
				time.Sleep(1 * time.Second)
				continue
			}

			for _, msg := range msgs {
				jobId := string(msg.Data)
				fmt.Printf("Processing Job %s\n", jobId)

				job := pgStore.GetJobById(jobId)

				ext := filepath.Ext(job.OriginalPath)
				outputPath := fmt.Sprintf("%s-compressed%s", strings.TrimSuffix(job.OriginalPath, ext), ext)
				if err := processor.CompressImage(job.OriginalPath, outputPath); err != nil {
					msg.Nak()
					panic(err)
				}

				pgStore.UpdateJobStatus("complete", outputPath, jobId)

				msg.Ack()

				fmt.Printf("Processiong Completed for Job %s\n", jobId)
			}
		}
	}
}
