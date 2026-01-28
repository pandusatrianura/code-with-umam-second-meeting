// Server represents an HTTP server with an address for listening to incoming requests.
package api

import (
	"log"
	"net/http"

	route "github.com/pandusatrianura/code-with-umam-second-meeting/api/router"
	categoryHandler "github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/delivery/http"
	categoryRepository "github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/repository"
	categoryService "github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/service"
	productHandler "github.com/pandusatrianura/code-with-umam-second-meeting/internal/products/delivery/http"
	productRepository "github.com/pandusatrianura/code-with-umam-second-meeting/internal/products/repository"
	productService "github.com/pandusatrianura/code-with-umam-second-meeting/internal/products/service"
	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/database"
)

type Server struct {
	addr string
	db   *database.DB
}

// NewAPIServer initializes and returns a new Server instance configured to listen on the specified address.
func NewAPIServer(addr string, db *database.DB) *Server {
	return &Server{
		addr: addr,
		db:   db,
	}
}

// Run starts the server, initializes dependencies, registers routes, and listens for incoming HTTP requests.
func (s *Server) Run() error {

	categoriesRepo := categoryRepository.NewCategoryRepository(s.db)
	categoriesService := categoryService.NewCategoryService(categoriesRepo)
	categoriesHandler := categoryHandler.NewCategoryHandler(categoriesService)

	productsRepo := productRepository.NewProductRepository(s.db)
	productsService := productService.NewProductService(productsRepo)
	productsHandler := productHandler.NewProductHandler(productsService)

	r := route.NewRouter(categoriesHandler, productsHandler)
	routes := r.RegisterRoutes()
	router := http.NewServeMux()
	router.Handle("/api/", http.StripPrefix("/api", routes))
	log.Println("Starting server on port", s.addr)
	return http.ListenAndServe(s.addr, router)
}
