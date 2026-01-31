package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	categoriesHandler "github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/delivery/http"
	categoriesEntity "github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/entity"
	healthHandler "github.com/pandusatrianura/code-with-umam-second-meeting/internal/health/delivery/http"
	healthEntity "github.com/pandusatrianura/code-with-umam-second-meeting/internal/health/entity"
	productsHandler "github.com/pandusatrianura/code-with-umam-second-meeting/internal/products/delivery/http"
	productsEntity "github.com/pandusatrianura/code-with-umam-second-meeting/internal/products/entity"
)

type fakeCategoryService struct{}

type fakeProductService struct{}

type fakeHealthService struct{}

func (fakeCategoryService) CreateCategory(*categoriesEntity.RequestCategory) error {
	return nil
}

func (fakeCategoryService) UpdateCategory(int64, *categoriesEntity.RequestCategory) error {
	return nil
}

func (fakeCategoryService) DeleteCategory(int64) error {
	return nil
}

func (fakeCategoryService) GetCategoryByID(int64) (*categoriesEntity.ResponseCategory, error) {
	return &categoriesEntity.ResponseCategory{}, nil
}

func (fakeCategoryService) GetAllCategories() ([]categoriesEntity.ResponseCategory, error) {
	return []categoriesEntity.ResponseCategory{}, nil
}

func (fakeCategoryService) API() categoriesEntity.HealthCheck {
	return categoriesEntity.HealthCheck{}
}

func (fakeProductService) CreateProduct(*productsEntity.RequestProduct) error {
	return nil
}

func (fakeProductService) UpdateProduct(int64, *productsEntity.RequestProduct) error {
	return nil
}

func (fakeProductService) DeleteProduct(int64) error {
	return nil
}

func (fakeProductService) GetProductByID(int64) (*productsEntity.ResponseProductWithCategories, error) {
	return &productsEntity.ResponseProductWithCategories{}, nil
}

func (fakeProductService) GetAllProducts() ([]productsEntity.ResponseProductWithCategories, error) {
	return []productsEntity.ResponseProductWithCategories{}, nil
}

func (fakeProductService) API() productsEntity.HealthCheck {
	return productsEntity.HealthCheck{}
}

func (fakeHealthService) API() healthEntity.HealthCheck {
	return healthEntity.HealthCheck{}
}

func (fakeHealthService) DB() (healthEntity.HealthCheck, error) {
	return healthEntity.HealthCheck{}, nil
}

func TestNewRouter(t *testing.T) {
	categories := categoriesHandler.NewCategoryHandler(fakeCategoryService{})
	products := productsHandler.NewProductHandler(fakeProductService{})
	health := healthHandler.NewHealthHandler(fakeHealthService{})

	got := NewRouter(categories, products, health)

	if got.categories != categories {
		t.Fatalf("categories handler mismatch")
	}
	if got.products != products {
		t.Fatalf("products handler mismatch")
	}
	if got.health != health {
		t.Fatalf("health handler mismatch")
	}
}

func TestRegisterRoutes(t *testing.T) {
	r := NewRouter(
		categoriesHandler.NewCategoryHandler(fakeCategoryService{}),
		productsHandler.NewProductHandler(fakeProductService{}),
		healthHandler.NewHealthHandler(fakeHealthService{}),
	)
	mux := r.RegisterRoutes()

	cases := []struct {
		name        string
		method      string
		path        string
		wantPattern string
	}{
		{name: "health-service", method: http.MethodGet, path: "/health/service", wantPattern: "GET /health/service"},
		{name: "health-db", method: http.MethodGet, path: "/health/db", wantPattern: "GET /health/db"},
		{name: "products-health", method: http.MethodGet, path: "/products/health", wantPattern: "GET /products/health"},
		{name: "products-create", method: http.MethodPost, path: "/products", wantPattern: "POST /products"},
		{name: "products-list", method: http.MethodGet, path: "/products", wantPattern: "GET /products"},
		{name: "products-get", method: http.MethodGet, path: "/products/123", wantPattern: "GET /products/{id}"},
		{name: "products-update", method: http.MethodPut, path: "/products/123", wantPattern: "PUT /products/{id}"},
		{name: "products-delete", method: http.MethodDelete, path: "/products/123", wantPattern: "DELETE /products/{id}"},
		{name: "categories-health", method: http.MethodGet, path: "/categories/health", wantPattern: "GET /categories/health"},
		{name: "categories-create", method: http.MethodPost, path: "/categories", wantPattern: "POST /categories"},
		{name: "categories-list", method: http.MethodGet, path: "/categories", wantPattern: "GET /categories"},
		{name: "categories-get", method: http.MethodGet, path: "/categories/123", wantPattern: "GET /categories/{id}"},
		{name: "categories-update", method: http.MethodPut, path: "/categories/123", wantPattern: "PUT /categories/{id}"},
		{name: "categories-delete", method: http.MethodDelete, path: "/categories/123", wantPattern: "DELETE /categories/{id}"},
		{name: "docs", method: http.MethodGet, path: "/docs", wantPattern: "GET /docs"},
		{name: "method-mismatch", method: http.MethodPost, path: "/health/service", wantPattern: ""},
		{name: "unknown", method: http.MethodGet, path: "/unknown", wantPattern: ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "http://example.com"+tc.path, nil)
			_, gotPattern := mux.Handler(req)
			if gotPattern != tc.wantPattern {
				t.Fatalf("pattern mismatch: got %q want %q", gotPattern, tc.wantPattern)
			}
		})
	}

	t.Run("docs-response", func(t *testing.T) {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("getwd: %v", err)
		}
		tempDir := t.TempDir()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("chdir: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Chdir(cwd)
		})

		docsDir := filepath.Join(tempDir, "docs")
		if err := os.MkdirAll(docsDir, 0o755); err != nil {
			t.Fatalf("mkdir docs: %v", err)
		}
		swaggerPath := filepath.Join(docsDir, "swagger.json")
		if err := os.WriteFile(swaggerPath, []byte(`{"openapi":"3.0.0"}`), 0o644); err != nil {
			t.Fatalf("write swagger: %v", err)
		}

		mux := r.RegisterRoutes()
		req := httptest.NewRequest(http.MethodGet, "http://example.com/docs", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusOK)
		}
		if !strings.Contains(rec.Body.String(), "Test Kasir API") {
			t.Fatalf("missing page title")
		}
	})
}
