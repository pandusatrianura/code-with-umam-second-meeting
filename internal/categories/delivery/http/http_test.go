package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	constants "github.com/pandusatrianura/code-with-umam-second-meeting/constant"
	"github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/entity"
)

type mockCategoryService struct {
	createFn  func(*entity.RequestCategory) error
	updateFn  func(int64, *entity.RequestCategory) error
	deleteFn  func(int64) error
	getByIDFn func(int64) (*entity.ResponseCategory, error)
	getAllFn  func() ([]entity.ResponseCategory, error)
	apiFn     func() entity.HealthCheck

	createCalls  int
	updateCalls  int
	deleteCalls  int
	getByIDCalls int
	getAllCalls  int
	apiCalls     int

	createReq *entity.RequestCategory
	updateReq *entity.RequestCategory
	updateID  int64
	deleteID  int64
	getByIDID int64
}

func (m *mockCategoryService) CreateCategory(requestCategory *entity.RequestCategory) error {
	m.createCalls++
	m.createReq = requestCategory
	if m.createFn != nil {
		return m.createFn(requestCategory)
	}
	return nil
}

func (m *mockCategoryService) UpdateCategory(id int64, requestCategory *entity.RequestCategory) error {
	m.updateCalls++
	m.updateID = id
	m.updateReq = requestCategory
	if m.updateFn != nil {
		return m.updateFn(id, requestCategory)
	}
	return nil
}

func (m *mockCategoryService) DeleteCategory(id int64) error {
	m.deleteCalls++
	m.deleteID = id
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return nil
}

func (m *mockCategoryService) GetCategoryByID(id int64) (*entity.ResponseCategory, error) {
	m.getByIDCalls++
	m.getByIDID = id
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, nil
}

func (m *mockCategoryService) GetAllCategories() ([]entity.ResponseCategory, error) {
	m.getAllCalls++
	if m.getAllFn != nil {
		return m.getAllFn()
	}
	return nil, nil
}

func (m *mockCategoryService) API() entity.HealthCheck {
	m.apiCalls++
	if m.apiFn != nil {
		return m.apiFn()
	}
	return entity.HealthCheck{}
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return body
}

func TestNewCategoryHandler(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "init"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockCategoryService{}
			h := NewCategoryHandler(svc)
			if h == nil {
				t.Fatal("expected handler")
			}
			if h.service != svc {
				t.Fatalf("expected service to be set")
			}
		})
	}
}

func TestCategoryHandlerAPI(t *testing.T) {
	cases := []struct {
		name       string
		health     entity.HealthCheck
		wantStatus int
		wantCode   string
		wantMsg    string
	}{
		{
			name:       "healthy",
			health:     entity.HealthCheck{Name: "svc", IsHealthy: true},
			wantStatus: http.StatusOK,
			wantCode:   "1000",
			wantMsg:    "svc is healthy",
		},
		{
			name:       "unhealthy",
			health:     entity.HealthCheck{Name: "svc", IsHealthy: false},
			wantStatus: http.StatusServiceUnavailable,
			wantCode:   "2000",
			wantMsg:    "svc is not healthy",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockCategoryService{
				apiFn: func() entity.HealthCheck {
					return tc.health
				},
			}
			h := NewCategoryHandler(svc)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/categories/health", nil)

			h.API(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, rec.Code)
			}
			body := decodeBody(t, rec)
			if body["code"] != tc.wantCode {
				t.Fatalf("expected code %q, got %v", tc.wantCode, body["code"])
			}
			if body["message"] != tc.wantMsg {
				t.Fatalf("expected message %q, got %v", tc.wantMsg, body["message"])
			}
			if svc.apiCalls != 1 {
				t.Fatalf("expected api calls 1, got %d", svc.apiCalls)
			}
		})
	}
}

func TestCategoryHandlerCreateCategory(t *testing.T) {
	cases := []struct {
		name       string
		body       *strings.Reader
		bodyNil    bool
		createErr  error
		wantStatus int
		wantCode   string
		wantMsg    string
		wantCalls  int
	}{
		{
			name:       "no-body",
			bodyNil:    true,
			wantStatus: http.StatusBadRequest,
			wantCode:   "2000",
			wantMsg:    constants.ErrInvalidCategoryRequest,
			wantCalls:  0,
		},
		{
			name:       "bad-json",
			body:       strings.NewReader("{"),
			wantStatus: http.StatusBadRequest,
			wantCode:   "2000",
			wantMsg:    constants.ErrInvalidCategoryRequest,
			wantCalls:  0,
		},
		{
			name:       "service-error",
			body:       strings.NewReader(`{"name":"A","description":"B"}`),
			createErr:  errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "2000",
			wantMsg:    "Category created failed",
			wantCalls:  1,
		},
		{
			name:       "ok",
			body:       strings.NewReader(`{"name":"A","description":"B"}`),
			wantStatus: http.StatusCreated,
			wantCode:   "1000",
			wantMsg:    "Category created successfully",
			wantCalls:  1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockCategoryService{
				createFn: func(_ *entity.RequestCategory) error {
					return tc.createErr
				},
			}
			h := NewCategoryHandler(svc)
			rec := httptest.NewRecorder()
			var req *http.Request
			if tc.bodyNil {
				req = httptest.NewRequest(http.MethodPost, "/api/categories", nil)
				req.Body = nil
			} else {
				req = httptest.NewRequest(http.MethodPost, "/api/categories", tc.body)
			}

			h.CreateCategory(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, rec.Code)
			}
			body := decodeBody(t, rec)
			if body["code"] != tc.wantCode {
				t.Fatalf("expected code %q, got %v", tc.wantCode, body["code"])
			}
			msg, _ := body["message"].(string)
			if !strings.Contains(msg, tc.wantMsg) {
				t.Fatalf("expected message to contain %q, got %q", tc.wantMsg, msg)
			}
			if svc.createCalls != tc.wantCalls {
				t.Fatalf("expected create calls %d, got %d", tc.wantCalls, svc.createCalls)
			}
			if tc.name == "ok" && svc.createReq != nil {
				if svc.createReq.Name != "A" || svc.createReq.Description != "B" {
					t.Fatalf("unexpected create request: %#v", svc.createReq)
				}
			}
		})
	}
}

func TestCategoryHandlerUpdateCategory(t *testing.T) {
	cases := []struct {
		name       string
		path       string
		body       *strings.Reader
		updateErr  error
		wantStatus int
		wantCode   string
		wantMsg    string
		wantCalls  int
	}{
		{
			name:       "bad-id",
			path:       "/categories/abc",
			body:       strings.NewReader(`{"name":"A"}`),
			wantStatus: http.StatusBadRequest,
			wantCode:   "2000",
			wantMsg:    constants.ErrInvalidCategoryID,
			wantCalls:  0,
		},
		{
			name:       "bad-json",
			path:       "/categories/1",
			body:       strings.NewReader("{"),
			wantStatus: http.StatusBadRequest,
			wantCode:   "2000",
			wantMsg:    constants.ErrInvalidCategoryRequest,
			wantCalls:  0,
		},
		{
			name:       "service-error",
			path:       "/categories/1",
			body:       strings.NewReader(`{"name":"A","description":"B"}`),
			updateErr:  errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "2000",
			wantMsg:    "Category updated failed",
			wantCalls:  1,
		},
		{
			name:       "ok",
			path:       "/categories/1",
			body:       strings.NewReader(`{"name":"A","description":"B"}`),
			wantStatus: http.StatusOK,
			wantCode:   "1000",
			wantMsg:    "Category updated successfully",
			wantCalls:  1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockCategoryService{
				updateFn: func(_ int64, _ *entity.RequestCategory) error {
					return tc.updateErr
				},
			}
			h := NewCategoryHandler(svc)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, tc.path, tc.body)

			h.UpdateCategory(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, rec.Code)
			}
			body := decodeBody(t, rec)
			if body["code"] != tc.wantCode {
				t.Fatalf("expected code %q, got %v", tc.wantCode, body["code"])
			}
			msg, _ := body["message"].(string)
			if !strings.Contains(msg, tc.wantMsg) {
				t.Fatalf("expected message to contain %q, got %q", tc.wantMsg, msg)
			}
			if svc.updateCalls != tc.wantCalls {
				t.Fatalf("expected update calls %d, got %d", tc.wantCalls, svc.updateCalls)
			}
			if tc.name == "ok" && svc.updateReq != nil {
				if svc.updateID != 1 || svc.updateReq.Name != "A" || svc.updateReq.Description != "B" {
					t.Fatalf("unexpected update request: id=%d req=%#v", svc.updateID, svc.updateReq)
				}
			}
		})
	}
}

func TestCategoryHandlerDeleteCategory(t *testing.T) {
	cases := []struct {
		name       string
		path       string
		deleteErr  error
		wantStatus int
		wantCode   string
		wantMsg    string
		wantCalls  int
	}{
		{
			name:       "bad-id",
			path:       "/categories/abc",
			wantStatus: http.StatusBadRequest,
			wantCode:   "2000",
			wantMsg:    constants.ErrInvalidCategoryID,
			wantCalls:  0,
		},
		{
			name:       "service-error",
			path:       "/categories/1",
			deleteErr:  errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "2000",
			wantMsg:    "Category delete failed",
			wantCalls:  1,
		},
		{
			name:       "ok",
			path:       "/categories/1",
			wantStatus: http.StatusOK,
			wantCode:   "1000",
			wantMsg:    "Category deleted successfully",
			wantCalls:  1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockCategoryService{
				deleteFn: func(_ int64) error {
					return tc.deleteErr
				},
			}
			h := NewCategoryHandler(svc)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, tc.path, nil)

			h.DeleteCategory(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, rec.Code)
			}
			body := decodeBody(t, rec)
			if body["code"] != tc.wantCode {
				t.Fatalf("expected code %q, got %v", tc.wantCode, body["code"])
			}
			msg, _ := body["message"].(string)
			if !strings.Contains(msg, tc.wantMsg) {
				t.Fatalf("expected message to contain %q, got %q", tc.wantMsg, msg)
			}
			if svc.deleteCalls != tc.wantCalls {
				t.Fatalf("expected delete calls %d, got %d", tc.wantCalls, svc.deleteCalls)
			}
			if tc.name == "ok" && svc.deleteID != 1 {
				t.Fatalf("unexpected delete id: %d", svc.deleteID)
			}
		})
	}
}

func TestCategoryHandlerGetCategoryByID(t *testing.T) {
	cases := []struct {
		name       string
		path       string
		result     *entity.ResponseCategory
		getErr     error
		wantStatus int
		wantCode   string
		wantMsg    string
		wantCalls  int
	}{
		{
			name:       "bad-id",
			path:       "/categories/abc",
			wantStatus: http.StatusBadRequest,
			wantCode:   "2000",
			wantMsg:    constants.ErrInvalidCategoryID,
			wantCalls:  0,
		},
		{
			name:       "service-error",
			path:       "/categories/1",
			getErr:     errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "2000",
			wantMsg:    "Category retrieved failed",
			wantCalls:  1,
		},
		{
			name: "ok",
			path: "/categories/1",
			result: &entity.ResponseCategory{
				ID:          1,
				Name:        "A",
				Description: "B",
				CreatedAt:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			wantStatus: http.StatusOK,
			wantCode:   "1000",
			wantMsg:    "Category retrieved successfully",
			wantCalls:  1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockCategoryService{
				getByIDFn: func(_ int64) (*entity.ResponseCategory, error) {
					return tc.result, tc.getErr
				},
			}
			h := NewCategoryHandler(svc)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)

			h.GetCategoryByID(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, rec.Code)
			}
			body := decodeBody(t, rec)
			if body["code"] != tc.wantCode {
				t.Fatalf("expected code %q, got %v", tc.wantCode, body["code"])
			}
			msg, _ := body["message"].(string)
			if !strings.Contains(msg, tc.wantMsg) {
				t.Fatalf("expected message to contain %q, got %q", tc.wantMsg, msg)
			}
			if svc.getByIDCalls != tc.wantCalls {
				t.Fatalf("expected getByID calls %d, got %d", tc.wantCalls, svc.getByIDCalls)
			}
			if tc.name == "ok" {
				data, _ := body["data"].(map[string]any)
				if data["id"] != float64(1) || data["name"] != "A" || data["description"] != "B" {
					t.Fatalf("unexpected data: %v", data)
				}
			}
		})
	}
}

func TestCategoryHandlerGetAllCategories(t *testing.T) {
	cases := []struct {
		name       string
		result     []entity.ResponseCategory
		getErr     error
		wantStatus int
		wantCode   string
		wantMsg    string
		wantCalls  int
	}{
		{
			name:       "service-error",
			getErr:     errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "2000",
			wantMsg:    "Categories retrieved failed",
			wantCalls:  1,
		},
		{
			name: "ok",
			result: []entity.ResponseCategory{
				{ID: 1, Name: "A", Description: "B"},
				{ID: 2, Name: "C", Description: "D"},
			},
			wantStatus: http.StatusOK,
			wantCode:   "1000",
			wantMsg:    "Categories retrieved successfully",
			wantCalls:  1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockCategoryService{
				getAllFn: func() ([]entity.ResponseCategory, error) {
					return tc.result, tc.getErr
				},
			}
			h := NewCategoryHandler(svc)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/categories", nil)

			h.GetAllCategories(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, rec.Code)
			}
			body := decodeBody(t, rec)
			if body["code"] != tc.wantCode {
				t.Fatalf("expected code %q, got %v", tc.wantCode, body["code"])
			}
			msg, _ := body["message"].(string)
			if !strings.Contains(msg, tc.wantMsg) {
				t.Fatalf("expected message to contain %q, got %q", tc.wantMsg, msg)
			}
			if svc.getAllCalls != tc.wantCalls {
				t.Fatalf("expected getAll calls %d, got %d", tc.wantCalls, svc.getAllCalls)
			}
			if tc.name == "ok" {
				data, _ := body["data"].([]any)
				if len(data) != 2 {
					t.Fatalf("unexpected data length: %d", len(data))
				}
				first, _ := data[0].(map[string]any)
				if first["id"] != float64(1) || first["name"] != "A" || first["description"] != "B" {
					t.Fatalf("unexpected data: %v", first)
				}
			}
		})
	}
}
