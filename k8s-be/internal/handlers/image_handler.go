package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/amitp07/CloudCrush/k8s-be/internal/store"
)

type ImageEventPublisher interface {
	PublishImageJob(context.Context, string) error
}

type ImageCreator interface {
	CreateImage(store.ImageJobsData) string
}

type PGStore struct {
	Store  ImageCreator
	Broker ImageEventPublisher
}

func (pgStore PGStore) CreateImage(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Request hit")
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		fmt.Println(err.Error())
		http.Error(w, "File size should not be more than 10MB", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Invlid file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), header.Filename)
	savePath := filepath.Join("./upload", filename)

	dst, err := os.Create(savePath)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		return
	}

	dbInput := store.ImageJobsData{
		Filename:     header.Filename,
		OriginalPath: savePath,
	}

	id := pgStore.Store.CreateImage(dbInput)

	pgStore.Broker.PublishImageJob(r.Context(), string(id))

	w.Write([]byte(fmt.Sprintf("Created id %v\n", id)))
}
