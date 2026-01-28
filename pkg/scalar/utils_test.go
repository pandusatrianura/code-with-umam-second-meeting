package scalar

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureFileURL(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	baseDir := t.TempDir()
	absPath := filepath.Join(baseDir, "abs.txt")

	baseSetup := func(t *testing.T) func() {
		if err := os.Chdir(baseDir); err != nil {
			t.Fatalf("chdir: %v", err)
		}
		return func() {
			_ = os.Chdir(origDir)
		}
	}

	cases := []struct {
		name    string
		in      string
		want    string
		wantErr bool
		setup   func(t *testing.T) func()
	}{
		{name: "file-abs", in: "file://" + absPath, want: "file://" + absPath, setup: baseSetup},
		{name: "abs", in: absPath, want: "file://" + absPath, setup: baseSetup},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				t.Cleanup(tc.setup(t))
			}
			got, err := ensureFileURL(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ensureFileURL error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("want %q got %q", tc.want, got)
			}
		})
	}
}

func TestFetchContentFromURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	defer server.Close()

	cases := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{name: "ok", url: server.URL, want: "ok"},
		{name: "bad", url: "http://[::1", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := fetchContentFromURL(tc.url)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("fetchContentFromURL error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("want %q got %q", tc.want, got)
			}
		})
	}
}

func TestReadFileFromURL(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "data.txt")
	if err := os.WriteFile(filePath, []byte("data"), 0o600); err != nil {
		t.Fatalf("writefile: %v", err)
	}

	cases := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{name: "ok", url: "file://" + filePath, want: "data"},
		{name: "scheme", url: "http://example.com", wantErr: true},
		{name: "parse", url: "://bad", wantErr: true},
		{name: "missing", url: "file://" + filepath.Join(tempDir, "missing.txt"), wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := readFileFromURL(tc.url)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("readFileFromURL error: %v", err)
			}
			if string(got) != tc.want {
				t.Fatalf("want %q got %q", tc.want, string(got))
			}
		})
	}
}
