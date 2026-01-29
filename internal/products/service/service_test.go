package service

import (
	"errors"
	"testing"

	"github.com/pandusatrianura/code-with-umam-second-meeting/internal/products/entity"
)

type mockProductRepository struct {
	createProductFn  func(product *entity.Product) error
	updateProductFn  func(id int64, product *entity.Product) error
	deleteProductFn  func(id int64) error
	getProductByIDFn func(id int64) (*entity.ResponseProductWithCategories, error)
	getAllProductsFn func() ([]entity.ResponseProductWithCategories, error)
	getCategoryByIDFn func(id int64) (*entity.Category, error)

	createProductArg *entity.Product
	updateProductArg *entity.Product
	updateProductID  int64
	deleteProductID  int64
	getCategoryIDArg int64
	getProductIDArg  int64
}

func (m *mockProductRepository) CreateProduct(product *entity.Product) error {
	m.createProductArg = product
	if m.createProductFn == nil {
		return nil
	}
	return m.createProductFn(product)
}

func (m *mockProductRepository) UpdateProduct(id int64, product *entity.Product) error {
	m.updateProductID = id
	m.updateProductArg = product
	if m.updateProductFn == nil {
		return nil
	}
	return m.updateProductFn(id, product)
}

func (m *mockProductRepository) DeleteProduct(id int64) error {
	m.deleteProductID = id
	if m.deleteProductFn == nil {
		return nil
	}
	return m.deleteProductFn(id)
}

func (m *mockProductRepository) GetProductByID(id int64) (*entity.ResponseProductWithCategories, error) {
	m.getProductIDArg = id
	if m.getProductByIDFn == nil {
		return nil, nil
	}
	return m.getProductByIDFn(id)
}

func (m *mockProductRepository) GetAllProducts() ([]entity.ResponseProductWithCategories, error) {
	if m.getAllProductsFn == nil {
		return nil, nil
	}
	return m.getAllProductsFn()
}

func (m *mockProductRepository) GetCategoryByID(id int64) (*entity.Category, error) {
	m.getCategoryIDArg = id
	if m.getCategoryByIDFn == nil {
		return nil, nil
	}
	return m.getCategoryByIDFn(id)
}

func TestNewProductService(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "ok"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProductRepository{}
			service := NewProductService(repo)
			if service == nil {
				t.Fatal("expected service")
			}
			if service.productRepository != repo {
				t.Fatal("repository not set")
			}
		})
	}
}

func TestProductService_API(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "ok"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &productService{productRepository: &mockProductRepository{}}
			got := svc.API()
			if got.Name != "Products API" || got.IsHealthy != true {
				t.Fatalf("unexpected healthcheck: %+v", got)
			}
		})
	}
}

func TestProductService_CreateProduct(t *testing.T) {
	tests := []struct {
		name        string
		req         *entity.RequestProduct
		setupMock   func(m *mockProductRepository)
		wantErr     string
		wantProduct *entity.Product
		wantCatID   int64
	}{
		{
			name: "category-miss",
			req:  &entity.RequestProduct{Name: "n", Price: 10, Stock: 1, CategoryID: 2},
			setupMock: func(m *mockProductRepository) {
				m.getCategoryByIDFn = func(id int64) (*entity.Category, error) {
					return nil, errors.New("nope")
				}
			},
			wantErr: "category not found",
		},
		{
			name: "create-err",
			req:  &entity.RequestProduct{Name: "n", Price: 10, Stock: 1, CategoryID: 2},
			setupMock: func(m *mockProductRepository) {
				m.getCategoryByIDFn = func(id int64) (*entity.Category, error) {
					return &entity.Category{ID: int(id)}, nil
				}
				m.createProductFn = func(product *entity.Product) error {
					return errors.New("db down")
				}
			},
			wantErr: "db down",
		},
		{
			name: "ok",
			req:  &entity.RequestProduct{Name: "n", Price: 10, Stock: 1, CategoryID: 2},
			setupMock: func(m *mockProductRepository) {
				m.getCategoryByIDFn = func(id int64) (*entity.Category, error) {
					return &entity.Category{ID: int(id)}, nil
				}
			},
			wantProduct: &entity.Product{Name: "n", Price: 10, Stock: 1, CategoryID: 2},
			wantCatID:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProductRepository{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}
			svc := &productService{productRepository: repo}
			err := svc.CreateProduct(tt.req)

			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantProduct != nil {
				got := repo.createProductArg
				if got == nil {
					t.Fatal("expected product to be passed")
				}
				if *got != *tt.wantProduct {
					t.Fatalf("unexpected product: %+v", *got)
				}
				if repo.getCategoryIDArg != tt.wantCatID {
					t.Fatalf("unexpected category id: %d", repo.getCategoryIDArg)
				}
			}
		})
	}
}

func TestProductService_UpdateProduct(t *testing.T) {
	tests := []struct {
		name        string
		id          int64
		req         *entity.RequestProduct
		setupMock   func(m *mockProductRepository)
		wantErr     string
		wantProduct *entity.Product
		wantCatID   int64
		wantID      int64
	}{
		{
			name: "product-miss",
			id:   10,
			req:  &entity.RequestProduct{Name: "n", Price: 10, Stock: 1, CategoryID: 2},
			setupMock: func(m *mockProductRepository) {
				m.getProductByIDFn = func(id int64) (*entity.ResponseProductWithCategories, error) {
					return nil, errors.New("no product")
				}
			},
			wantErr: "product not found",
		},
		{
			name: "category-miss",
			id:   10,
			req:  &entity.RequestProduct{Name: "n", Price: 10, Stock: 1, CategoryID: 2},
			setupMock: func(m *mockProductRepository) {
				m.getProductByIDFn = func(id int64) (*entity.ResponseProductWithCategories, error) {
					return &entity.ResponseProductWithCategories{ID: int(id)}, nil
				}
				m.getCategoryByIDFn = func(id int64) (*entity.Category, error) {
					return nil, errors.New("no category")
				}
			},
			wantErr: "category not found",
		},
		{
			name: "update-err",
			id:   10,
			req:  &entity.RequestProduct{Name: "n", Price: 10, Stock: 1, CategoryID: 2},
			setupMock: func(m *mockProductRepository) {
				m.getProductByIDFn = func(id int64) (*entity.ResponseProductWithCategories, error) {
					return &entity.ResponseProductWithCategories{ID: int(id)}, nil
				}
				m.getCategoryByIDFn = func(id int64) (*entity.Category, error) {
					return &entity.Category{ID: int(id)}, nil
				}
				m.updateProductFn = func(id int64, product *entity.Product) error {
					return errors.New("update fail")
				}
			},
			wantErr: "update fail",
		},
		{
			name: "ok",
			id:   10,
			req:  &entity.RequestProduct{Name: "n", Price: 10, Stock: 1, CategoryID: 2},
			setupMock: func(m *mockProductRepository) {
				m.getProductByIDFn = func(id int64) (*entity.ResponseProductWithCategories, error) {
					return &entity.ResponseProductWithCategories{ID: int(id)}, nil
				}
				m.getCategoryByIDFn = func(id int64) (*entity.Category, error) {
					return &entity.Category{ID: int(id)}, nil
				}
			},
			wantProduct: &entity.Product{Name: "n", Price: 10, Stock: 1, CategoryID: 2},
			wantCatID:   2,
			wantID:      10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProductRepository{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}
			svc := &productService{productRepository: repo}
			err := svc.UpdateProduct(tt.id, tt.req)

			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantProduct != nil {
				got := repo.updateProductArg
				if got == nil {
					t.Fatal("expected product to be passed")
				}
				if *got != *tt.wantProduct {
					t.Fatalf("unexpected product: %+v", *got)
				}
				if repo.getCategoryIDArg != tt.wantCatID {
					t.Fatalf("unexpected category id: %d", repo.getCategoryIDArg)
				}
				if repo.updateProductID != tt.wantID {
					t.Fatalf("unexpected update id: %d", repo.updateProductID)
				}
			}
		})
	}
}

func TestProductService_DeleteProduct(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		setupMock func(m *mockProductRepository)
		wantErr   string
		wantID    int64
	}{
		{
			name: "product-miss",
			id:   10,
			setupMock: func(m *mockProductRepository) {
				m.getProductByIDFn = func(id int64) (*entity.ResponseProductWithCategories, error) {
					return nil, errors.New("no product")
				}
			},
			wantErr: "product not found",
		},
		{
			name: "delete-err",
			id:   10,
			setupMock: func(m *mockProductRepository) {
				m.getProductByIDFn = func(id int64) (*entity.ResponseProductWithCategories, error) {
					return &entity.ResponseProductWithCategories{ID: int(id)}, nil
				}
				m.deleteProductFn = func(id int64) error {
					return errors.New("delete fail")
				}
			},
			wantErr: "delete fail",
		},
		{
			name: "ok",
			id:   10,
			setupMock: func(m *mockProductRepository) {
				m.getProductByIDFn = func(id int64) (*entity.ResponseProductWithCategories, error) {
					return &entity.ResponseProductWithCategories{ID: int(id)}, nil
				}
			},
			wantID: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProductRepository{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}
			svc := &productService{productRepository: repo}
			err := svc.DeleteProduct(tt.id)

			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if repo.deleteProductID != tt.wantID {
				t.Fatalf("unexpected delete id: %d", repo.deleteProductID)
			}
		})
	}
}

func TestProductService_GetProductByID(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		setupMock func(m *mockProductRepository)
		want      *entity.ResponseProductWithCategories
		wantErr   string
	}{
		{
			name: "ok",
			id:   10,
			setupMock: func(m *mockProductRepository) {
				m.getProductByIDFn = func(id int64) (*entity.ResponseProductWithCategories, error) {
					return &entity.ResponseProductWithCategories{ID: int(id)}, nil
				}
			},
			want: &entity.ResponseProductWithCategories{ID: 10},
		},
		{
			name: "err",
			id:   10,
			setupMock: func(m *mockProductRepository) {
				m.getProductByIDFn = func(id int64) (*entity.ResponseProductWithCategories, error) {
					return nil, errors.New("boom")
				}
			},
			wantErr: "boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProductRepository{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}
			svc := &productService{productRepository: repo}
			got, err := svc.GetProductByID(tt.id)

			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if got != nil {
					t.Fatalf("expected nil result, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil || *got != *tt.want {
				t.Fatalf("unexpected result: %+v", got)
			}
		})
	}
}

func TestProductService_GetAllProducts(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(m *mockProductRepository)
		want      []entity.ResponseProductWithCategories
		wantErr   string
	}{
		{
			name: "ok",
			setupMock: func(m *mockProductRepository) {
				m.getAllProductsFn = func() ([]entity.ResponseProductWithCategories, error) {
					return []entity.ResponseProductWithCategories{{ID: 1}, {ID: 2}}, nil
				}
			},
			want: []entity.ResponseProductWithCategories{{ID: 1}, {ID: 2}},
		},
		{
			name: "err",
			setupMock: func(m *mockProductRepository) {
				m.getAllProductsFn = func() ([]entity.ResponseProductWithCategories, error) {
					return nil, errors.New("boom")
				}
			},
			wantErr: "boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProductRepository{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}
			svc := &productService{productRepository: repo}
			got, err := svc.GetAllProducts()

			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if got != nil {
					t.Fatalf("expected nil result, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("unexpected result length: %d", len(got))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("unexpected result: %+v", got)
				}
			}
		})
	}
}
