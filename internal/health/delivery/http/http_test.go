package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pandusatrianura/code-with-umam-second-meeting/internal/health/entity"
	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/response"
)

type mockHealthService struct {
	apiResult entity.HealthCheck
	dbResult  entity.HealthCheck
	dbErr     error
}

func (m mockHealthService) API() entity.HealthCheck {
	return m.apiResult
}

func (m mockHealthService) DB() (entity.HealthCheck, error) {
	return m.dbResult, m.dbErr
}

func TestNewHealthHandler(t *testing.T) {
	svc := mockHealthService{}
	h := NewHealthHandler(svc)
	if h == nil {
		t.Fatalf("expected handler")
	}
	if h.service != svc {
		t.Fatalf("expected service to match")
	}
}

func TestHealthHandlerAPI(t *testing.T) {
	cases := []struct {
		name        string
		apiResult   entity.HealthCheck
		wantStatus  int
		wantCode    string
		wantMessage string
	}{
		{
			name: "healthy",
			apiResult: entity.HealthCheck{
				Name:      "svc",
				IsHealthy: true,
			},
			wantStatus:  http.StatusOK,
			wantCode:    "1000",
			wantMessage: "svc is healthy",
		},
		{
			name: "unhealthy",
			apiResult: entity.HealthCheck{
				Name:      "svc",
				IsHealthy: false,
			},
			wantStatus:  http.StatusServiceUnavailable,
			wantCode:    "2000",
			wantMessage: "svc is not healthy",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHealthHandler(mockHealthService{apiResult: tc.apiResult})
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/health/service", nil)

			h.API(w, r)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tc.wantStatus)
			}
			if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
				t.Fatalf("content-type = %q", ct)
			}

			body := decodeAPIResponse(t, resp)
			if body.Code != tc.wantCode {
				t.Fatalf("code = %q, want %q", body.Code, tc.wantCode)
			}
			msg, ok := body.Message.(string)
			if !ok {
				t.Fatalf("message type = %T", body.Message)
			}
			if msg != tc.wantMessage {
				t.Fatalf("message = %q, want %q", msg, tc.wantMessage)
			}
		})
	}
}

func TestHealthHandlerDB(t *testing.T) {
	cases := []struct {
		name        string
		dbResult    entity.HealthCheck
		dbErr       error
		wantStatus  int
		wantCode    string
		wantMessage string
		wantPanic   bool
	}{
		{
			name: "healthy",
			dbResult: entity.HealthCheck{
				Name:      "db",
				IsHealthy: true,
			},
			wantStatus:  http.StatusOK,
			wantCode:    "1000",
			wantMessage: "db is healthy",
		},
		{
			name: "unhealthy",
			dbResult: entity.HealthCheck{
				Name:      "db",
				IsHealthy: false,
			},
			dbErr:       errors.New("no conn"),
			wantStatus:  http.StatusServiceUnavailable,
			wantCode:    "2000",
			wantMessage: "db is not healthy because no conn",
		},
		{
			name: "err",
			dbResult: entity.HealthCheck{
				Name:      "db",
				IsHealthy: true,
			},
			dbErr:       errors.New("timeout"),
			wantStatus:  http.StatusServiceUnavailable,
			wantCode:    "2000",
			wantMessage: "db is not healthy because timeout",
		},
		{
			name: "nilerr",
			dbResult: entity.HealthCheck{
				Name:      "db",
				IsHealthy: false,
			},
			wantPanic: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHealthHandler(mockHealthService{dbResult: tc.dbResult, dbErr: tc.dbErr})
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/health/db", nil)

			if tc.wantPanic {
				defer func() {
					if recover() == nil {
						t.Fatalf("expected panic")
					}
				}()
				h.DB(w, r)
				return
			}

			h.DB(w, r)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tc.wantStatus)
			}
			if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
				t.Fatalf("content-type = %q", ct)
			}

			body := decodeAPIResponse(t, resp)
			if body.Code != tc.wantCode {
				t.Fatalf("code = %q, want %q", body.Code, tc.wantCode)
			}
			msg, ok := body.Message.(string)
			if !ok {
				t.Fatalf("message type = %T", body.Message)
			}
			if msg != tc.wantMessage {
				t.Fatalf("message = %q, want %q", msg, tc.wantMessage)
			}
		})
	}
}

func decodeAPIResponse(t *testing.T, resp *http.Response) response.APIResponse {
	t.Helper()
	var body response.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body
}
