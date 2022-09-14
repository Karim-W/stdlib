package stdlib

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	tracer "github.com/BetaLixT/appInsightsTrace"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type NativeDatabase interface {
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
	Conn(ctx context.Context) (*sql.Conn, error)
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Ping() error
	PingContext(ctx context.Context) error
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	SetConnMaxIdleTime(d time.Duration)
	SetConnMaxLifetime(d time.Duration)
	Stats() sql.DBStats
}

type dbImpl struct {
	logger   *zap.Logger
	db       *sql.DB
	pingLock sync.Mutex
	t        *tracer.AppInsightsCore
	driver   string
	name     string
}

func NativeDatabaseProvider(Driver string, DSN string) NativeDatabase {
	l := getLoggerInstance()
	switch Driver {
	case "postgres":
		if db, err := sql.Open("postgres", DSN); err != nil {
			panic(err)
		} else {
			l.Info("[DATABASE]\tSucessfuly Connected to postgres database")
			ndb := &dbImpl{
				logger:   getLoggerInstance().logger.Desugar(),
				db:       db,
				pingLock: sync.Mutex{},
				driver:   Driver,
			}
			ndb.Ping()
			return ndb
		}
	default:
		panic("Unsupported driver")
	}
}

func TracedNativeDBWrapper(
	Driver string,
	DSN string,
	t *tracer.AppInsightsCore,
	name string,
) NativeDatabase {
	l := getLoggerInstance()
	switch Driver {
	case "postgres":
		if db, err := sql.Open("postgres", DSN); err != nil {
			panic(err)
		} else {
			l.Info("[DATABASE]\tSucessfuly Connected to postgres database")
			ndb := &dbImpl{
				logger:   getLoggerInstance().logger.Desugar(),
				db:       db,
				pingLock: sync.Mutex{},
				t:        t,
				driver:   Driver,
				name:     name,
			}
			ndb.Ping()
			return ndb
		}
	default:
		panic("Unsupported driver")
	}
}

func (d *dbImpl) WithLogger(l *zap.Logger) *dbImpl {
	d.logger = l
	return d
}

func (d *dbImpl) Begin() (*sql.Tx, error) {
	return d.db.Begin()
}

func (d *dbImpl) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return d.db.BeginTx(ctx, opts)
}

func (d *dbImpl) Close() error {
	return d.db.Close()
}

func (d *dbImpl) Conn(ctx context.Context) (*sql.Conn, error) {
	return d.db.Conn(ctx)
}

func (d *dbImpl) Exec(query string, args ...any) (sql.Result, error) {
	now := time.Now()
	res, err := d.db.Exec(query, args...)
	elapsed := float64(time.Since(now).Microseconds()) / 1000.0
	if err != nil {
		d.logger.Error("[DATABASE]  Error executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Error(err))
	} else {
		d.logger.Info("[DATABASE]  Executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Float64("elapsed(ms)", elapsed))
	}
	return res, err
}

func (d *dbImpl) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	now := time.Now()
	res, err := d.db.ExecContext(ctx, query, args...)
	after := time.Now()
	elapsed := float64(after.Sub(now).Microseconds()) / 1000.0
	if err != nil {
		d.logger.Error("[DATABASE]  Error executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Error(err))
		d.t.TraceException(ctx, err, 0, map[string]string{
			"query": query,
			"args":  fmt.Sprintf("%v", args),
			"error": err.Error(),
		})
		d.t.TraceDependency(ctx, "", d.driver, d.name, "EXEC"+query, false, now, after, map[string]string{
			"query": query,
			"args":  fmt.Sprintf("%v", args),
			"error": err.Error(),
		})
	} else {
		d.logger.Info("[DATABASE]  Executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Float64("elapsed(ms)", elapsed))
		d.t.TraceDependency(ctx, "", d.driver, d.name, "EXEC"+query, true, now, after, map[string]string{
			"query": query,
			"args":  fmt.Sprintf("%v", args),
		})
	}
	return res, err
}

func (d *dbImpl) Ping() error {
	d.pingLock.Lock()
	go func() {
		for {
			time.Sleep(time.Second * 5)
			d.db.Ping()
		}
	}()
	return d.db.Ping()
}

func (d *dbImpl) PingContext(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *dbImpl) Prepare(query string) (*sql.Stmt, error) {
	return d.db.Prepare(query)
}

func (d *dbImpl) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return d.db.PrepareContext(ctx, query)
}

func (d *dbImpl) Query(query string, args ...any) (*sql.Rows, error) {
	now := time.Now()
	res, err := d.db.Query(query, args...)
	elapsed := float64(time.Since(now).Microseconds()) / 1000.0
	if err != nil {
		d.logger.Error("[DATABASE]  Error executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Error(err))
	} else {
		d.logger.Info("[DATABASE]  Executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Float64("elapsed(ms)", elapsed))
	}
	return res, err
}

func (d *dbImpl) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	now := time.Now()
	res, err := d.db.QueryContext(ctx, query, args...)
	after := time.Now()
	elapsed := float64(after.Sub(now).Microseconds()) / 1000.0
	if err != nil {
		d.logger.Error("[DATABASE]  Error executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Error(err))
		d.t.TraceException(ctx, err, 0, map[string]string{
			"query": query,
			"args":  fmt.Sprintf("%v", args),
			"error": err.Error(),
		})
		d.t.TraceDependency(ctx, "", d.driver, d.name, "EXEC"+query, false, now, after, map[string]string{
			"query": query,
			"args":  fmt.Sprintf("%v", args),
			"error": err.Error(),
		})
	} else {
		d.logger.Info("[DATABASE]  Executing query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Float64("elapsed(ms)", elapsed))
		d.t.TraceDependency(ctx, "", d.driver, d.name, "EXEC"+query, true, now, after, map[string]string{
			"query": query,
			"args":  fmt.Sprintf("%v", args),
		})
	}
	return res, err
}

func (d *dbImpl) QueryRow(query string, args ...any) *sql.Row {
	now := time.Now()
	r := d.db.QueryRow(query, args...)
	elapsed := float64(time.Since(now).Microseconds()) / 1000.0
	d.logger.Info("[DATABASE]  Executing query",
		zap.String("query", query),
		zap.Any("args", args),
		zap.Float64("elapsed(ms)", elapsed))
	return r
}

func (d *dbImpl) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	now := time.Now()
	r := d.db.QueryRowContext(ctx, query, args...)
	after := time.Now()
	elapsed := float64(after.Sub(now).Microseconds()) / 1000.0
	d.logger.Info("[DATABASE]  Executing query",
		zap.String("query", query),
		zap.Any("args", args),
		zap.Float64("elapsed(ms)", elapsed))
	d.t.TraceDependency(ctx, "", d.driver, d.name, "EXEC"+query, true, now, after, map[string]string{
		"query": query,
		"args":  fmt.Sprintf("%v", args),
	})
	return r
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
