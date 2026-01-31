package service

import (
	"errors"
	"testing"

	"github.com/pandusatrianura/code-with-umam-second-meeting/internal/health/entity"
)

type stubHealthRepository struct {
	err error
}

func (s stubHealthRepository) DB() error {
	return s.err
}

func TestHealthServiceAPI(t *testing.T) {
	tests := []struct {
		name string
		want entity.HealthCheck
	}{
		{
			name: "ok",
			want: entity.HealthCheck{
				Name:      "Connection to Kasir API",
				IsHealthy: true,
			},
		},
	}

	svc := &healthService{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.API()
			if got != tt.want {
				t.Fatalf("API() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestHealthServiceDB(t *testing.T) {
	errBoom := errors.New("boom")

	tests := []struct {
		name    string
		repoErr error
		want    entity.HealthCheck
		wantErr error
	}{
		{
			name: "ok",
			want: entity.HealthCheck{
				Name:      "Connection to Kasir Database",
				IsHealthy: true,
			},
		},
		{
			name:    "repo err",
			repoErr: errBoom,
			wantErr: errBoom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &healthService{healthRepository: stubHealthRepository{err: tt.repoErr}}
			got, err := svc.DB()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("DB() error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("DB() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestNewHealthService(t *testing.T) {
	tests := []struct {
		name string
		repo stubHealthRepository
	}{
		{name: "ok"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewHealthService(tt.repo)
			if svc == nil {
				t.Fatal("NewHealthService() = nil")
			}
		})
	}
}
