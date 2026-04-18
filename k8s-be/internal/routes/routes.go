package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ImageHandler interface {
	CreateImage(http.ResponseWriter, *http.Request)
}

type Config struct {
	Image ImageHandler
}

func NewRouter(cfg Config) chi.Router {
	router := chi.NewMux()

	router.Route("/api", func(r chi.Router) {
		r.Post("/create-image", cfg.Image.CreateImage)
	})

	return router
}
