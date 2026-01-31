package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestInitConfig(t *testing.T) {
	cases := []struct {
		name   string
		setup  func(t *testing.T, dir string)
		assert func(t *testing.T)
	}{
		{
			name: "env",
			setup: func(t *testing.T, _ string) {
				t.Setenv("FOO_BAR", "fromenv")
			},
			assert: func(t *testing.T) {
				got := viper.GetString("foo.bar")
				if got != "fromenv" {
					t.Fatalf("expected env value, got %q", got)
				}
			},
		},
		{
			name: "dotenv",
			setup: func(t *testing.T, dir string) {
				path := filepath.Join(dir, ".env")
				if err := os.WriteFile(path, []byte("APP_NAME=fromfile\n"), 0644); err != nil {
					t.Fatalf("write .env: %v", err)
				}
			},
			assert: func(t *testing.T) {
				got := viper.GetString("APP_NAME")
				if got != "fromfile" {
					t.Fatalf("expected file value, got %q", got)
				}
			},
		},
		{
			name:  "missing",
			setup: nil,
			assert: func(t *testing.T) {
				if used := viper.ConfigFileUsed(); used != "" {
					t.Fatalf("expected no config file, got %q", used)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			viper.Reset()
			dir := t.TempDir()
			withWorkingDir(t, dir, func() {
				if tc.setup != nil {
					tc.setup(t, dir)
				}
				InitConfig()
				tc.assert(t)
			})
		})
	}
}

func withWorkingDir(t *testing.T, dir string, fn func()) {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})
	fn()
}
