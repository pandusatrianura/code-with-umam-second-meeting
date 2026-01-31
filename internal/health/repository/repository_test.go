package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/database"
)

type pingDriver struct {
	pingErr error
}

type pingConn struct {
	pingErr error
}

func (d *pingDriver) Open(name string) (driver.Conn, error) {
	return &pingConn{pingErr: d.pingErr}, nil
}

func (c *pingConn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}

func (c *pingConn) Close() error {
	return nil
}

func (c *pingConn) Begin() (driver.Tx, error) {
	return nil, errors.New("not implemented")
}

func (c *pingConn) Ping(ctx context.Context) error {
	return c.pingErr
}

var driverCounter uint64

func newTestDB(t *testing.T, pingErr error) *database.DB {
	t.Helper()
	name := fmt.Sprintf("ping-driver-%d", atomic.AddUint64(&driverCounter, 1))
	sql.Register(name, &pingDriver{pingErr: pingErr})
	sqlDB, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	return &database.DB{DB: sqlDB, Logging: true}
}

func TestNewHealthRepository(t *testing.T) {
	cases := []struct {
		name string
		db   *database.DB
	}{
		{name: "nil", db: nil},
		{name: "set", db: &database.DB{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewHealthRepository(tc.db)
			if repo == nil {
				t.Fatalf("expected repo")
			}
			hr, ok := repo.(*healthRepository)
			if !ok {
				t.Fatalf("expected *healthRepository")
			}
			if hr.db != tc.db {
				t.Fatalf("expected db to match")
			}
		})
	}
}

func TestHealthRepositoryDB(t *testing.T) {
	cases := []struct {
		name    string
		pingErr error
		wantErr bool
	}{
		{name: "ok", pingErr: nil, wantErr: false},
		{name: "err", pingErr: errors.New("ping failed"), wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &healthRepository{db: newTestDB(t, tc.pingErr)}
			err := repo.DB()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
