// Server represents an HTTP server with an address for listening to incoming requests.
package api

import (
	"fmt"
	"log"
	"net/http"

	route "github.com/pandusatrianura/code-with-umam-second-meeting/api/router"
	categoriesHandler "github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/delivery/http"
	categoryRepository "github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/repository"
	categoriesService "github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/service"
	productHandler "github.com/pandusatrianura/code-with-umam-second-meeting/internal/products/delivery/http"
	productRepository "github.com/pandusatrianura/code-with-umam-second-meeting/internal/products/repository"
	productService "github.com/pandusatrianura/code-with-umam-second-meeting/internal/products/service"
	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/database"
	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/scalar"
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
	categoriesService := categoriesService.NewCategoryService(categoriesRepo)
	categoriesHandler := categoriesHandler.NewCategoryHandler(categoriesService)

	productsRepo := productRepository.NewProductRepository(s.db)
	productsService := productService.NewProductService(productsRepo)
	productsHandler := productHandler.NewProductHandler(productsService)

	r := route.NewRouter(categoriesHandler, productsHandler)
	routes := r.RegisterRoutes()
	router := http.NewServeMux()
	router.Handle("/kasir/api/", http.StripPrefix("/kasir/api", routes))
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
			SpecURL: "./docs/swagger.json",
			CustomOptions: scalar.CustomOptions{
				PageTitle: "Test Kasir API",
			},
			DarkMode: true,
		})

		if err != nil {
			fmt.Printf("%v", err)
		}

		_, err = fmt.Fprintln(w, htmlContent)
		if err != nil {
			return
		}
	})

	log.Println("Starting server on port", s.addr)
	return http.ListenAndServe(s.addr, router)
}
