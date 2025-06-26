package database

import (
	"context"
	"database/sql"
)

// NewDB returns a DB wrapping the sql.DB.
//
// This is to make it possible to add tracing and logging
// to queries and such.
//
// The returned DB can be used just like the sql.DB
// and does not care about the underlying database tech or
// SQL dialect.
func NewDB(db *sql.DB) *DB {
	return &DB{DB: db}
}

// DB wraps sql.DB to provide tracing and logging.
type DB struct {
	*sql.DB
}

func (db *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return db.DB.Query(query, args...)
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.DB.QueryContext(ctx, query, args...)
}

func (db *DB) QueryRow(query string, args ...any) *sql.Row {
	return db.DB.QueryRow(query, args...)
}

func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return db.DB.QueryRowContext(ctx, query, args...)
}

func (db *DB) Exec(query string, args ...any) (sql.Result, error) {
	return db.DB.Exec(query, args...)
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.DB.ExecContext(ctx, query, args...)
}

func (db *DB) Prepare(query string) (*sql.Stmt, error) {
	return db.DB.Prepare(query)
}

func (db *DB) Begin() (*Tx, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}

	return &Tx{Tx: tx}, nil
}

func (db *DB) Close() error {
	if db.DB != nil {
		return db.DB.Close()
	}

	return nil
}

type Tx struct {
	*sql.Tx
}

func (t *Tx) Query(query string, args ...any) (*sql.Rows, error) {
	return t.Tx.Query(query, args...)
}

func (t *Tx) QueryRow(query string, args ...any) *sql.Row {
	return t.Tx.QueryRow(query, args...)
}

func (t *Tx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return t.Tx.QueryRowContext(ctx, query, args...)
}

func (t *Tx) Exec(query string, args ...any) (sql.Result, error) {
	return t.Tx.Exec(query, args...)
}

func (t *Tx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.Tx.ExecContext(ctx, query, args...)
}

func (t *Tx) Prepare(query string) (*sql.Stmt, error) {
	return t.Tx.Prepare(query)
}

func (t *Tx) Commit() error {
	return t.Tx.Commit()
}

func (t *Tx) Rollback() error {
	return t.Tx.Rollback()
}
