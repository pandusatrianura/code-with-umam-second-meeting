package router

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	categoriesHandler "github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/delivery/http"
	categoriesEntity "github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/entity"
	productsHandler "github.com/pandusatrianura/code-with-umam-second-meeting/internal/products/delivery/http"
	productsEntity "github.com/pandusatrianura/code-with-umam-second-meeting/internal/products/entity"
)

type stubProductService struct{}

func (stubProductService) CreateProduct(*productsEntity.RequestProduct) error {
	return errors.New("fail")
}

func (stubProductService) UpdateProduct(int64, *productsEntity.RequestProduct) error {
	return errors.New("fail")
}

func (stubProductService) DeleteProduct(int64) error {
	return errors.New("fail")
}

func (stubProductService) GetProductByID(int64) (*productsEntity.ResponseProductWithCategories, error) {
	return nil, errors.New("fail")
}

func (stubProductService) GetAllProducts() ([]productsEntity.ResponseProductWithCategories, error) {
	return nil, errors.New("fail")
}

func (stubProductService) API() productsEntity.HealthCheck {
	return productsEntity.HealthCheck{Name: "Products API", IsHealthy: false}
}

type stubCategoryService struct{}

func (stubCategoryService) CreateCategory(*categoriesEntity.RequestCategory) error {
	return nil
}

func (stubCategoryService) UpdateCategory(int64, *categoriesEntity.RequestCategory) error {
	return nil
}

func (stubCategoryService) DeleteCategory(int64) error {
	return nil
}

func (stubCategoryService) GetCategoryByID(int64) (*categoriesEntity.ResponseCategory, error) {
	return &categoriesEntity.ResponseCategory{}, nil
}

func (stubCategoryService) GetAllCategories() ([]categoriesEntity.ResponseCategory, error) {
	return []categoriesEntity.ResponseCategory{}, nil
}

func (stubCategoryService) API() categoriesEntity.HealthCheck {
	return categoriesEntity.HealthCheck{Name: "Categories API", IsHealthy: true}
}

func TestNewRouter(t *testing.T) {
	categoryHandler := categoriesHandler.NewCategoryHandler(stubCategoryService{})
	productHandler := productsHandler.NewProductHandler(stubProductService{})

	tests := []struct {
		name       string
		categories *categoriesHandler.CategoryHandler
		products   *productsHandler.ProductHandler
	}{
		{name: "nil", categories: nil, products: nil},
		{name: "set", categories: categoryHandler, products: productHandler},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter(tt.categories, tt.products)
			if r == nil {
				t.Fatalf("expected router")
			}
			if r.categories != tt.categories {
				t.Fatalf("categories mismatch")
			}
			if r.products != tt.products {
				t.Fatalf("products mismatch")
			}
		})
	}
}

func TestRegisterRoutes(t *testing.T) {
	categoryHandler := categoriesHandler.NewCategoryHandler(stubCategoryService{})
	productHandler := productsHandler.NewProductHandler(stubProductService{})
	r := NewRouter(categoryHandler, productHandler)
	mux := r.RegisterRoutes()
	if mux == nil {
		t.Fatalf("expected mux")
	}

	productBody := `{"name":"tea","price":100,"stock":1,"category_id":1}`
	categoryBody := `{"name":"drink","description":"cold"}`

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{name: "pcreate", method: http.MethodPost, path: "/products", body: productBody, wantStatus: http.StatusInternalServerError},
		{name: "pupdate", method: http.MethodPut, path: "/products/1", body: productBody, wantStatus: http.StatusInternalServerError},
		{name: "ccreate", method: http.MethodPost, path: "/categories", body: categoryBody, wantStatus: http.StatusCreated},
		{name: "cupdate", method: http.MethodPut, path: "/categories/1", body: categoryBody, wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body *bytes.Buffer
			if tt.body != "" {
				body = bytes.NewBufferString(tt.body)
			}
			req := httptest.NewRequest(tt.method, tt.path, body)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
