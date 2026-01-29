package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	constants "github.com/pandusatrianura/code-with-umam-second-meeting/constant"
	"github.com/pandusatrianura/code-with-umam-second-meeting/internal/products/entity"
	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/response"
)

type mockProductService struct {
	createFn func(*entity.RequestProduct) error
	updateFn func(int64, *entity.RequestProduct) error
	deleteFn func(int64) error
	getByID  func(int64) (*entity.ResponseProductWithCategories, error)
	getAllFn func() ([]entity.ResponseProductWithCategories, error)
	apiFn    func() entity.HealthCheck
}

func (m *mockProductService) CreateProduct(product *entity.RequestProduct) error {
	if m.createFn == nil {
		return nil
	}
	return m.createFn(product)
}

func (m *mockProductService) UpdateProduct(id int64, product *entity.RequestProduct) error {
	if m.updateFn == nil {
		return nil
	}
	return m.updateFn(id, product)
}

func (m *mockProductService) DeleteProduct(id int64) error {
	if m.deleteFn == nil {
		return nil
	}
	return m.deleteFn(id)
}

func (m *mockProductService) GetProductByID(id int64) (*entity.ResponseProductWithCategories, error) {
	if m.getByID == nil {
		return nil, nil
	}
	return m.getByID(id)
}

func (m *mockProductService) GetAllProducts() ([]entity.ResponseProductWithCategories, error) {
	if m.getAllFn == nil {
		return nil, nil
	}
	return m.getAllFn()
}

func (m *mockProductService) API() entity.HealthCheck {
	if m.apiFn == nil {
		return entity.HealthCheck{}
	}
	return m.apiFn()
}

func decodeAPIResponse(t *testing.T, rec *httptest.ResponseRecorder) response.APIResponse {
	t.Helper()
	var resp response.APIResponse
	dec := json.NewDecoder(rec.Body)
	if err := dec.Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}

func TestNewProductHandler(t *testing.T) {
	svc := &mockProductService{}
	h := NewProductHandler(svc)
	if h == nil {
		t.Fatalf("handler is nil")
	}
	if h.service != svc {
		t.Fatalf("service mismatch")
	}
}

func TestProductHandlerAPI(t *testing.T) {
	cases := []struct {
		name       string
		health     entity.HealthCheck
		wantStatus int
		wantCode   string
		wantMsg    string
	}{
		{name: "healthy", health: entity.HealthCheck{Name: "products", IsHealthy: true}, wantStatus: http.StatusOK, wantCode: strconv.Itoa(constants.SuccessCode), wantMsg: "products is healthy"},
		{name: "unhealthy", health: entity.HealthCheck{Name: "products", IsHealthy: false}, wantStatus: http.StatusServiceUnavailable, wantCode: strconv.Itoa(constants.ErrorCode), wantMsg: "products is not healthy"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockProductService{
				apiFn: func() entity.HealthCheck { return tc.health },
			}
			h := NewProductHandler(svc)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/products/health", nil)
			h.API(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			resp := decodeAPIResponse(t, rec)
			if resp.Code != tc.wantCode {
				t.Fatalf("code = %q, want %q", resp.Code, tc.wantCode)
			}
			msg, ok := resp.Message.(string)
			if !ok {
				t.Fatalf("message type = %T, want string", resp.Message)
			}
			if msg != tc.wantMsg {
				t.Fatalf("message = %q, want %q", msg, tc.wantMsg)
			}
		})
	}
}

func TestProductHandlerCreateProduct(t *testing.T) {
	validBody := `{"name":"a","price":10,"stock":2,"category_id":3}`
	validReq := entity.RequestProduct{Name: "a", Price: 10, Stock: 2, CategoryID: 3}

	cases := []struct {
		name       string
		body       string
		bodyNil    bool
		svcErr     error
		wantStatus int
		wantCode   string
		wantMsg    string
		wantPrefix bool
		wantCalled bool
	}{
		{name: "bad-json", body: `{"name":`, wantStatus: http.StatusBadRequest, wantCode: strconv.Itoa(constants.ErrorCode), wantMsg: constants.ErrInvalidProductRequest, wantPrefix: true, wantCalled: false},
		{name: "nil-body", bodyNil: true, wantStatus: http.StatusBadRequest, wantCode: strconv.Itoa(constants.ErrorCode), wantMsg: constants.ErrInvalidProductRequest, wantPrefix: true, wantCalled: false},
		{name: "svc-error", body: validBody, svcErr: errors.New("db"), wantStatus: http.StatusInternalServerError, wantCode: strconv.Itoa(constants.ErrorCode), wantMsg: "Product created failed: db", wantCalled: true},
		{name: "ok", body: validBody, wantStatus: http.StatusCreated, wantCode: strconv.Itoa(constants.SuccessCode), wantMsg: "Product created successfully", wantCalled: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			svc := &mockProductService{
				createFn: func(p *entity.RequestProduct) error {
					called = true
					if *p != validReq {
						t.Fatalf("request = %+v, want %+v", *p, validReq)
					}
					return tc.svcErr
				},
			}
			h := NewProductHandler(svc)
			rec := httptest.NewRecorder()

			var req *http.Request
			if tc.bodyNil {
				req = &http.Request{Body: nil}
			} else {
				req = httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(tc.body))
			}

			h.CreateProduct(rec, req)

			if called != tc.wantCalled {
				t.Fatalf("service called = %v, want %v", called, tc.wantCalled)
			}
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			resp := decodeAPIResponse(t, rec)
			if resp.Code != tc.wantCode {
				t.Fatalf("code = %q, want %q", resp.Code, tc.wantCode)
			}
			msg, ok := resp.Message.(string)
			if !ok {
				t.Fatalf("message type = %T, want string", resp.Message)
			}
			if tc.wantPrefix {
				if !strings.HasPrefix(msg, tc.wantMsg) {
					t.Fatalf("message = %q, want prefix %q", msg, tc.wantMsg)
				}
			} else if msg != tc.wantMsg {
				t.Fatalf("message = %q, want %q", msg, tc.wantMsg)
			}
			if tc.wantStatus == http.StatusCreated && resp.Data != nil {
				t.Fatalf("data = %v, want nil", resp.Data)
			}
		})
	}
}

func TestProductHandlerUpdateProduct(t *testing.T) {
	validBody := `{"name":"a","price":10,"stock":2,"category_id":3}`
	validReq := entity.RequestProduct{Name: "a", Price: 10, Stock: 2, CategoryID: 3}

	cases := []struct {
		name       string
		path       string
		body       string
		wantStatus int
		wantCode   string
		wantMsg    string
		wantPrefix bool
		svcErr     error
		wantCalled bool
		wantID     int64
	}{
		{name: "bad-id", path: "/products/abc", body: validBody, wantStatus: http.StatusBadRequest, wantCode: strconv.Itoa(constants.ErrorCode), wantMsg: constants.ErrInvalidProductID, wantPrefix: true, wantCalled: false},
		{name: "bad-json", path: "/products/12", body: `{"name":`, wantStatus: http.StatusBadRequest, wantCode: strconv.Itoa(constants.ErrorCode), wantMsg: constants.ErrInvalidProductRequest, wantPrefix: true, wantCalled: false, wantID: 12},
		{name: "svc-error", path: "/products/12", body: validBody, svcErr: errors.New("db"), wantStatus: http.StatusInternalServerError, wantCode: strconv.Itoa(constants.ErrorCode), wantMsg: "Product updated failed: db", wantCalled: true, wantID: 12},
		{name: "ok", path: "/products/12", body: validBody, wantStatus: http.StatusOK, wantCode: strconv.Itoa(constants.SuccessCode), wantMsg: "Product updated successfully", wantCalled: true, wantID: 12},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			var gotID int64
			svc := &mockProductService{
				updateFn: func(id int64, p *entity.RequestProduct) error {
					called = true
					gotID = id
					if *p != validReq {
						t.Fatalf("request = %+v, want %+v", *p, validReq)
					}
					return tc.svcErr
				},
			}
			h := NewProductHandler(svc)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, tc.path, strings.NewReader(tc.body))

			h.UpdateProduct(rec, req)

			if called != tc.wantCalled {
				t.Fatalf("service called = %v, want %v", called, tc.wantCalled)
			}
			if tc.wantCalled && gotID != tc.wantID {
				t.Fatalf("id = %d, want %d", gotID, tc.wantID)
			}
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			resp := decodeAPIResponse(t, rec)
			if resp.Code != tc.wantCode {
				t.Fatalf("code = %q, want %q", resp.Code, tc.wantCode)
			}
			msg, ok := resp.Message.(string)
			if !ok {
				t.Fatalf("message type = %T, want string", resp.Message)
			}
			if tc.wantPrefix {
				if !strings.HasPrefix(msg, tc.wantMsg) {
					t.Fatalf("message = %q, want prefix %q", msg, tc.wantMsg)
				}
			} else if msg != tc.wantMsg {
				t.Fatalf("message = %q, want %q", msg, tc.wantMsg)
			}
		})
	}
}

func TestProductHandlerDeleteProduct(t *testing.T) {
	cases := []struct {
		name       string
		path       string
		wantStatus int
		wantCode   string
		wantMsg    string
		wantPrefix bool
		svcErr     error
		wantCalled bool
		wantID     int64
	}{
		{name: "bad-id", path: "/products/abc", wantStatus: http.StatusBadRequest, wantCode: strconv.Itoa(constants.ErrorCode), wantMsg: constants.ErrInvalidProductID, wantPrefix: true, wantCalled: false},
		{name: "svc-error", path: "/products/9", svcErr: errors.New("db"), wantStatus: http.StatusInternalServerError, wantCode: strconv.Itoa(constants.ErrorCode), wantMsg: "Product delete failed: db", wantCalled: true, wantID: 9},
		{name: "ok", path: "/products/9", wantStatus: http.StatusOK, wantCode: strconv.Itoa(constants.SuccessCode), wantMsg: "Product deleted successfully", wantCalled: true, wantID: 9},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			var gotID int64
			svc := &mockProductService{
				deleteFn: func(id int64) error {
					called = true
					gotID = id
					return tc.svcErr
				},
			}
			h := NewProductHandler(svc)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, tc.path, nil)

			h.DeleteProduct(rec, req)

			if called != tc.wantCalled {
				t.Fatalf("service called = %v, want %v", called, tc.wantCalled)
			}
			if tc.wantCalled && gotID != tc.wantID {
				t.Fatalf("id = %d, want %d", gotID, tc.wantID)
			}
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			resp := decodeAPIResponse(t, rec)
			if resp.Code != tc.wantCode {
				t.Fatalf("code = %q, want %q", resp.Code, tc.wantCode)
			}
			msg, ok := resp.Message.(string)
			if !ok {
				t.Fatalf("message type = %T, want string", resp.Message)
			}
			if tc.wantPrefix {
				if !strings.HasPrefix(msg, tc.wantMsg) {
					t.Fatalf("message = %q, want prefix %q", msg, tc.wantMsg)
				}
			} else if msg != tc.wantMsg {
				t.Fatalf("message = %q, want %q", msg, tc.wantMsg)
			}
		})
	}
}

func TestProductHandlerGetProductByID(t *testing.T) {
	product := &entity.ResponseProductWithCategories{ID: 7, Name: "p1", Price: 10, Stock: 2, CategoryID: 3, CategoryName: "c1"}

	cases := []struct {
		name       string
		path       string
		wantStatus int
		wantCode   string
		wantMsg    string
		wantPrefix bool
		svcErr     error
		wantCalled bool
		wantID     int64
	}{
		{name: "bad-id", path: "/products/abc", wantStatus: http.StatusBadRequest, wantCode: strconv.Itoa(constants.ErrorCode), wantMsg: constants.ErrInvalidProductID, wantPrefix: true, wantCalled: false},
		{name: "svc-error", path: "/products/7", svcErr: errors.New("db"), wantStatus: http.StatusInternalServerError, wantCode: strconv.Itoa(constants.ErrorCode), wantMsg: "Product retrieved failed: db", wantCalled: true, wantID: 7},
		{name: "ok", path: "/products/7", wantStatus: http.StatusOK, wantCode: strconv.Itoa(constants.SuccessCode), wantMsg: "Product retrieved successfully", wantCalled: true, wantID: 7},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			var gotID int64
			svc := &mockProductService{
				getByID: func(id int64) (*entity.ResponseProductWithCategories, error) {
					called = true
					gotID = id
					if tc.svcErr != nil {
						return nil, tc.svcErr
					}
					return product, nil
				},
			}
			h := NewProductHandler(svc)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)

			h.GetProductByID(rec, req)

			if called != tc.wantCalled {
				t.Fatalf("service called = %v, want %v", called, tc.wantCalled)
			}
			if tc.wantCalled && gotID != tc.wantID {
				t.Fatalf("id = %d, want %d", gotID, tc.wantID)
			}
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			resp := decodeAPIResponse(t, rec)
			if resp.Code != tc.wantCode {
				t.Fatalf("code = %q, want %q", resp.Code, tc.wantCode)
			}
			msg, ok := resp.Message.(string)
			if !ok {
				t.Fatalf("message type = %T, want string", resp.Message)
			}
			if tc.wantPrefix {
				if !strings.HasPrefix(msg, tc.wantMsg) {
					t.Fatalf("message = %q, want prefix %q", msg, tc.wantMsg)
				}
			} else if msg != tc.wantMsg {
				t.Fatalf("message = %q, want %q", msg, tc.wantMsg)
			}
			if tc.name == "ok" {
				data, ok := resp.Data.(map[string]any)
				if !ok {
					t.Fatalf("data type = %T, want map", resp.Data)
				}
				if data["id"] != float64(product.ID) {
					t.Fatalf("data.id = %v, want %d", data["id"], product.ID)
				}
				if data["name"] != product.Name {
					t.Fatalf("data.name = %v, want %s", data["name"], product.Name)
				}
			}
		})
	}
}

func TestProductHandlerGetAllProducts(t *testing.T) {
	products := []entity.ResponseProductWithCategories{
		{ID: 1, Name: "p1", Price: 10, Stock: 2, CategoryID: 3, CategoryName: "c1"},
		{ID: 2, Name: "p2", Price: 11, Stock: 3, CategoryID: 4, CategoryName: "c2"},
	}

	cases := []struct {
		name       string
		wantStatus int
		wantCode   string
		wantMsg    string
		svcErr     error
	}{
		{name: "svc-error", wantStatus: http.StatusInternalServerError, wantCode: strconv.Itoa(constants.ErrorCode), wantMsg: "Products retrieved failed: db", svcErr: errors.New("db")},
		{name: "ok", wantStatus: http.StatusOK, wantCode: strconv.Itoa(constants.SuccessCode), wantMsg: "Products retrieved successfully"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockProductService{
				getAllFn: func() ([]entity.ResponseProductWithCategories, error) {
					if tc.svcErr != nil {
						return nil, tc.svcErr
					}
					return products, nil
				},
			}
			h := NewProductHandler(svc)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/products", nil)

			h.GetAllProducts(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			resp := decodeAPIResponse(t, rec)
			if resp.Code != tc.wantCode {
				t.Fatalf("code = %q, want %q", resp.Code, tc.wantCode)
			}
			msg, ok := resp.Message.(string)
			if !ok {
				t.Fatalf("message type = %T, want string", resp.Message)
			}
			if msg != tc.wantMsg {
				t.Fatalf("message = %q, want %q", msg, tc.wantMsg)
			}
			if tc.name == "ok" {
				data, ok := resp.Data.([]any)
				if !ok {
					t.Fatalf("data type = %T, want slice", resp.Data)
				}
				if len(data) != len(products) {
					t.Fatalf("data len = %d, want %d", len(data), len(products))
				}
			}
		})
	}
}
