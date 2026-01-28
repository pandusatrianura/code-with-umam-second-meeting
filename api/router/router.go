package router

import (
	"net/http"

	categoriesHandler "github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/delivery/http"
	productsHandler "github.com/pandusatrianura/code-with-umam-second-meeting/internal/products/delivery/http"
)

type Router struct {
	categories *categoriesHandler.CategoryHandler
	products   *productsHandler.ProductHandler
}

func NewRouter(categoriesHandler *categoriesHandler.CategoryHandler, productHandler *productsHandler.ProductHandler) *Router {
	return &Router{
		categories: categoriesHandler,
		products:   productHandler,
	}
}

func (h *Router) RegisterRoutes() *http.ServeMux {
	r := http.NewServeMux()
	r.HandleFunc("GET /products/health", h.products.API)
	r.HandleFunc("POST /products", h.products.CreateProduct)
	r.HandleFunc("GET /products", h.products.GetAllProducts)
	r.HandleFunc("GET /products/{id}", h.products.GetProductByID)
	r.HandleFunc("PUT /products/{id}", h.products.UpdateProduct)
	r.HandleFunc("DELETE /products/{id}", h.products.DeleteProduct)
	r.HandleFunc("GET /categories/health", h.categories.API)
	r.HandleFunc("POST /categories", h.categories.CreateCategory)
	r.HandleFunc("GET /categories", h.categories.GetAllCategories)
	r.HandleFunc("GET /categories/{id}", h.categories.GetCategoryByID)
	r.HandleFunc("PUT /categories/{id}", h.categories.UpdateCategory)
	r.HandleFunc("DELETE /categories/{id}", h.categories.DeleteCategory)
	return r
}
