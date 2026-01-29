package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

type testQuery struct {
	columns  []string
	rows     [][]driver.Value
	queryErr error
	nextErr  error
	closeErr error
}

type testConfig struct {
	prepareErr  map[string]error
	query       map[string]testQuery
	beginErr    error
	commitErr   error
	rollbackErr error

	mu     sync.Mutex
	lastTx *testTx
}

func (c *testConfig) setLastTx(tx *testTx) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastTx = tx
}

func (c *testConfig) getLastTx() *testTx {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lastTx
}

func (c *testConfig) getPrepareErr(query string) error {
	if c.prepareErr == nil {
		return nil
	}
	return c.prepareErr[query]
}

func (c *testConfig) getQuery(query string) testQuery {
	if c.query == nil {
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
	tx := &testTx{cfg: c.cfg}
	c.cfg.setLastTx(tx)
	return tx, nil
}

func (c *testConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	q := c.cfg.getQuery(query)
	if q.queryErr != nil {
		return nil, q.queryErr
	}
	return &testRows{columns: q.columns, values: q.rows, nextErr: q.nextErr, closeErr: q.closeErr}, nil
}

func (c *testConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	vals := make([]driver.Value, len(args))
	for i, v := range args {
		vals[i] = v.Value
	}
	return c.Query(query, vals)
}

type testStmt struct {
	cfg   *testConfig
	query string
}

func (s *testStmt) Close() error  { return nil }
func (s *testStmt) NumInput() int { return -1 }
func (s *testStmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, errors.New("exec not supported")
}

func (s *testStmt) Query(args []driver.Value) (driver.Rows, error) {
	q := s.cfg.getQuery(s.query)
	if q.queryErr != nil {
		return nil, q.queryErr
	}
	return &testRows{columns: q.columns, values: q.rows, nextErr: q.nextErr, closeErr: q.closeErr}, nil
}

func (s *testStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	vals := make([]driver.Value, len(args))
	for i, v := range args {
		vals[i] = v.Value
	}
	return s.Query(vals)
}

type testTx struct {
	cfg       *testConfig
	committed bool
	rolled    bool
}

func (t *testTx) Commit() error {
	t.committed = true
	if t.cfg.commitErr != nil {
		return t.cfg.commitErr
	}
	return nil
}

func (t *testTx) Rollback() error {
	t.rolled = true
	if t.cfg.rollbackErr != nil {
		return t.cfg.rollbackErr
	}
	return nil
}

type testRows struct {
	columns  []string
	values   [][]driver.Value
	idx      int
	nextErr  error
	closeErr error
}

func (r *testRows) Columns() []string { return r.columns }

func (r *testRows) Close() error { return r.closeErr }

func (r *testRows) Next(dest []driver.Value) error {
	if r.nextErr != nil && r.idx == 0 {
		return r.nextErr
	}
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

func newTestDB(t *testing.T, cfg *testConfig) *DB {
	t.Helper()
	name := fmt.Sprintf("testdriver_%d", atomic.AddInt64(&driverCounter, 1))
	sql.Register(name, &testDriver{cfg: cfg})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return &DB{DB: db, Logging: true}
}

func TestOpen(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		cfg := &testConfig{}
		name := fmt.Sprintf("testdriver_open_%d", atomic.AddInt64(&driverCounter, 1))
		sql.Register(name, &testDriver{cfg: cfg})
		db, err := Open(name, "")
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if db == nil || db.DB == nil {
			t.Fatalf("expected db instance")
		}
		_ = db.Close()
	})

	t.Run("unknown", func(t *testing.T) {
		db, err := Open("missing_driver", "")
		if err == nil {
			t.Fatalf("expected error")
		}
		if db == nil {
			t.Fatalf("expected wrapper db")
		}
		if db.DB != nil {
			t.Fatalf("expected nil sql.DB")
		}
	})
}

func TestDBWithStmt(t *testing.T) {
	origLogFn := LogFn
	t.Cleanup(func() { LogFn = origLogFn })

	tests := []struct {
		name       string
		prepareErr error
		fnErr      error
		wantErr    error
		wantLog    bool
		wantLogSub string
	}{
		{name: "ok", wantLog: true},
		{name: "fnerr", fnErr: errors.New("fn"), wantErr: errors.New("fn"), wantLog: true, wantLogSub: "fn"},
		{name: "prepare", prepareErr: errors.New("prep"), wantErr: errors.New("prep")},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			query := "select 1"
			cfg := &testConfig{prepareErr: map[string]error{query: tt.prepareErr}}
			db := newTestDB(t, cfg)
			var called int
			var last string
			LogFn = func(format string, args ...interface{}) {
				called++
				last = fmt.Sprintf(format, args...)
			}
			err := db.WithStmt(query, func(stmt *Stmt) error {
				return tt.fnErr
			})
			if (err == nil) != (tt.wantErr == nil) {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr != nil && err != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantLog && called != 1 {
				t.Fatalf("expected log call")
			}
			if !tt.wantLog && called != 0 {
				t.Fatalf("did not expect log call")
			}
			if tt.wantLog && !strings.Contains(last, query) {
				t.Fatalf("expected log to contain query")
			}
			if tt.wantLogSub != "" && !strings.Contains(last, tt.wantLogSub) {
				t.Fatalf("expected log to contain error")
			}
		})
	}
}

func TestDBWithTx(t *testing.T) {
	tests := []struct {
		name         string
		beginErr     error
		fnErr        error
		commitErr    error
		wantErr      error
		wantRollback bool
		wantCommit   bool
	}{
		{name: "ok", wantCommit: true},
		{name: "begin", beginErr: errors.New("begin"), wantErr: errors.New("begin")},
		{name: "fn", fnErr: errors.New("fn"), wantErr: errors.New("fn"), wantRollback: true},
		{name: "commit", commitErr: errors.New("commit"), wantErr: errors.New("commit"), wantCommit: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cfg := &testConfig{beginErr: tt.beginErr, commitErr: tt.commitErr}
			db := newTestDB(t, cfg)
			err := db.WithTx(func(tx *Tx) error {
				return tt.fnErr
			})
			if (err == nil) != (tt.wantErr == nil) {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr != nil && err != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			tx := cfg.getLastTx()
			if tt.wantRollback && (tx == nil || !tx.rolled) {
				t.Fatalf("expected rollback")
			}
			if tt.wantCommit && (tx == nil || !tx.committed) {
				t.Fatalf("expected commit")
			}
		})
	}
}

func TestTxWithStmt(t *testing.T) {
	origLogFn := LogFn
	t.Cleanup(func() { LogFn = origLogFn })

	tests := []struct {
		name       string
		prepareErr error
		fnErr      error
		wantErr    error
		wantLog    bool
		wantLogSub string
	}{
		{name: "ok", wantLog: true},
		{name: "fnerr", fnErr: errors.New("fn"), wantErr: errors.New("fn"), wantLog: true, wantLogSub: "fn"},
		{name: "prepare", prepareErr: errors.New("prep"), wantErr: errors.New("prep")},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			query := "select 1"
			cfg := &testConfig{prepareErr: map[string]error{query: tt.prepareErr}}
			db := newTestDB(t, cfg)
			sqlTx, err := db.Begin()
			if err != nil {
				t.Fatalf("begin: %v", err)
			}
			defer func() { _ = sqlTx.Rollback() }()
			tx := &Tx{Tx: sqlTx}

			var called int
			var last string
			LogFn = func(format string, args ...interface{}) {
				called++
				last = fmt.Sprintf(format, args...)
			}
			err = tx.WithStmt(query, func(stmt *Stmt) error {
				return tt.fnErr
			})
			if (err == nil) != (tt.wantErr == nil) {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr != nil && err != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantLog && called != 1 {
				t.Fatalf("expected log call")
			}
			if !tt.wantLog && called != 0 {
				t.Fatalf("did not expect log call")
			}
			if tt.wantLog && !strings.Contains(last, "tx:") {
				t.Fatalf("expected tx log prefix")
			}
			if tt.wantLogSub != "" && !strings.Contains(last, tt.wantLogSub) {
				t.Fatalf("expected log to contain error")
			}
		})
	}
}

func TestStmtQuery(t *testing.T) {
	query := "select id"
	tests := []struct {
		name      string
		q         testQuery
		rowFnErr  error
		wantErr   error
		wantCount int
	}{
		{name: "ok", q: testQuery{columns: []string{"id"}, rows: [][]driver.Value{{1}, {2}}}, wantCount: 2},
		{name: "rowfn", q: testQuery{columns: []string{"id"}, rows: [][]driver.Value{{1}, {2}}}, rowFnErr: errors.New("row"), wantErr: errors.New("row"), wantCount: 1},
		{name: "queryerr", q: testQuery{queryErr: errors.New("query")}, wantErr: errors.New("query")},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cfg := &testConfig{query: map[string]testQuery{query: tt.q}}
			db := newTestDB(t, cfg)
			sqlStmt, err := db.Prepare(query)
			if err != nil {
				t.Fatalf("prepare: %v", err)
			}
			defer sqlStmt.Close()
			stmt := &Stmt{Stmt: sqlStmt}
			var count int
			err = stmt.Query(func(rows *Rows) error {
				var id int
				if err := rows.Scan(&id); err != nil {
					return err
				}
				count++
				if tt.rowFnErr != nil {
					return tt.rowFnErr
				}
				return nil
			})
			if (err == nil) != (tt.wantErr == nil) {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr != nil && err != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if count != tt.wantCount {
				t.Fatalf("expected count %d, got %d", tt.wantCount, count)
			}
		})
	}
}

func TestStmtQueryRow(t *testing.T) {
	query := "select id"
	tests := []struct {
		name    string
		q       testQuery
		wantErr error
		wantVal int
	}{
		{name: "ok", q: testQuery{columns: []string{"id"}, rows: [][]driver.Value{{3}}}, wantVal: 3},
		{name: "queryerr", q: testQuery{queryErr: errors.New("query")}, wantErr: errors.New("query")},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cfg := &testConfig{query: map[string]testQuery{query: tt.q}}
			db := newTestDB(t, cfg)
			sqlStmt, err := db.Prepare(query)
			if err != nil {
				t.Fatalf("prepare: %v", err)
			}
			defer sqlStmt.Close()
			stmt := &Stmt{Stmt: sqlStmt}
			row := stmt.QueryRow()
			var id int
			err = row.Scan(&id)
			if (err == nil) != (tt.wantErr == nil) {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr != nil && err != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && id != tt.wantVal {
				t.Fatalf("expected %d, got %d", tt.wantVal, id)
			}
		})
	}
}

func TestDBQueryRow(t *testing.T) {
	query := "select id"
	tests := []struct {
		name    string
		q       testQuery
		wantErr error
		wantVal int
	}{
		{name: "ok", q: testQuery{columns: []string{"id"}, rows: [][]driver.Value{{7}}}, wantVal: 7},
		{name: "queryerr", q: testQuery{queryErr: errors.New("query")}, wantErr: errors.New("query")},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cfg := &testConfig{query: map[string]testQuery{query: tt.q}}
			db := newTestDB(t, cfg)
			row := db.QueryRow(query)
			var id int
			err := row.Scan(&id)
			if (err == nil) != (tt.wantErr == nil) {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr != nil && err != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && id != tt.wantVal {
				t.Fatalf("expected %d, got %d", tt.wantVal, id)
			}
		})
	}
}

func TestRowError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()
	var r Row
	_ = r.Error()
}

func TestRowScan(t *testing.T) {
	query := "select id"
	tests := []struct {
		name    string
		q       testQuery
		rowFn   func(db *DB, q string) *Row
		setup   func(row *Row)
		dest    func() ([]interface{}, *int, *sql.RawBytes)
		wantErr error
		wantVal int
	}{
		{
			name:  "ok",
			q:     testQuery{columns: []string{"id"}, rows: [][]driver.Value{{11}}},
			rowFn: func(db *DB, q string) *Row { return db.QueryRow(q) },
			dest: func() ([]interface{}, *int, *sql.RawBytes) {
				var id int
				return []interface{}{&id}, &id, nil
			},
			wantVal: 11,
		},
		{
			name:  "rawbytes",
			q:     testQuery{columns: []string{"id"}, rows: [][]driver.Value{{"x"}}},
			rowFn: func(db *DB, q string) *Row { return db.QueryRow(q) },
			dest: func() ([]interface{}, *int, *sql.RawBytes) {
				var rb sql.RawBytes
				return []interface{}{&rb}, nil, &rb
			},
			wantErr: errors.New("sql: RawBytes isn't allowed on Row.Scan"),
		},
		{
			name:  "norows",
			q:     testQuery{columns: []string{"id"}, rows: [][]driver.Value{}},
			rowFn: func(db *DB, q string) *Row { return db.QueryRow(q) },
			dest: func() ([]interface{}, *int, *sql.RawBytes) {
				var id int
				return []interface{}{&id}, &id, nil
			},
			wantErr: sql.ErrNoRows,
		},
		{
			name:  "rowserr",
			q:     testQuery{columns: []string{"id"}, nextErr: errors.New("next")},
			rowFn: func(db *DB, q string) *Row { return db.QueryRow(q) },
			dest: func() ([]interface{}, *int, *sql.RawBytes) {
				var id int
				return []interface{}{&id}, &id, nil
			},
			wantErr: errors.New("next"),
		},
		{
			name:  "closeerr",
			q:     testQuery{columns: []string{"id"}, rows: [][]driver.Value{{1}}, closeErr: errors.New("close")},
			rowFn: func(db *DB, q string) *Row { return db.QueryRow(q) },
			dest: func() ([]interface{}, *int, *sql.RawBytes) {
				var id int
				return []interface{}{&id}, &id, nil
			},
			wantErr: errors.New("close"),
		},
		{
			name:  "queryerr",
			q:     testQuery{queryErr: errors.New("query")},
			rowFn: func(db *DB, q string) *Row { return db.QueryRow(q) },
			dest: func() ([]interface{}, *int, *sql.RawBytes) {
				var id int
				return []interface{}{&id}, &id, nil
			},
			wantErr: errors.New("query"),
		},
		{
			name:  "closed",
			q:     testQuery{columns: []string{"id"}, rows: [][]driver.Value{{1}}},
			rowFn: func(db *DB, q string) *Row { return db.QueryRow(q) },
			setup: func(row *Row) {
				_ = row.rows.Close()
			},
			dest: func() ([]interface{}, *int, *sql.RawBytes) {
				var id int
				return []interface{}{&id}, &id, nil
			},
			wantErr: errors.New("sql: Rows are closed"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cfg := &testConfig{query: map[string]testQuery{query: tt.q}}
			db := newTestDB(t, cfg)
			rowFn := tt.rowFn
			if rowFn == nil {
				rowFn = func(db *DB, q string) *Row { return db.QueryRow(q) }
			}
			row := rowFn(db, query)
			if tt.setup != nil {
				tt.setup(row)
			}
			dest, idPtr, _ := tt.dest()
			err := row.Scan(dest...)
			if (err == nil) != (tt.wantErr == nil) {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr != nil && err != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && idPtr != nil && *idPtr != tt.wantVal {
				t.Fatalf("expected %d, got %d", tt.wantVal, *idPtr)
			}
		})
	}
}

func TestRowsScan(t *testing.T) {
	query := "select id, name"
	type item struct {
		ID   int    `sql:"id"`
		Name string `sql:"name"`
	}
	bad := struct {
		ID int `sql:"missing"`
	}{}

	tests := []struct {
		name     string
		q        testQuery
		dest     func() ([]interface{}, *item)
		setup    func(rows *sql.Rows)
		skipNext bool
		wantErr  error
		wantVal  item
	}{
		{
			name: "ok",
			q:    testQuery{columns: []string{"id", "name"}, rows: [][]driver.Value{{5, "a"}}},
			dest: func() ([]interface{}, *item) {
				var it item
				return []interface{}{&it}, &it
			},
			wantVal: item{ID: 5, Name: "a"},
		},
		{
			name: "maperr",
			q:    testQuery{columns: []string{"id"}, rows: [][]driver.Value{{1}}},
			dest: func() ([]interface{}, *item) {
				return []interface{}{&bad}, nil
			},
			wantErr: errors.New("Could not find column 'missing'.\n"),
		},
		{
			name: "closed",
			q:    testQuery{columns: []string{"id", "name"}, rows: [][]driver.Value{{5, "a"}}},
			dest: func() ([]interface{}, *item) {
				var it item
				return []interface{}{&it}, &it
			},
			setup: func(rows *sql.Rows) {
				_ = rows.Close()
			},
			skipNext: true,
			wantErr:  errors.New("sql: Rows are closed"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cfg := &testConfig{query: map[string]testQuery{query: tt.q}}
			db := newTestDB(t, cfg)
			rows, err := db.Query(query)
			if err != nil {
				t.Fatalf("query: %v", err)
			}
			if tt.setup != nil {
				tt.setup(rows)
			}
			defer rows.Close()
			if !tt.skipNext {
				if !rows.Next() {
					t.Fatalf("expected row")
				}
			}
			dest, it := tt.dest()
			err = (&Rows{Rows: rows}).Scan(dest...)
			if (err == nil) != (tt.wantErr == nil) {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr != nil && err != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && it != nil && *it != tt.wantVal {
				t.Fatalf("expected %+v, got %+v", tt.wantVal, *it)
			}
		})
	}
}

func TestFind(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		value  string
		want   int
	}{
		{name: "found", values: []string{"a", "b"}, value: "b", want: 1},
		{name: "missing", values: []string{"a"}, value: "z", want: -1},
		{name: "empty", values: nil, value: "a", want: -1},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := find(tt.values, tt.value); got != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestMapColumns(t *testing.T) {
	type inner struct {
		Name string `sql:"name"`
	}
	type outer struct {
		ID int `sql:"id"`
		In inner
	}
	type sliceItem struct {
		ID int `sql:"id"`
	}
	missing := struct {
		ID int `sql:"missing"`
	}{}

	tests := []struct {
		name    string
		columns []string
		input   interface{}
		wantErr error
		check   func(t *testing.T, dest []interface{})
	}{
		{
			name:    "struct",
			columns: []string{"id", "name"},
			input:   &outer{},
			check: func(t *testing.T, dest []interface{}) {
				o := dest[0].(*int)
				if o == nil {
					t.Fatalf("expected pointer")
				}
			},
		},
		{
			name:    "default",
			columns: []string{"id"},
			input:   new(int),
			check: func(t *testing.T, dest []interface{}) {
				if _, ok := dest[0].(*int); !ok {
					t.Fatalf("expected *int")
				}
			},
		},
		{
			name:    "slice",
			columns: []string{"id"},
			input:   []sliceItem{{}},
			check: func(t *testing.T, dest []interface{}) {
				if _, ok := dest[0].(*int); !ok {
					t.Fatalf("expected *int")
				}
			},
		},
		{
			name:    "missing",
			columns: []string{"id"},
			input:   &missing,
			wantErr: errors.New("Could not find column 'missing'.\n"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			dest := make([]interface{}, len(tt.columns))
			idx := 0
			err := mapColumns(dest, tt.input, tt.columns, "", &idx)
			if (err == nil) != (tt.wantErr == nil) {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr != nil && err != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && tt.check != nil {
				tt.check(t, dest)
			}
		})
	}
}
