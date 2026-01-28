package entity

import "time"

type Product struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Price      int    `json:"price"`
	Stock      int    `json:"stock"`
	CategoryID int    `json:"category_id"`
	CreatedAt  string `json:"created_at", omitempty`
	UpdatedAt  string `json:"updated_at", omitempty`
}

type RequestProduct struct {
	Name       string `json:"name"`
	Price      int    `json:"price"`
	Stock      int    `json:"stock"`
	CategoryID int    `json:"category_id"`
}

type HealthCheck struct {
	Name      string `json:"name"`
	IsHealthy bool   `json:"is_healthy"`
}

type ProductWithCategories struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Price        int    `json:"price"`
	Stock        int    `json:"stock"`
	CategoryID   int    `json:"category_id,omitempty"`
	CategoryName string `json:"category_name"`
	CreatedAt    string `json:"created_at", omitempty`
	UpdatedAt    string `json:"updated_at", omitempty`
}

type ResponseProductWithCategories struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Price        int       `json:"price"`
	Stock        int       `json:"stock"`
	CategoryID   int       `json:"category_id,omitempty"`
	CategoryName string    `json:"category_name"`
	CreatedAt    time.Time `json:"created_at", omitempty`
	UpdatedAt    time.Time `json:"updated_at", omitempty`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
