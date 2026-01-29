package repository

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pandusatrianura/code-with-umam-second-meeting/internal/products/entity"
	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/database"
)

type testQuery struct {
	columns  []string
	rows     [][]driver.Value
	queryErr error
}

type testConfig struct {
	prepareErr  map[string]error
	execErr     map[string]error
	query       map[string]testQuery
	beginErr    error
	commitErr   error
	rollbackErr error
}

func (c *testConfig) getPrepareErr(query string) error {
	if c == nil || c.prepareErr == nil {
		return nil
	}
	return c.prepareErr[query]
}

func (c *testConfig) getExecErr(query string) error {
	if c == nil || c.execErr == nil {
		return nil
	}
	return c.execErr[query]
}

func (c *testConfig) getQuery(query string) testQuery {
	if c == nil || c.query == nil {
		return testQuery{}
	}
	return c.query[query]
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
	if err := c.cfg.getPrepareErr(query); err != nil {
		return nil, err
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

type testStmt struct {
	cfg   *testConfig
	query string
}

func (s *testStmt) Close() error  { return nil }
func (s *testStmt) NumInput() int { return -1 }

func (s *testStmt) Exec(args []driver.Value) (driver.Result, error) {
	if err := s.cfg.getExecErr(s.query); err != nil {
		return nil, err
	}
	return driver.RowsAffected(1), nil
}

func (s *testStmt) Query(args []driver.Value) (driver.Rows, error) {
	q := s.cfg.getQuery(s.query)
	if q.queryErr != nil {
		return nil, q.queryErr
	}
	return &testRows{columns: q.columns, values: q.rows}, nil
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

func (t *testTx) Rollback() error {
	if t.cfg.rollbackErr != nil {
		return t.cfg.rollbackErr
	}
	return nil
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
		} else {
			dest[i] = nil
		}
	}
	r.idx++
	return nil
}

var driverCounter int64

func newTestDB(t *testing.T, cfg *testConfig) *database.DB {
	t.Helper()
	name := fmt.Sprintf("repo_test_driver_%d", atomic.AddInt64(&driverCounter, 1))
	sql.Register(name, &testDriver{cfg: cfg})
	db, err := database.Open(name, "")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func TestNewProductRepository(t *testing.T) {
	db := newTestDB(t, &testConfig{})
	repo := NewProductRepository(db)
	if repo == nil {
		t.Fatalf("expected repository")
	}
	r, ok := repo.(*productRepository)
	if !ok {
		t.Fatalf("expected productRepository")
	}
	if r.db != db {
		t.Fatalf("expected db to match")
	}
}

func TestProductRepositoryCreateProduct(t *testing.T) {
	query := "INSERT INTO products (name, price, stock, category_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)"
	product := &entity.Product{Name: "p1", Price: 10, Stock: 2, CategoryID: 3}
	errPrepare := errors.New("prepare")
	errExec := errors.New("exec")
	errBegin := errors.New("begin")
	errCommit := errors.New("commit")

	tests := []struct {
		name    string
		cfg     *testConfig
		wantErr error
	}{
		{name: "ok", cfg: &testConfig{}},
		{name: "prepare", cfg: &testConfig{prepareErr: map[string]error{query: errPrepare}}, wantErr: errPrepare},
		{name: "exec", cfg: &testConfig{execErr: map[string]error{query: errExec}}, wantErr: errExec},
		{name: "begin", cfg: &testConfig{beginErr: errBegin}, wantErr: errBegin},
		{name: "commit", cfg: &testConfig{commitErr: errCommit}, wantErr: errCommit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t, tt.cfg)
			repo := NewProductRepository(db)
			err := repo.CreateProduct(product)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				return
			}
			if err == nil || !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestProductRepositoryUpdateProduct(t *testing.T) {
	query := "UPDATE products SET name = $1, price = $2, stock = $3, category_id = $4, updated_at = $5 WHERE id = $6"
	product := &entity.Product{Name: "p2", Price: 20, Stock: 5, CategoryID: 4}
	errExec := errors.New("exec")
	errCommit := errors.New("commit")

	tests := []struct {
		name    string
		cfg     *testConfig
		wantErr error
	}{
		{name: "ok", cfg: &testConfig{}},
		{name: "exec", cfg: &testConfig{execErr: map[string]error{query: errExec}}, wantErr: errExec},
		{name: "commit", cfg: &testConfig{commitErr: errCommit}, wantErr: errCommit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t, tt.cfg)
			repo := NewProductRepository(db)
			err := repo.UpdateProduct(9, product)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				return
			}
			if err == nil || !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestProductRepositoryDeleteProduct(t *testing.T) {
	query := "DELETE FROM products WHERE id = $1"
	errPrepare := errors.New("prepare")
	errBegin := errors.New("begin")

	tests := []struct {
		name    string
		cfg     *testConfig
		wantErr error
	}{
		{name: "ok", cfg: &testConfig{}},
		{name: "prepare", cfg: &testConfig{prepareErr: map[string]error{query: errPrepare}}, wantErr: errPrepare},
		{name: "begin", cfg: &testConfig{beginErr: errBegin}, wantErr: errBegin},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t, tt.cfg)
			repo := NewProductRepository(db)
			err := repo.DeleteProduct(1)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				return
			}
			if err == nil || !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestProductRepositoryGetAllProducts(t *testing.T) {
	query := "SELECT products.id, products.name, products.price, products.stock, products.created_at, products.updated_at, categories.id as category_id, categories.name as category_name FROM products JOIN categories ON products.category_id = categories.id"
	errQuery := errors.New("query")
	time1 := "2023-01-02T03:04:05Z"
	time2 := "2023-02-02T03:04:05Z"
	loc, _ := time.LoadLocation("Asia/Jakarta")
	parsed1, _ := time.Parse(time.RFC3339, time1)
	parsed2, _ := time.Parse(time.RFC3339, time2)

	tests := []struct {
		name      string
		cfg       *testConfig
		wantErr   error
		wantCount int
		wantFirst *entity.ResponseProductWithCategories
	}{
		{
			name: "ok",
			cfg: &testConfig{query: map[string]testQuery{
				query: {
					columns: []string{"id", "name", "price", "stock", "created_at", "updated_at", "category_id", "category_name"},
					rows: [][]driver.Value{
						{int64(1), "p1", int64(10), int64(2), time1, time2, int64(7), "c1"},
						{int64(2), "p2", int64(20), int64(3), time2, time1, int64(8), "c2"},
					},
				},
			}},
			wantCount: 2,
			wantFirst: &entity.ResponseProductWithCategories{
				ID:           1,
				Name:         "p1",
				Price:        10,
				Stock:        2,
				CategoryID:   7,
				CategoryName: "c1",
				CreatedAt:    parsed1.In(loc),
				UpdatedAt:    parsed2.In(loc),
			},
		},
		{
			name:      "empty",
			cfg:       &testConfig{query: map[string]testQuery{query: {columns: []string{"id"}}}},
			wantCount: 0,
		},
		{
			name:    "query",
			cfg:     &testConfig{query: map[string]testQuery{query: {queryErr: errQuery}}},
			wantErr: errQuery,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t, tt.cfg)
			repo := NewProductRepository(db)
			got, err := repo.GetAllProducts()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if len(got) != tt.wantCount {
					t.Fatalf("expected %d products, got %d", tt.wantCount, len(got))
				}
				if tt.wantFirst != nil && len(got) > 0 {
					if got[0].ID != tt.wantFirst.ID || got[0].Name != tt.wantFirst.Name || got[0].Price != tt.wantFirst.Price || got[0].Stock != tt.wantFirst.Stock || got[0].CategoryID != tt.wantFirst.CategoryID || got[0].CategoryName != tt.wantFirst.CategoryName {
						t.Fatalf("unexpected first product: %+v", got[0])
					}
					if !got[0].CreatedAt.Equal(tt.wantFirst.CreatedAt) || !got[0].UpdatedAt.Equal(tt.wantFirst.UpdatedAt) {
						t.Fatalf("unexpected times: %v %v", got[0].CreatedAt, got[0].UpdatedAt)
					}
					if got[0].CreatedAt.Location().String() != loc.String() || got[0].UpdatedAt.Location().String() != loc.String() {
						t.Fatalf("unexpected location: %s %s", got[0].CreatedAt.Location(), got[0].UpdatedAt.Location())
					}
				}
				return
			}
			if err == nil || !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestProductRepositoryGetProductByID(t *testing.T) {
	query := "SELECT products.id, products.name, products.price, products.stock, products.created_at, products.updated_at, categories.id as category_id, categories.name as category_name FROM products JOIN categories ON products.category_id = categories.id WHERE products.id = $1"
	errQuery := errors.New("query")
	time1 := "2023-01-02T03:04:05Z"
	time2 := "2023-02-02T03:04:05Z"
	loc, _ := time.LoadLocation("Asia/Jakarta")
	parsed1, _ := time.Parse(time.RFC3339, time1)
	parsed2, _ := time.Parse(time.RFC3339, time2)

	tests := []struct {
		name    string
		cfg     *testConfig
		wantErr string
		want    *entity.ResponseProductWithCategories
	}{
		{
			name: "ok",
			cfg: &testConfig{query: map[string]testQuery{
				query: {
					columns: []string{"id", "name", "price", "stock", "created_at", "updated_at", "category_id", "category_name"},
					rows:    [][]driver.Value{{int64(1), "p1", int64(10), int64(2), time1, time2, int64(7), "c1"}},
				},
			}},
			want: &entity.ResponseProductWithCategories{
				ID:           1,
				Name:         "p1",
				Price:        10,
				Stock:        2,
				CategoryID:   7,
				CategoryName: "c1",
				CreatedAt:    parsed1.In(loc),
				UpdatedAt:    parsed2.In(loc),
			},
		},
		{
			name:    "missing",
			cfg:     &testConfig{query: map[string]testQuery{query: {columns: []string{"id"}}}},
			wantErr: "product not found",
		},
		{
			name:    "query",
			cfg:     &testConfig{query: map[string]testQuery{query: {queryErr: errQuery}}},
			wantErr: errQuery.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t, tt.cfg)
			repo := NewProductRepository(db)
			got, err := repo.GetProductByID(1)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if got == nil {
					t.Fatalf("expected product")
				}
				if got.ID != tt.want.ID || got.Name != tt.want.Name || got.Price != tt.want.Price || got.Stock != tt.want.Stock || got.CategoryID != tt.want.CategoryID || got.CategoryName != tt.want.CategoryName {
					t.Fatalf("unexpected product: %+v", got)
				}
				if !got.CreatedAt.Equal(tt.want.CreatedAt) || !got.UpdatedAt.Equal(tt.want.UpdatedAt) {
					t.Fatalf("unexpected times: %v %v", got.CreatedAt, got.UpdatedAt)
				}
				if got.CreatedAt.Location().String() != loc.String() || got.UpdatedAt.Location().String() != loc.String() {
					t.Fatalf("unexpected location: %s %s", got.CreatedAt.Location(), got.UpdatedAt.Location())
				}
				return
			}
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("expected error %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestProductRepositoryGetCategoryByID(t *testing.T) {
	query := "SELECT id, name FROM categories WHERE id = $1"
	errQuery := errors.New("query")

	tests := []struct {
		name    string
		cfg     *testConfig
		wantErr string
		want    *entity.Category
	}{
		{
			name: "ok",
			cfg: &testConfig{query: map[string]testQuery{
				query: {
					columns: []string{"id", "name"},
					rows:    [][]driver.Value{{int64(1), "c1"}},
				},
			}},
			want: &entity.Category{ID: 1, Name: "c1"},
		},
		{
			name:    "missing",
			cfg:     &testConfig{query: map[string]testQuery{query: {columns: []string{"id"}}}},
			wantErr: "category not found",
		},
		{
			name:    "query",
			cfg:     &testConfig{query: map[string]testQuery{query: {queryErr: errQuery}}},
			wantErr: errQuery.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t, tt.cfg)
			repo := NewProductRepository(db)
			got, err := repo.GetCategoryByID(1)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if got == nil {
					t.Fatalf("expected category")
				}
				if got.ID != tt.want.ID || got.Name != tt.want.Name {
					t.Fatalf("unexpected category: %+v", got)
				}
				return
			}
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("expected error %q, got %v", tt.wantErr, err)
			}
		})
	}
}

