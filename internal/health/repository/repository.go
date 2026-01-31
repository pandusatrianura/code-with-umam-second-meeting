package repository

import (
	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/database"
)

type healthRepository struct {
	db *database.DB
}

type HealthRepository interface {
	DB() error
}

func NewHealthRepository(db *database.DB) HealthRepository {
	return &healthRepository{db: db}
}

func (h *healthRepository) DB() error {
	err := h.db.DB.Ping()
	if err != nil {
		return err
	}

	return nil
}
