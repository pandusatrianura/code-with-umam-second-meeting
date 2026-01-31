package api

import (
	"io"
	"log"
	"testing"

	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/database"
)

func TestNewAPIServer(t *testing.T) {
	db := &database.DB{}
	tests := []struct {
		name string
		addr string
		db   *database.DB
	}{
		{name: "addr", addr: ":8080", db: db},
		{name: "empty", addr: "", db: db},
		{name: "nil-db", addr: ":9090", db: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := NewAPIServer(tt.addr, tt.db)
			if srv == nil {
				t.Fatal("expected server")
			}
			if srv.addr != tt.addr {
				t.Fatalf("addr = %q, want %q", srv.addr, tt.addr)
			}
			if srv.db != tt.db {
				t.Fatal("db mismatch")
			}
		})
	}
}

func TestServerRun(t *testing.T) {
	oldWriter := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(oldWriter)

	tests := []struct {
		name string
		addr string
	}{
		{name: "bad", addr: "bad"},
		{name: "scheme", addr: "http://127.0.0.1:8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{addr: tt.addr, db: nil}
			if err := srv.Run(); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
