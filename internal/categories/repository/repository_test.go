package repository

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pandusatrianura/code-with-umam-second-meeting/internal/categories/entity"
	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/database"
)

type testQuery struct {
	columns  []string
	rows     [][]driver.Value
	queryErr error
}

type testConfig struct {
	prepareErr error
	execErr    error
	query      testQuery
	beginErr   error
	commitErr  error

	mu            sync.Mutex
	lastExecArgs  []driver.Value
	lastQueryArgs []driver.Value
}

func (c *testConfig) setLastExecArgs(args []driver.Value) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastExecArgs = append([]driver.Value(nil), args...)
}

func (c *testConfig) setLastQueryArgs(args []driver.Value) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastQueryArgs = append([]driver.Value(nil), args...)
}

func (c *testConfig) getLastExecArgs() []driver.Value {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]driver.Value(nil), c.lastExecArgs...)
}

func (c *testConfig) getLastQueryArgs() []driver.Value {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]driver.Value(nil), c.lastQueryArgs...)
}

type testDriver struct {
	cfg *testConfig
}

func (d *testDriver) Open(name string) (driver.Conn, error) {
	return &testConn{cfg: d.cfg}, nil
}

type testConn struct {
	cfg *testConfig
}

func (c *testConn) Prepare(query string) (driver.Stmt, error) {
	if c.cfg.prepareErr != nil {
		return nil, c.cfg.prepareErr
	}
	return &testStmt{cfg: c.cfg, query: query}, nil
}

func (c *testConn) Close() error { return nil }

func (c *testConn) Begin() (driver.Tx, error) {
	if c.cfg.beginErr != nil {
		return nil, c.cfg.beginErr
	}
	return &testTx{cfg: c.cfg}, nil
}

type testTx struct {
	cfg *testConfig
}

func (t *testTx) Commit() error {
	if t.cfg.commitErr != nil {
		return t.cfg.commitErr
	}
	return nil
}

func (t *testTx) Rollback() error { return nil }

type testStmt struct {
	cfg   *testConfig
	query string
}

func (s *testStmt) Close() error  { return nil }
func (s *testStmt) NumInput() int { return -1 }

func (s *testStmt) Exec(args []driver.Value) (driver.Result, error) {
	if s.cfg.execErr != nil {
		return nil, s.cfg.execErr
	}
	s.cfg.setLastExecArgs(args)
	return driver.RowsAffected(1), nil
}

func (s *testStmt) Query(args []driver.Value) (driver.Rows, error) {
	if s.cfg.query.queryErr != nil {
		return nil, s.cfg.query.queryErr
	}
	s.cfg.setLastQueryArgs(args)
	return &testRows{columns: s.cfg.query.columns, values: s.cfg.query.rows}, nil
}

type testRows struct {
	columns []string
	values  [][]driver.Value
	idx     int
}

func (r *testRows) Columns() []string { return r.columns }
func (r *testRows) Close() error      { return nil }

func (r *testRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.values) {
		return io.EOF
	}
	row := r.values[r.idx]
	for i := range dest {
		if i < len(row) {
			dest[i] = row[i]
		}
	}
	r.idx++
	return nil
}

var driverCounter int64

func newTestDB(t *testing.T, cfg *testConfig) *database.DB {
	t.Helper()
	name := fmt.Sprintf("category_repo_driver_%d", atomic.AddInt64(&driverCounter, 1))
	sql.Register(name, &testDriver{cfg: cfg})
	sqlDB, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return &database.DB{DB: sqlDB, Logging: true}
}

func mustParseTime(t *testing.T, value string) time.Time {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time: %v", err)
	}
	return parsed.In(loc)
}

func TestNewCategoryRepository(t *testing.T) {
	db := newTestDB(t, &testConfig{})
	repo := NewCategoryRepository(db)
	if repo == nil {
		t.Fatalf("expected repository")
	}
	cr, ok := repo.(*categoryRepository)
	if !ok {
		t.Fatalf("expected *categoryRepository")
	}
	if cr.db != db {
		t.Fatalf("expected db to match")
	}
}

func TestCategoryRepository_CreateCategory(t *testing.T) {
	tests := []struct {
		name      string
		cfg       testConfig
		category  entity.Category
		wantErr   error
		wantArgs  []driver.Value
		checkArgs bool
	}{
		{
			name:      "ok",
			category:  entity.Category{Name: "food", Description: "fresh"},
			wantArgs:  []driver.Value{"food", "fresh", "now()", "now()"},
			checkArgs: true,
		},
		{
			name:     "begin",
			cfg:      testConfig{beginErr: errors.New("begin")},
			category: entity.Category{Name: "food", Description: "fresh"},
			wantErr:  errors.New("begin"),
		},
		{
			name:     "prepare",
			cfg:      testConfig{prepareErr: errors.New("prepare")},
			category: entity.Category{Name: "food", Description: "fresh"},
			wantErr:  errors.New("prepare"),
		},
		{
			name:     "exec",
			cfg:      testConfig{execErr: errors.New("exec")},
			category: entity.Category{Name: "food", Description: "fresh"},
			wantErr:  errors.New("exec"),
		},
		{
			name:      "commit",
			cfg:       testConfig{commitErr: errors.New("commit")},
			category:  entity.Category{Name: "food", Description: "fresh"},
			wantErr:   errors.New("commit"),
			wantArgs:  []driver.Value{"food", "fresh", "now()", "now()"},
			checkArgs: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.cfg
			db := newTestDB(t, &cfg)
			repo := NewCategoryRepository(db)
			err := repo.CreateCategory(&tt.category)
			if (err == nil) != (tt.wantErr == nil) {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr != nil && err != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.checkArgs {
				if got := cfg.getLastExecArgs(); !reflect.DeepEqual(got, tt.wantArgs) {
					t.Fatalf("expected args %v, got %v", tt.wantArgs, got)
				}
			}
		})
	}
}

func TestCategoryRepository_UpdateCategory(t *testing.T) {
	tests := []struct {
		name      string
		cfg       testConfig
		id        int64
		category  entity.Category
		wantErr   error
		wantArgs  []driver.Value
		checkArgs bool
	}{
		{
			name:      "ok",
			id:        9,
			category:  entity.Category{Name: "tech", Description: "gadgets"},
			wantArgs:  []driver.Value{"tech", "gadgets", "now()", int64(9)},
			checkArgs: true,
		},
		{
			name:     "begin",
			cfg:      testConfig{beginErr: errors.New("begin")},
			id:       9,
			category: entity.Category{Name: "tech", Description: "gadgets"},
			wantErr:  errors.New("begin"),
		},
		{
			name:     "prepare",
			cfg:      testConfig{prepareErr: errors.New("prepare")},
			id:       9,
			category: entity.Category{Name: "tech", Description: "gadgets"},
			wantErr:  errors.New("prepare"),
		},
		{
			name:     "exec",
			cfg:      testConfig{execErr: errors.New("exec")},
			id:       9,
			category: entity.Category{Name: "tech", Description: "gadgets"},
			wantErr:  errors.New("exec"),
		},
		{
			name:      "commit",
			cfg:       testConfig{commitErr: errors.New("commit")},
			id:        9,
			category:  entity.Category{Name: "tech", Description: "gadgets"},
			wantErr:   errors.New("commit"),
			wantArgs:  []driver.Value{"tech", "gadgets", "now()", int64(9)},
			checkArgs: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.cfg
			db := newTestDB(t, &cfg)
			repo := NewCategoryRepository(db)
			err := repo.UpdateCategory(tt.id, &tt.category)
			if (err == nil) != (tt.wantErr == nil) {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr != nil && err != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.checkArgs {
				if got := cfg.getLastExecArgs(); !reflect.DeepEqual(got, tt.wantArgs) {
					t.Fatalf("expected args %v, got %v", tt.wantArgs, got)
				}
			}
		})
	}
}

func TestCategoryRepository_DeleteCategory(t *testing.T) {
	tests := []struct {
		name      string
		cfg       testConfig
		id        int64
		wantErr   error
		wantArgs  []driver.Value
		checkArgs bool
	}{
		{
			name:      "ok",
			id:        4,
			wantArgs:  []driver.Value{int64(4)},
			checkArgs: true,
		},
		{
			name:    "begin",
			cfg:     testConfig{beginErr: errors.New("begin")},
			id:      4,
			wantErr: errors.New("begin"),
		},
		{
			name:    "prepare",
			cfg:     testConfig{prepareErr: errors.New("prepare")},
			id:      4,
			wantErr: errors.New("prepare"),
		},
		{
			name:    "exec",
			cfg:     testConfig{execErr: errors.New("exec")},
			id:      4,
			wantErr: errors.New("exec"),
		},
		{
			name:      "commit",
			cfg:       testConfig{commitErr: errors.New("commit")},
			id:        4,
			wantErr:   errors.New("commit"),
			wantArgs:  []driver.Value{int64(4)},
			checkArgs: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.cfg
			db := newTestDB(t, &cfg)
			repo := NewCategoryRepository(db)
			err := repo.DeleteCategory(tt.id)
			if (err == nil) != (tt.wantErr == nil) {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr != nil && err != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.checkArgs {
				if got := cfg.getLastExecArgs(); !reflect.DeepEqual(got, tt.wantArgs) {
					t.Fatalf("expected args %v, got %v", tt.wantArgs, got)
				}
			}
		})
	}
}

func TestCategoryRepository_GetCategoryByID(t *testing.T) {
	created := "2024-01-02T03:04:05Z"
	updated := "2024-01-03T04:05:06Z"
	tests := []struct {
		name      string
		cfg       testConfig
		id        int64
		wantErr   error
		want      *entity.ResponseCategory
		wantArgs  []driver.Value
		checkArgs bool
	}{
		{
			name: "ok",
			cfg: testConfig{query: testQuery{
				columns: []string{"id", "name", "description", "created_at", "updated_at"},
				rows: [][]driver.Value{{
					int64(2), "book", "paper", created, updated,
				}},
			}},
			id: 2,
			want: &entity.ResponseCategory{
				ID:          2,
				Name:        "book",
				Description: "paper",
				CreatedAt:   mustParseTime(t, created),
				UpdatedAt:   mustParseTime(t, updated),
			},
			wantArgs:  []driver.Value{int64(2)},
			checkArgs: true,
		},
		{
			name: "notfound",
			cfg: testConfig{query: testQuery{
				columns: []string{"id", "name", "description", "created_at", "updated_at"},
				rows:    [][]driver.Value{},
			}},
			id:      2,
			wantErr: errors.New("category not found"),
		},
		{
			name:    "queryerr",
			cfg:     testConfig{query: testQuery{queryErr: errors.New("query")}},
			id:      2,
			wantErr: errors.New("query"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.cfg
			db := newTestDB(t, &cfg)
			repo := NewCategoryRepository(db)
			got, err := repo.GetCategoryByID(tt.id)
			if (err == nil) != (tt.wantErr == nil) {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr != nil && err != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil {
				if got == nil {
					t.Fatalf("expected category")
				}
				if got.ID != tt.want.ID || got.Name != tt.want.Name || got.Description != tt.want.Description {
					t.Fatalf("expected %+v, got %+v", tt.want, got)
				}
				if !got.CreatedAt.Equal(tt.want.CreatedAt) || !got.UpdatedAt.Equal(tt.want.UpdatedAt) {
					t.Fatalf("expected times %+v, got %+v", tt.want, got)
				}
			}
			if tt.checkArgs {
				if gotArgs := cfg.getLastQueryArgs(); !reflect.DeepEqual(gotArgs, tt.wantArgs) {
					t.Fatalf("expected args %v, got %v", tt.wantArgs, gotArgs)
				}
			}
		})
	}
}

func TestCategoryRepository_GetAllCategories(t *testing.T) {
	created := "2024-01-02T03:04:05Z"
	updated := "2024-01-03T04:05:06Z"
	tests := []struct {
		name      string
		cfg       testConfig
		wantErr   error
		want      []entity.ResponseCategory
		wantArgs  []driver.Value
		checkArgs bool
	}{
		{
			name: "ok",
			cfg: testConfig{query: testQuery{
				columns: []string{"id", "name", "description", "created_at", "updated_at"},
				rows: [][]driver.Value{
					{int64(1), "a", "one", created, updated},
					{int64(2), "b", "two", created, updated},
				},
			}},
			want: []entity.ResponseCategory{
				{ID: 1, Name: "a", Description: "one", CreatedAt: mustParseTime(t, created), UpdatedAt: mustParseTime(t, updated)},
				{ID: 2, Name: "b", Description: "two", CreatedAt: mustParseTime(t, created), UpdatedAt: mustParseTime(t, updated)},
			},
			wantArgs:  []driver.Value{},
			checkArgs: false,
		},
		{
			name: "empty",
			cfg: testConfig{query: testQuery{
				columns: []string{"id", "name", "description", "created_at", "updated_at"},
				rows:    [][]driver.Value{},
			}},
			want: nil,
		},
		{
			name:    "queryerr",
			cfg:     testConfig{query: testQuery{queryErr: errors.New("query")}},
			wantErr: errors.New("query"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.cfg
			db := newTestDB(t, &cfg)
			repo := NewCategoryRepository(db)
			got, err := repo.GetAllCategories()
			if (err == nil) != (tt.wantErr == nil) {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr != nil && err != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil {
				if len(got) != len(tt.want) {
					t.Fatalf("expected len %d, got %d", len(tt.want), len(got))
				}
				for i := range got {
					if got[i].ID != tt.want[i].ID || got[i].Name != tt.want[i].Name || got[i].Description != tt.want[i].Description {
						t.Fatalf("expected %+v, got %+v", tt.want[i], got[i])
					}
					if !got[i].CreatedAt.Equal(tt.want[i].CreatedAt) || !got[i].UpdatedAt.Equal(tt.want[i].UpdatedAt) {
						t.Fatalf("expected times %+v, got %+v", tt.want[i], got[i])
					}
				}
			}
			if tt.checkArgs {
				if gotArgs := cfg.getLastQueryArgs(); !reflect.DeepEqual(gotArgs, tt.wantArgs) {
					t.Fatalf("expected args %v, got %v", tt.wantArgs, gotArgs)
				}
			}
		})
	}
}
