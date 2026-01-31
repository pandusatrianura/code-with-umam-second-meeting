package service

import (
	"github.com/pandusatrianura/code-with-umam-second-meeting/internal/health/entity"
	"github.com/pandusatrianura/code-with-umam-second-meeting/internal/health/repository"
)

type healthService struct {
	healthRepository repository.HealthRepository
}

type HealthService interface {
	API() entity.HealthCheck
	DB() (entity.HealthCheck, error)
}

func NewHealthService(healthRepo repository.HealthRepository) HealthService {
	return &healthService{healthRepository: healthRepo}
}

func (h *healthService) API() entity.HealthCheck {

	return entity.HealthCheck{
		Name:      "Connection to Kasir API",
		IsHealthy: true,
	}
}

func (h *healthService) DB() (entity.HealthCheck, error) {
	err := h.healthRepository.DB()
	if err != nil {
		return entity.HealthCheck{}, err
	}

	return entity.HealthCheck{
		Name:      "Connection to Kasir Database",
		IsHealthy: true,
	}, nil
}
