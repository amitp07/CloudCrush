package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{
		db: db,
	}
}

type ImageJobsData struct {
	Filename       string
	Status         string
	OriginalPath   string
	CompressedPath string
}

func (pg *Store) CreateImage(input ImageJobsData) string {

	query := `INSERT INTO image_jobs(id, filename, status, original_path, compressed_path) VALUES
	($1, $2, $3, $4, $5) returning id;
	`
	args := []interface{}{uuid.New(), input.Filename, input.Status, input.OriginalPath, input.CompressedPath}

	var id string
	err := pg.db.QueryRow(context.TODO(), query, args...).Scan(&id)

	if err != nil {
		panic(err)
	}

	return id
}

func (pg *Store) GetJobById(id string) ImageJobsData {
	var imageJobsData ImageJobsData

	query := `select filename, status, original_path, compressed_path from image_jobs where id = $1`

	err := pg.db.QueryRow(context.Background(), query, id).Scan(&imageJobsData.Filename, &imageJobsData.Status, &imageJobsData.OriginalPath, &imageJobsData.CompressedPath)

	if err != nil {
		panic(err)
	}

	return imageJobsData
}

func (pg *Store) UpdateJobStatus(status, compressedPath, id string) {
	query := `update image_jobs set status = $1, compressed_path = $2, updated_at = NOW() where id = $3`

	_, err := pg.db.Exec(context.Background(), query, status, compressedPath, id)

	if err != nil {
		panic(err)
	}
}
