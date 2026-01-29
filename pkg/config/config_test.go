package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestInitConfig(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		envKey         string
		envVal         string
		envLookup      string
		wantEnvVal     string
		envFileContent string
		fileKey        string
		wantFileVal    string
		wantConfigUsed bool
	}

	cases := []testCase{
		{
			name:           "dotenv",
			envFileContent: "FOO=bar\n",
			fileKey:        "FOO",
			wantFileVal:    "bar",
			wantConfigUsed: true,
		},
		{
			name: "none",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			viper.Reset()

			tmpDir := t.TempDir()
			wd, err := os.Getwd()
			if err != nil {
				t.Fatalf("getwd: %v", err)
			}
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("chdir: %v", err)
			}
			t.Cleanup(func() {
				_ = os.Chdir(wd)
			})

			if tc.envKey != "" {
				t.Setenv(tc.envKey, tc.envVal)
			}

			if tc.envFileContent != "" {
				if err := os.WriteFile(".env", []byte(tc.envFileContent), 0o600); err != nil {
					t.Fatalf("write .env: %v", err)
				}
			}

			InitConfig()

			if tc.envLookup != "" {
				if got := viper.GetString(tc.envLookup); got != tc.wantEnvVal {
					t.Fatalf("env lookup %q: got %q want %q", tc.envLookup, got, tc.wantEnvVal)
				}
			}

			if tc.fileKey != "" {
				if got := viper.GetString(tc.fileKey); got != tc.wantFileVal {
					t.Fatalf("file key %q: got %q want %q", tc.fileKey, got, tc.wantFileVal)
				}
			}

			if tc.wantConfigUsed {
				if filepath.Base(viper.ConfigFileUsed()) != ".env" {
					t.Fatalf("config file used: got %q want %q", viper.ConfigFileUsed(), ".env")
				}
			} else {
				if viper.ConfigFileUsed() != "" {
					t.Fatalf("config file used: got %q want %q", viper.ConfigFileUsed(), "")
				}
			}
		})
	}
}
