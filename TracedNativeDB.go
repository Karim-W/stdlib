package stdlib

import (
	"database/sql"
	"time"
)

type NativeDatabase interface {
	Begin() (*sql.Tx, error)
	BeginTx(ctx Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
	Conn(ctx Context) (*sql.Conn, error)
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx Context, query string, args ...any) (sql.Result, error)
	Ping() error
	PingContext(ctx Context) error
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx Context, query string) (*sql.Stmt, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx Context, query string, args ...any) *sql.Row
	SetConnMaxIdleTime(d time.Duration)
	SetConnMaxLifetime(d time.Duration)
	Stats() sql.DBStats
}

type dbImpl struct {
	db *sql.DB
}

func NativeDatabaseProvider(db *sql.DB) NativeDatabase {
	return &dbImpl{
		db: db,
	}
}

func (d *dbImpl) Begin() (*sql.Tx, error) {
	return d.db.Begin()
}

func (d *dbImpl) BeginTx(ctx Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return d.db.BeginTx(ctx.Context, opts)
}

func (d *dbImpl) Close() error {
	return d.db.Close()
}

func (d *dbImpl) Conn(ctx Context) (*sql.Conn, error) {
	return d.db.Conn(ctx.Context)
}

func (d *dbImpl) Exec(query string, args ...any) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

func (d *dbImpl) ExecContext(ctx Context, query string, args ...any) (sql.Result, error) {
	return d.db.ExecContext(ctx.Context, query, args...)
}

func (d *dbImpl) Ping() error {
	return d.db.Ping()
}

func (d *dbImpl) PingContext(ctx Context) error {
	return d.db.PingContext(ctx.Context)
}

func (d *dbImpl) Prepare(query string) (*sql.Stmt, error) {
	return d.db.Prepare(query)
}

func (d *dbImpl) PrepareContext(ctx Context, query string) (*sql.Stmt, error) {
	return d.db.PrepareContext(ctx.Context, query)
}

func (d *dbImpl) Query(query string, args ...any) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

func (d *dbImpl) QueryContext(ctx Context, query string, args ...any) (*sql.Rows, error) {
	return d.db.QueryContext(ctx.Context, query, args...)
}

func (d *dbImpl) QueryRow(query string, args ...any) *sql.Row {
	return d.db.QueryRow(query, args...)
}

func (d *dbImpl) QueryRowContext(ctx Context, query string, args ...any) *sql.Row {
	return d.db.QueryRowContext(ctx.Context, query, args...)
}

func (i *dbImpl) SetConnMaxIdleTime(d time.Duration) {
	i.db.SetConnMaxIdleTime(d)
}

func (i *dbImpl) SetConnMaxLifetime(d time.Duration) {
	i.db.SetConnMaxLifetime(d)
}

func (i *dbImpl) Stats() sql.DBStats {
	return i.db.Stats()
}
