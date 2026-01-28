package scalar

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSafeJSONConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		options *Options
		want    string
	}{
		{
			name:    "cdn",
			options: &Options{CDN: "cdn"},
			want:    `{&quot;cdn&quot;:&quot;cdn&quot;}`,
		},
		{
			name:    "bool",
			options: &Options{DarkMode: true},
			want:    `{&quot;darkMode&quot;:true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := safeJSONConfiguration(tt.options)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSpecContentHandler(t *testing.T) {
	tests := []struct {
		name       string
		input      interface{}
		wantString string
		wantMap    map[string]interface{}
	}{
		{
			name:    "func",
			input:   func() map[string]interface{} { return map[string]interface{}{"k": "v"} },
			wantMap: map[string]interface{}{"k": "v"},
		},
		{
			name:    "map",
			input:   map[string]interface{}{"k": "v"},
			wantMap: map[string]interface{}{"k": "v"},
		},
		{
			name:       "string",
			input:      "raw",
			wantString: "raw",
		},
		{
			name:       "other",
			input:      123,
			wantString: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := specContentHandler(tt.input)
			if tt.wantMap != nil {
				var gotMap map[string]interface{}
				if err := json.Unmarshal([]byte(got), &gotMap); err != nil {
					t.Fatalf("unmarshal got %q: %v", got, err)
				}
				if !reflect.DeepEqual(gotMap, tt.wantMap) {
					t.Fatalf("got %v, want %v", gotMap, tt.wantMap)
				}
				return
			}
			if got != tt.wantString {
				t.Fatalf("got %q, want %q", got, tt.wantString)
			}
		})
	}
}

func TestApiReferenceHTML(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "spec.json")
	if err := os.WriteFile(filePath, []byte("file-spec"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("remote-spec"))
	}))
	defer server.Close()

	tests := []struct {
		name    string
		options *Options
		wantErr bool
		check   func(t *testing.T, got string)
	}{
		{
			name:    "missing",
			options: &Options{},
			wantErr: true,
		},
		{
			name: "specContent",
			options: &Options{
				CDN:         "cdn",
				SpecContent: map[string]interface{}{"openapi": "3.0.0"},
			},
			check: func(t *testing.T, got string) {
				if !strings.Contains(got, "<title>Scalar API Reference</title>") {
					t.Fatalf("expected default title in output")
				}
				if !strings.Contains(got, "--scalar-color-1") {
					t.Fatalf("expected custom theme css in output")
				}
				if !strings.Contains(got, `{"openapi":"3.0.0"}`) {
					t.Fatalf("expected spec content in output")
				}
				if !strings.Contains(got, `&quot;cdn&quot;:&quot;cdn&quot;`) {
					t.Fatalf("expected escaped config in output")
				}
			},
		},
		{
			name: "customTitleTheme",
			options: &Options{
				SpecContent: "spec",
				Theme:       ThemeMoon,
				CustomOptions: CustomOptions{
					PageTitle: "My Title",
				},
			},
			check: func(t *testing.T, got string) {
				if !strings.Contains(got, "<title>My Title</title>") {
					t.Fatalf("expected custom title in output")
				}
				if strings.Contains(got, "--scalar-color-1") {
					t.Fatalf("expected no custom theme css in output")
				}
				if !strings.Contains(got, ">spec</script>") {
					t.Fatalf("expected spec content in output")
				}
			},
		},
		{
			name: "http",
			options: &Options{
				SpecURL: server.URL,
			},
			check: func(t *testing.T, got string) {
				if !strings.Contains(got, "remote-spec") {
					t.Fatalf("expected remote spec content in output")
				}
			},
		},
		{
			name: "file",
			options: &Options{
				SpecURL: filePath,
			},
			check: func(t *testing.T, got string) {
				if !strings.Contains(got, "file-spec") {
					t.Fatalf("expected file spec content in output")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ApiReferenceHTML(tt.options)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
