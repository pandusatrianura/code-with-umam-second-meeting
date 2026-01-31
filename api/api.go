// Package api Server represents an HTTP server with an address for listening to incoming requests.
package api

import (
	"fmt"
	"log"
	"net/http"

	route "github.com/pandusatrianura/code-with-umam-second-meeting/api/router"
	categoryHandler "github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/delivery/http"
	categoryRepository "github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/repository"
	categoryService "github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/service"
	healthHandler "github.com/pandusatrianura/code-with-umam-second-meeting/internal/health/delivery/http"
	healthRepository "github.com/pandusatrianura/code-with-umam-second-meeting/internal/health/repository"
	healthService "github.com/pandusatrianura/code-with-umam-second-meeting/internal/health/service"
	productHandler "github.com/pandusatrianura/code-with-umam-second-meeting/internal/products/delivery/http"
	productRepository "github.com/pandusatrianura/code-with-umam-second-meeting/internal/products/repository"
	productService "github.com/pandusatrianura/code-with-umam-second-meeting/internal/products/service"
	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/database"
)

type Server struct {
	addr string
	db   *database.DB
}

// NewAPIServer initializes and returns a new Server instance configured to listen to the specified address.
func NewAPIServer(addr string, db *database.DB) *Server {
	return &Server{
		addr: addr,
		db:   db,
	}
}

// Run starts the server, initializes dependencies, registers routes, and listens for incoming HTTP requests.
func (s *Server) Run() error {

	categoriesRepo := categoryRepository.NewCategoryRepository(s.db)
	categoriesSvc := categoryService.NewCategoryService(categoriesRepo)
	categoriesHandler := categoryHandler.NewCategoryHandler(categoriesSvc)

	productsRepo := productRepository.NewProductRepository(s.db)
	productsSvc := productService.NewProductService(productsRepo)
	productsHandler := productHandler.NewProductHandler(productsSvc)

	healthRepo := healthRepository.NewHealthRepository(s.db)
	healthSvc := healthService.NewHealthService(healthRepo)
	healthHandle := healthHandler.NewHealthHandler(healthSvc)

	r := route.NewRouter(categoriesHandler, productsHandler, healthHandle)
	routes := r.RegisterRoutes()
	router := http.NewServeMux()
	router.Handle("/api/", http.StripPrefix("/api", routes))

	addr := fmt.Sprintf("%s%s", "0.0.0.0", s.addr)
	log.Println("Starting server on", addr)
	return http.ListenAndServe(s.addr, router)
}
