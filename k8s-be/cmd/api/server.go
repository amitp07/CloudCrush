package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/amitp07/CloudCrush/k8s-be/internal/broker"
	"github.com/amitp07/CloudCrush/k8s-be/internal/config"
	"github.com/amitp07/CloudCrush/k8s-be/internal/handlers"
	"github.com/amitp07/CloudCrush/k8s-be/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Server struct {
	imageHandler func(http.ResponseWriter, *http.Request)
}

func NewServer() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found loading env vars from system environment")
	}

	ctx := context.Background()

	cfg := config.Load()

	log.Println("url ==>", cfg.NatsURL, cfg.PostgresURL)

	//fetch nats connection
	nc, err := broker.ConnectNats(cfg.NatsURL)
	if err != nil {
		log.Fatalln("Critical: could not connect to nats", err)
	}
	err = nc.Publish("abc", []byte("some test data"))
	if err != nil {
		panic(err)
	}

	defer nc.Close()

	_, err = nc.JetStream()
	if err != nil {
		log.Fatalln("Critical: could not initialize Nats Jetstream", err)
	}

	db, err := pgxpool.New(ctx, cfg.PostgresURL)
	mux := http.NewServeMux()

	store := store.NewStore(db)
	handler := handlers.PGStore{Store: store}

	mux.HandleFunc("POST /create-image", handler.CreateImage)

	// app := &Application{
	// 	db: db,
	// }

	if err != nil {
		panic(err)
	}

	err = db.Ping(ctx)
	if err != nil {
		panic(err)
	}

	// mux.HandleFunc("GET /user", app.createUser)

	srv := http.Server{
		Addr:    ":3000",
		Handler: mux,
	}

	fmt.Println("Server is running on port 3000")
	panic(srv.ListenAndServe())

}
