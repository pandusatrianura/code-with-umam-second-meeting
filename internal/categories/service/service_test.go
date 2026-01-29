package service

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/entity"
	"github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/repository"
)

type mockCategoryRepository struct {
	createFunc  func(*entity.Category) error
	updateFunc  func(int64, *entity.Category) error
	deleteFunc  func(int64) error
	getByIDFunc func(int64) (*entity.ResponseCategory, error)
	getAllFunc  func() ([]entity.ResponseCategory, error)
}

func (m *mockCategoryRepository) CreateCategory(category *entity.Category) error {
	if m.createFunc == nil {
		return errors.New("not implemented")
	}
	return m.createFunc(category)
}

func (m *mockCategoryRepository) UpdateCategory(id int64, category *entity.Category) error {
	if m.updateFunc == nil {
		return errors.New("not implemented")
	}
	return m.updateFunc(id, category)
}

func (m *mockCategoryRepository) DeleteCategory(id int64) error {
	if m.deleteFunc == nil {
		return errors.New("not implemented")
	}
	return m.deleteFunc(id)
}

func (m *mockCategoryRepository) GetCategoryByID(id int64) (*entity.ResponseCategory, error) {
	if m.getByIDFunc == nil {
		return nil, errors.New("not implemented")
	}
	return m.getByIDFunc(id)
}

func (m *mockCategoryRepository) GetAllCategories() ([]entity.ResponseCategory, error) {
	if m.getAllFunc == nil {
		return nil, errors.New("not implemented")
	}
	return m.getAllFunc()
}

var _ repository.CategoryRepository = (*mockCategoryRepository)(nil)

func TestNewCategoryService(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "ok"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCategoryRepository{}
			svc := NewCategoryService(repo)
			if svc == nil {
				t.Fatal("expected non-nil service")
			}
			if _, ok := svc.(*categoryService); !ok {
				t.Fatalf("expected *categoryService, got %T", svc)
			}
		})
	}
}

func TestCategoryServiceAPI(t *testing.T) {
	tests := []struct {
		name string
		want entity.HealthCheck
	}{
		{
			name: "ok",
			want: entity.HealthCheck{Name: "Categories API", IsHealthy: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &categoryService{}
			got := svc.API()
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("expected %+v, got %+v", tt.want, got)
			}
		})
	}
}

func TestCategoryServiceCreateCategory(t *testing.T) {
	req := &entity.RequestCategory{Name: "Food", Description: "Daily"}
	repoErr := errors.New("repo error")

	tests := []struct {
		name    string
		err     error
		wantErr string
	}{
		{name: "ok"},
		{name: "err", err: repoErr, wantErr: repoErr.Error()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotCategory *entity.Category
			called := false
			repo := &mockCategoryRepository{
				createFunc: func(category *entity.Category) error {
					called = true
					gotCategory = category
					return tt.err
				},
			}
			svc := &categoryService{categoryRepository: repo}
			err := svc.CreateCategory(req)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !called {
				t.Fatal("expected repository CreateCategory to be called")
			}
			if gotCategory == nil {
				t.Fatal("expected category to be passed")
			}
			if gotCategory.Name != req.Name || gotCategory.Description != req.Description {
				t.Fatalf("expected category %+v, got %+v", *req, *gotCategory)
			}
		})
	}
}

func TestCategoryServiceUpdateCategory(t *testing.T) {
	req := &entity.RequestCategory{Name: "Books", Description: "Reading"}
	missingErr := errors.New("missing")

	tests := []struct {
		name       string
		getErr     error
		updateErr  error
		wantErr    string
		wantUpdate bool
	}{
		{name: "missing", getErr: missingErr, wantErr: "category not found"},
		{name: "ok", wantUpdate: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				gotGetID    int64
				gotUpdate   bool
				gotUpdateID int64
				gotCategory *entity.Category
			)
			repo := &mockCategoryRepository{
				getByIDFunc: func(id int64) (*entity.ResponseCategory, error) {
					gotGetID = id
					if tt.getErr != nil {
						return nil, tt.getErr
					}
					return &entity.ResponseCategory{ID: id}, nil
				},
				updateFunc: func(id int64, category *entity.Category) error {
					gotUpdate = true
					gotUpdateID = id
					gotCategory = category
					return tt.updateErr
				},
			}

			svc := &categoryService{categoryRepository: repo}
			err := svc.UpdateCategory(7, req)
			if gotGetID != 7 {
				t.Fatalf("expected GetCategoryByID id 7, got %d", gotGetID)
			}
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if gotUpdate {
					t.Fatal("did not expect UpdateCategory to be called")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantUpdate != gotUpdate {
				t.Fatalf("expected update call %v, got %v", tt.wantUpdate, gotUpdate)
			}
			if gotUpdateID != 7 {
				t.Fatalf("expected UpdateCategory id 7, got %d", gotUpdateID)
			}
			if gotCategory == nil {
				t.Fatal("expected category to be passed")
			}
			if gotCategory.Name != req.Name || gotCategory.Description != req.Description {
				t.Fatalf("expected category %+v, got %+v", *req, *gotCategory)
			}
		})
	}
}

func TestCategoryServiceDeleteCategory(t *testing.T) {
	missingErr := errors.New("missing")

	tests := []struct {
		name       string
		getErr     error
		deleteErr  error
		wantErr    string
		wantDelete bool
	}{
		{name: "missing", getErr: missingErr, wantErr: "category not found"},
		{name: "ok", wantDelete: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				gotGetID    int64
				gotDelete   bool
				gotDeleteID int64
			)
			repo := &mockCategoryRepository{
				getByIDFunc: func(id int64) (*entity.ResponseCategory, error) {
					gotGetID = id
					if tt.getErr != nil {
						return nil, tt.getErr
					}
					return &entity.ResponseCategory{ID: id}, nil
				},
				deleteFunc: func(id int64) error {
					gotDelete = true
					gotDeleteID = id
					return tt.deleteErr
				},
			}

			svc := &categoryService{categoryRepository: repo}
			err := svc.DeleteCategory(9)
			if gotGetID != 9 {
				t.Fatalf("expected GetCategoryByID id 9, got %d", gotGetID)
			}
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if gotDelete {
					t.Fatal("did not expect DeleteCategory to be called")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantDelete != gotDelete {
				t.Fatalf("expected delete call %v, got %v", tt.wantDelete, gotDelete)
			}
			if gotDeleteID != 9 {
				t.Fatalf("expected DeleteCategory id 9, got %d", gotDeleteID)
			}
		})
	}
}

func TestCategoryServiceGetCategoryByID(t *testing.T) {
	resp := &entity.ResponseCategory{
		ID:          3,
		Name:        "Music",
		Description: "Instruments",
		CreatedAt:   time.Date(2023, 1, 2, 3, 4, 5, 0, time.UTC),
		UpdatedAt:   time.Date(2023, 1, 2, 4, 5, 6, 0, time.UTC),
	}
	repoErr := errors.New("repo error")

	tests := []struct {
		name    string
		resp    *entity.ResponseCategory
		err     error
		wantErr string
	}{
		{name: "ok", resp: resp},
		{name: "err", err: repoErr, wantErr: repoErr.Error()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotID int64
			repo := &mockCategoryRepository{
				getByIDFunc: func(id int64) (*entity.ResponseCategory, error) {
					gotID = id
					if tt.err != nil {
						return nil, tt.err
					}
					return tt.resp, nil
				},
			}

			svc := &categoryService{categoryRepository: repo}
			got, err := svc.GetCategoryByID(3)
			if gotID != 3 {
				t.Fatalf("expected GetCategoryByID id 3, got %d", gotID)
			}
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if got != nil {
					t.Fatalf("expected nil category, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.resp {
				t.Fatalf("expected response %+v, got %+v", tt.resp, got)
			}
		})
	}
}

func TestCategoryServiceGetAllCategories(t *testing.T) {
	resp := []entity.ResponseCategory{
		{
			ID:          1,
			Name:        "A",
			Description: "AA",
			CreatedAt:   time.Date(2023, 1, 2, 3, 4, 5, 0, time.UTC),
			UpdatedAt:   time.Date(2023, 1, 2, 4, 5, 6, 0, time.UTC),
		},
		{
			ID:          2,
			Name:        "B",
			Description: "BB",
			CreatedAt:   time.Date(2023, 2, 2, 3, 4, 5, 0, time.UTC),
			UpdatedAt:   time.Date(2023, 2, 2, 4, 5, 6, 0, time.UTC),
		},
	}
	repoErr := errors.New("repo error")

	tests := []struct {
		name    string
		resp    []entity.ResponseCategory
		err     error
		wantErr string
	}{
		{name: "empty", resp: []entity.ResponseCategory{}},
		{name: "ok", resp: resp},
		{name: "err", err: repoErr, wantErr: repoErr.Error()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCategoryRepository{
				getAllFunc: func() ([]entity.ResponseCategory, error) {
					if tt.err != nil {
						return nil, tt.err
					}
					return tt.resp, nil
				},
			}

			svc := &categoryService{categoryRepository: repo}
			got, err := svc.GetAllCategories()
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				if got != nil {
					t.Fatalf("expected nil categories, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.resp) {
				t.Fatalf("expected response %+v, got %+v", tt.resp, got)
			}
		})
	}
}
