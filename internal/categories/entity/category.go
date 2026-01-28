package entity

import "time"

type Category struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   string
	UpdatedAt   string
}

type RequestCategory struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ResponseCategory struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at", omitempty`
	UpdatedAt   time.Time `json:"updated_at", omitempty`
}

type HealthCheck struct {
	Name      string `json:"name"`
	IsHealthy bool   `json:"is_healthy"`
}
