package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/amitp07/CloudCrush/k8s-be/internal/dto"
	"github.com/amitp07/CloudCrush/k8s-be/internal/store"
)

type ImageEventPublisher interface {
	PublishImageJob(context.Context, dto.ImageJob) error
}

type ImageService interface {
	CreateImage(store.ImageJobsData) string
}

type ObjectStorage interface {
	UploadFile(ctx context.Context, key string, file []byte)
}

type ImageHandler struct {
	DB      ImageService
	Broker  ImageEventPublisher
	Storage ObjectStorage
}

func (h ImageHandler) CreateImage(w http.ResponseWriter, r *http.Request) {

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
	key := fmt.Sprintf("raw/%d-%s", time.Now().UnixNano(), header.Filename)

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	// store in s3
	h.Storage.UploadFile(r.Context(), key, fileBytes)

	dbInput := store.ImageJobsData{
		Filename:     header.Filename,
		OriginalPath: key,
	}

	id := h.DB.CreateImage(dbInput)

	h.Broker.PublishImageJob(r.Context(), dto.ImageJob{
		Id:  string(id),
		Key: key,
	})

	w.Write([]byte(fmt.Sprintf("Created id %v\n", id)))
}
