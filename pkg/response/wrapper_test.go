package response

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type samplePayload struct {
	A string `json:"a"`
	B string `json:"b"`
}

func decodeBody[T any](t *testing.T, body *bytes.Buffer, out *T) {
	t.Helper()
	dec := json.NewDecoder(body)
	err := dec.Decode(out)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
}

func TestWriteJSONResponse(t *testing.T) {
	cases := []struct {
		name   string
		status int
		v      samplePayload
	}{
		{name: "ok", status: http.StatusOK, v: samplePayload{A: "x", B: "y"}},
		{name: "created", status: http.StatusCreated, v: samplePayload{A: "a", B: "b"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			WriteJSONResponse(rec, tc.status, tc.v)

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d", rec.Code, tc.status)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Fatalf("content-type = %q, want %q", ct, "application/json")
			}
			var got samplePayload
			decodeBody(t, rec.Body, &got)
			if got != tc.v {
				t.Fatalf("body = %+v, want %+v", got, tc.v)
			}
		})
	}
}

func TestParseJSON(t *testing.T) {
	cases := []struct {
		name      string
		body      string
		bodyNil   bool
		wantErr   bool
		wantValue samplePayload
	}{
		{name: "valid", body: `{"a":"x","b":"y"}`, wantValue: samplePayload{A: "x", B: "y"}},
		{name: "invalid", body: `{"a":`, wantErr: true},
		{name: "nil-body", bodyNil: true, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.bodyNil {
				req = &http.Request{Body: nil}
			} else {
				req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body))
			}

			var payload samplePayload
			err := ParseJSON(req, &payload)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if payload != tc.wantValue {
				t.Fatalf("payload = %+v, want %+v", payload, tc.wantValue)
			}
		})
	}
}

func TestSuccess(t *testing.T) {
	cases := []struct {
		name    string
		status  int
		code    int
		message string
		data    map[string]string
	}{
		{name: "ok", status: http.StatusOK, code: 1, message: "done", data: map[string]string{"k": "v"}},
		{name: "accepted", status: http.StatusAccepted, code: 204, message: "queued", data: map[string]string{"a": "b"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			Success(rec, tc.status, tc.code, tc.message, tc.data)

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d", rec.Code, tc.status)
			}
			var got APIResponse
			decodeBody(t, rec.Body, &got)
			if got.Code != strconv.Itoa(tc.code) {
				t.Fatalf("code = %q, want %q", got.Code, strconv.Itoa(tc.code))
			}
			if got.Message != tc.message {
				t.Fatalf("message = %v, want %v", got.Message, tc.message)
			}
			gotData, ok := got.Data.(map[string]any)
			if !ok {
				t.Fatalf("data type = %T, want map", got.Data)
			}
			expected := map[string]any{}
			for k, v := range tc.data {
				expected[k] = v
			}
			if !reflect.DeepEqual(gotData, expected) {
				t.Fatalf("data = %+v, want %+v", gotData, expected)
			}
		})
	}
}

func TestError(t *testing.T) {
	cases := []struct {
		name    string
		status  int
		code    int
		message string
		err     error
		wantMsg string
	}{
		{name: "with-error", status: http.StatusBadRequest, code: 9, message: "bad", err: errors.New("boom"), wantMsg: "bad: boom"},
		{name: "nil-error", status: http.StatusInternalServerError, code: 500, message: "oops", err: nil, wantMsg: "oops: %!s(<nil>)"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			Error(rec, tc.status, tc.code, tc.message, tc.err)

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d", rec.Code, tc.status)
			}
			var got APIResponse
			decodeBody(t, rec.Body, &got)
			if got.Code != strconv.Itoa(tc.code) {
				t.Fatalf("code = %q, want %q", got.Code, strconv.Itoa(tc.code))
			}
			if got.Message != tc.wantMsg {
				t.Fatalf("message = %v, want %v", got.Message, tc.wantMsg)
			}
			if got.Data != nil {
				t.Fatalf("data = %v, want nil", got.Data)
			}
		})
	}
}
