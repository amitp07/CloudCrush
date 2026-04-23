package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/amitp07/CloudCrush/k8s-be/internal/broker"
	"github.com/amitp07/CloudCrush/k8s-be/internal/config"
	"github.com/amitp07/CloudCrush/k8s-be/internal/handlers"
	"github.com/amitp07/CloudCrush/k8s-be/internal/routes"
	"github.com/amitp07/CloudCrush/k8s-be/internal/storage"
	"github.com/amitp07/CloudCrush/k8s-be/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Application struct{}

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found loading env vars from system environment")
	}

	ctx := context.Background()
	cfg := config.Load()

	//fetch nats connection
	nc, err := broker.ConnectNats(cfg.NatsURL)
	if err != nil {
		log.Fatalln("Critical: could not connect to nats", err)
	}
	defer nc.Close()

	nb, err := broker.NewNatsBroker(nc)
	if err != nil {
		panic(err)
	}

	// create Nats js stream
	if err = broker.CreateStream(nb.Js); err != nil {
		panic(err)
	}

	_, err = nc.JetStream()
	if err != nil {
		log.Fatalln("Critical: could not initialize Nats Jetstream", err)
	}

	db, err := pgxpool.New(ctx, cfg.PostgresURL)
	err = db.Ping(ctx)
	if err != nil {
		panic(err)
	}
	store := store.NewStore(db)

	storage.NewS3Client(ctx)
	// init image handler
	imageHandler := handlers.ImageHandler{DB: store, Broker: nb}
	//init router
	routesConfig := routes.Config{
		Image: imageHandler,
	}
	router := routes.NewRouter(routesConfig)

	// mux.HandleFunc("GET /user", app.createUser)

	srv := http.Server{
		Addr:    ":3000",
		Handler: router,
	}

	fmt.Println("Server is running on port 3000")
	panic(srv.ListenAndServe())

}
