package database

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/spf13/viper"
)

const initDatabaseCaseEnv = "INITDATABASE_CASE"

func TestInitDatabase(t *testing.T) {
	if os.Getenv(initDatabaseCaseEnv) != "" {
		return
	}

	tests := []struct {
		name string
	}{
		{name: "open"},
		{name: "ping"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run", "^TestInitDatabaseChild$")
			cmd.Env = append(os.Environ(), initDatabaseCaseEnv+"="+tt.name)
			err := cmd.Run()
			if err == nil {
				t.Fatalf("expected non-zero exit")
			}
			if _, ok := err.(*exec.ExitError); !ok {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestInitDatabaseChild(t *testing.T) {
	caseName := os.Getenv(initDatabaseCaseEnv)
	if caseName == "" {
		t.Skip("helper")
	}

	viper.Reset()

	switch caseName {
	case "open":
		viper.Set("DATABASE_USER", "user")
		viper.Set("DATABASE_PASSWORD", "pass")
		viper.Set("DATABASE_HOST", "bad host")
		viper.Set("DATABASE_PORT", 5432)
		viper.Set("DATABASE_NAME", "db")
		viper.Set("DATABASE_MAX_LIFETIME_CONNECTION", time.Second)
		viper.Set("DATABASE_MAX_IDLE_CONNECTION", 1)
		viper.Set("DATABASE_MAX_OPEN_CONNECTION", 1)
	case "ping":
		viper.Set("DATABASE_USER", "user")
		viper.Set("DATABASE_PASSWORD", "pass")
		viper.Set("DATABASE_HOST", "127.0.0.1")
		viper.Set("DATABASE_PORT", 1)
		viper.Set("DATABASE_NAME", "db")
		viper.Set("DATABASE_MAX_LIFETIME_CONNECTION", time.Second)
		viper.Set("DATABASE_MAX_IDLE_CONNECTION", 1)
		viper.Set("DATABASE_MAX_OPEN_CONNECTION", 1)
	default:
		t.Fatalf("unknown case: %s", caseName)
	}

	_, _ = InitDatabase()
	t.Fatalf("expected fatal exit")
}
