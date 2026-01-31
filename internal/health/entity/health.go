package entity

type HealthCheck struct {
	Name      string `json:"name"`
	IsHealthy bool   `json:"is_healthy"`
}
