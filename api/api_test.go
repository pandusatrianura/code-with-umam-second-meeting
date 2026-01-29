package api

import (
	"testing"

	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/database"
)

func TestNewAPIServer(t *testing.T) {
	tests := []struct {
		name string
		addr string
		db   *database.DB
	}{
		{name: "empty", addr: "", db: nil},
		{name: "with-db", addr: "127.0.0.1:8080", db: &database.DB{}},
		{name: "addr-only", addr: "localhost:3000", db: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewAPIServer(tt.addr, tt.db)
			if s == nil {
				t.Fatal("expected server, got nil")
			}
			if s.addr != tt.addr {
				t.Fatalf("expected addr %q, got %q", tt.addr, s.addr)
			}
			if s.db != tt.db {
				t.Fatalf("expected db %v, got %v", tt.db, s.db)
			}
		})
	}
}

func TestServerRun(t *testing.T) {
	tests := []struct {
		name string
		addr string
	}{
		{name: "missing-port", addr: "127.0.0.1"},
		{name: "invalid", addr: "bad::addr"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{addr: tt.addr}
			if err := s.Run(); err == nil {
				t.Fatalf("expected error for addr %q", tt.addr)
			}
		})
	}
}
