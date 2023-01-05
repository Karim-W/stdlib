package sqldb

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSuccessfulBegin(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectBegin()
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	tx, err := db.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, tx)
}

func TestFailedBegin(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	tx, err := db.Begin()
	assert.NotNil(t, err)
	assert.Nil(t, tx)
}

func TestFailedBeginReturnError(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	err = fmt.Errorf("some error")
	mock.ExpectBegin().WillReturnError(err)
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	tx, err := db.Begin()
	assert.NotNil(t, err)
	assert.Nil(t, tx)
}

func TestClose(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectClose()
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	err = db.Close()
	assert.Nil(t, err)
}

func TestFailedClose(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	err = fmt.Errorf("some error")
	mock.ExpectClose().WillReturnError(err)
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	err = db.Close()
	assert.NotNil(t, err)
}

func TestPing(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectPing()
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	err = db.Ping()
	assert.Nil(t, err)
}

func TestPrepare(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectPrepare("INSERT INTO foo VALUES (?)")
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	stmt, err := db.Prepare("INSERT INTO foo VALUES (?)")
	assert.Nil(t, err)
	assert.NotNil(t, stmt)
}

func TestFailedPrepare(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	err = fmt.Errorf("some error")
	mock.ExpectPrepare("INSERT INTO foo VALUES (?)").WillReturnError(err)
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	stmt, err := db.Prepare("INSERT INTO foo VALUES (?)")
	assert.NotNil(t, err)
	assert.Nil(t, stmt)
}

func TestPrepareContext(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectPrepare("INSERT INTO foo VALUES (?)")
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	stmt, err := db.PrepareContext(context.TODO(), "INSERT INTO foo VALUES (?)")
	assert.Nil(t, err)
	assert.NotNil(t, stmt)
}

func TestFailedPrepareContext(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	err = fmt.Errorf("some error")
	mock.ExpectPrepare("INSERT INTO foo VALUES (?)").WillReturnError(err)
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	stmt, err := db.PrepareContext(context.TODO(), "INSERT INTO foo VALUES (?)")
	assert.NotNil(t, err)
	assert.Nil(t, stmt)
}

func TestQuery(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectQuery("SELECT bar FROM foo").WithArgs().WillReturnRows(sqlmock.NewRows([]string{"bar"}).AddRow("baz"))
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	rows, err := db.Query("SELECT bar FROM foo")
	assert.Nil(t, err)
	assert.NotNil(t, rows)
}

func TestFailedQuery(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	err = fmt.Errorf("some error")
	mock.ExpectQuery("SELECT bar FROM foo").WillReturnError(err)
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	rows, err := db.Query("SELECT bar FROM foo")
	assert.NotNil(t, err)
	assert.Nil(t, rows)
}

func TestFailedQueryMistmatch(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	err = fmt.Errorf("some error")
	mock.ExpectQuery("SELECT * FROM foo WHERE id = $1").WithArgs(1)
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	rows, err := db.Query("SELECT * FROM foo WHERE id = $1")
	assert.NotNil(t, err)
	assert.Nil(t, rows)
}

func TestQueryContext(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectQuery("SELECT bar FROM foo").WithArgs().WillReturnRows(sqlmock.NewRows([]string{"bar"}).AddRow("baz"))
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	rows, err := db.QueryContext(context.TODO(), "SELECT bar FROM foo")
	assert.Nil(t, err)
	assert.NotNil(t, rows)
}

func TestFailedQueryContext(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	err = fmt.Errorf("some error")
	mock.ExpectQuery("SELECT bar FROM foo").WillReturnError(err)
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	rows, err := db.QueryContext(context.TODO(), "SELECT bar FROM foo")
	assert.NotNil(t, err)
	assert.Nil(t, rows)
}

func TestFailedQueryContextMistmatch(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	err = fmt.Errorf("some error")
	mock.ExpectQuery("SELECT * FROM foo WHERE id = $1").WithArgs(1)
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	rows, err := db.QueryContext(context.TODO(), "SELECT * FROM foo WHERE id = $1")
	assert.NotNil(t, err)
	assert.Nil(t, rows)
}

func TestQueryRow(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectQuery("SELECT bar FROM foo").WithArgs().WillReturnRows(sqlmock.NewRows([]string{"bar"}).AddRow("baz"))
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	row := db.QueryRow("SELECT bar FROM foo")
	assert.NotNil(t, row)
}

func TestFailedQueryRow(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	err = fmt.Errorf("some error")
	mock.ExpectQuery("SELECT bar FROM foo").WillReturnError(err)
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	row := db.QueryRow("SELECT bar FROM foo")
	assert.NotNil(t, row)
}

func TestFailedQueryRowMistmatch(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	err = fmt.Errorf("some error")
	mock.ExpectQuery("SELECT * FROM foo WHERE id = $1").WithArgs(1)
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	row := db.QueryRow("SELECT * FROM foo WHERE id = $1")
	assert.NotNil(t, row)
}

func TestQueryRowContext(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectQuery("SELECT bar FROM foo").WithArgs().WillReturnRows(sqlmock.NewRows([]string{"bar"}).AddRow("baz"))
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	row := db.QueryRowContext(context.TODO(), "SELECT bar FROM foo")
	assert.NotNil(t, row)
}

func TestFailedQueryRowContext(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	err = fmt.Errorf("some error")
	mock.ExpectQuery("SELECT bar FROM foo").WillReturnError(err)
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	row := db.QueryRowContext(context.TODO(), "SELECT bar FROM foo")
	assert.NotNil(t, row)
}

func TestFailedQueryRowContextMistmatch(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	err = fmt.Errorf("some error")
	mock.ExpectQuery("SELECT * FROM foo WHERE id = $1").WithArgs(1)
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	row := db.QueryRowContext(context.TODO(), "SELECT * FROM foo WHERE id = $1")
	assert.NotNil(t, row)
}

func TestExec(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectExec("INSERT INTO foo").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	result, err := db.Exec("INSERT INTO foo")
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestFailedExec(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	err = fmt.Errorf("some error")
	mock.ExpectExec("INSERT INTO foo").WillReturnError(err)
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	result, err := db.Exec("INSERT INTO foo")
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestFailedExecMistmatch(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	err = fmt.Errorf("some error")
	mock.ExpectExec("INSERT INTO foo").WithArgs(1)
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	result, err := db.Exec("INSERT INTO foo")
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestExecContext(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectExec("INSERT INTO foo").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	result, err := db.ExecContext(context.TODO(), "INSERT INTO foo")
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestFailedExecContext(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	err = fmt.Errorf("some error")
	mock.ExpectExec("INSERT INTO foo").WillReturnError(err)
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	result, err := db.ExecContext(context.TODO(), "INSERT INTO foo")
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestFailedExecContextMistmatch(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	err = fmt.Errorf("some error")
	mock.ExpectExec("INSERT INTO foo").WithArgs(1)
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	result, err := db.ExecContext(context.TODO(), "INSERT INTO foo")
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestSuccessfulTransaction(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectBegin()
	mock.ExpectCommit()
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	tx, err := db.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	err = tx.Commit()
	assert.Nil(t, err)
}

func TestFailedTransaction(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectBegin()
	mock.ExpectRollback()
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	tx, err := db.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	err = tx.Rollback()
	assert.Nil(t, err)
}

func TestFailedTransactionCommit(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(fmt.Errorf("some error"))
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	tx, err := db.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	err = tx.Commit()
	assert.NotNil(t, err)
}

func TestSuccessTransactionExec(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO foo").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	tx, err := db.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	result, err := tx.Exec("INSERT INTO foo")
	assert.Nil(t, err)
	assert.NotNil(t, result)
	err = tx.Commit()
	assert.Nil(t, err)
}

func TestFailedTransactionExec(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO foo").WithArgs().WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	tx, err := db.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	result, err := tx.Exec("INSERT INTO foo")
	assert.NotNil(t, err)
	assert.Nil(t, result)
	err = tx.Rollback()
	assert.Nil(t, err)
}

func TestSuccessTranasactionExecContext(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO foo").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	tx, err := db.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	result, err := tx.ExecContext(context.TODO(), "INSERT INTO foo")
	assert.Nil(t, err)
	assert.NotNil(t, result)
	err = tx.Commit()
	assert.Nil(t, err)
}

func TestFailedTransactionExecContext(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO foo").WithArgs().WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	tx, err := db.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	result, err := tx.ExecContext(context.TODO(), "INSERT INTO foo")
	assert.NotNil(t, err)
	assert.Nil(t, result)
	err = tx.Rollback()
	assert.Nil(t, err)
}

func TestSuccessTransactionQuery(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT bar FROM foo").WithArgs().WillReturnRows(sqlmock.NewRows([]string{"bar"}).AddRow("baz"))
	mock.ExpectCommit()
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	tx, err := db.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	rows, err := tx.Query("SELECT bar FROM foo")
	assert.Nil(t, err)
	assert.NotNil(t, rows)
	err = tx.Commit()
	assert.Nil(t, err)
}

func TestFailedTransactionQuery(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT bar FROM foo").WithArgs().WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	tx, err := db.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	rows, err := tx.Query("SELECT bar FROM foo")
	assert.NotNil(t, err)
	assert.Nil(t, rows)
	err = tx.Rollback()
	assert.Nil(t, err)
}

func TestSuccessTransactionQueryContext(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT bar FROM foo").WithArgs().WillReturnRows(sqlmock.NewRows([]string{"bar"}).AddRow("baz"))
	mock.ExpectCommit()
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	tx, err := db.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	rows, err := tx.QueryContext(context.TODO(), "SELECT bar FROM foo")
	assert.Nil(t, err)
	assert.NotNil(t, rows)
	err = tx.Commit()
	assert.Nil(t, err)
}

func TestFailedTransactionQueryContext(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT bar FROM foo").WithArgs().WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	tx, err := db.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	rows, err := tx.QueryContext(context.TODO(), "SELECT bar FROM foo")
	assert.NotNil(t, err)
	assert.Nil(t, rows)
	err = tx.Rollback()
	assert.Nil(t, err)
}

func TestSuccessTransactionQueryRow(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT bar FROM foo").WithArgs().WillReturnRows(sqlmock.NewRows([]string{"bar"}).AddRow("baz"))
	mock.ExpectCommit()
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	tx, err := db.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	row := tx.QueryRow("SELECT bar FROM foo")
	assert.NotNil(t, row)
	err = tx.Commit()
	assert.Nil(t, err)
}

func TestFailedTransactionQueryRow(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT bar FROM foo").WithArgs().WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	tx, err := db.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	row := tx.QueryRow("SELECT bar FROM foo")
	assert.NotNil(t, row)
	err = tx.Rollback()
	assert.Nil(t, err)
}

func TestSuccessTransactionQueryRowContext(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT bar FROM foo").WithArgs().WillReturnRows(sqlmock.NewRows([]string{"bar"}).AddRow("baz"))
	mock.ExpectCommit()
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	tx, err := db.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	row := tx.QueryRowContext(context.TODO(), "SELECT bar FROM foo")
	assert.NotNil(t, row)
	err = tx.Commit()
	assert.Nil(t, err)
}

func TestFailedTransactionQueryRowContext(t *testing.T) {
	d, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer d.Close()
	assert.NotNil(t, mock)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT bar FROM foo").WithArgs().WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()
	logger := zap.NewExample()
	db := &dbImpl{
		db:     d,
		logger: logger,
	}
	tx, err := db.Begin()
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	row := tx.QueryRowContext(context.TODO(), "SELECT bar FROM foo")
	assert.NotNil(t, row)
	err = tx.Rollback()
	assert.Nil(t, err)
}
