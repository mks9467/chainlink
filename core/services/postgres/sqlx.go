package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx/reflectx"
	mapper "github.com/scylladb/go-reflectx"
	"github.com/smartcontractkit/sqlx"
	"gorm.io/gorm"
)

type Queryer interface {
	sqlx.Ext
	sqlx.ExtContext
	sqlx.Preparer
	sqlx.PreparerContext
	sqlx.Queryer
	Select(dest interface{}, query string, args ...interface{}) error
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

func WrapDbWithSqlx(rdb *sql.DB) *sqlx.DB {
	db := sqlx.NewDb(rdb, "postgres")
	db.MapperFunc(mapper.CamelToSnakeASCII)
	return db
}

func UnwrapGormDB(db *gorm.DB) *sqlx.DB {
	d, err := db.DB()
	if err != nil {
		panic(err)
	}
	return WrapDbWithSqlx(d)
}

func TryUnwrapGormDB(db *gorm.DB) *sqlx.DB {
	if db == nil {
		return nil
	}
	return UnwrapGormDB(db)
}

func UnwrapGorm(db *gorm.DB) Queryer {
	if tx, ok := db.Statement.ConnPool.(*sql.Tx); ok {
		// if a transaction is currently present use that instead
		mapper := reflectx.NewMapperFunc("db", mapper.CamelToSnakeASCII)
		txx := sqlx.NewTx(tx, db.Dialector.Name())
		txx.Mapper = mapper
		return txx
	}
	return UnwrapGormDB(db)
}

func SqlxTransactionWithDefaultCtx(q Queryer, fc func(tx *sqlx.Tx) error, txOpts ...sql.TxOptions) (err error) {
	ctx, cancel := DefaultQueryCtx()
	defer cancel()
	return SqlxTransaction(ctx, q, fc, txOpts...)
}

func SqlxTransaction(ctx context.Context, q Queryer, fc func(tx *sqlx.Tx) error, txOpts ...sql.TxOptions) (err error) {
	switch db := q.(type) {
	case *sqlx.Tx:
		// nested transaction: just use the outer transaction
		err = fc(db)

	case *sqlx.DB:
		opts := &DefaultSqlTxOptions
		if len(txOpts) > 0 {
			opts = &txOpts[0]
		}

		var tx *sqlx.Tx
		tx, err = db.BeginTxx(ctx, opts)
		panicked := false

		defer func() {
			// Make sure to rollback when panic, Block error or Commit error
			if panicked || err != nil {
				if perr := tx.Rollback(); perr != nil {
					panic(perr)
				}
			}
		}()

		_, err = tx.Exec(fmt.Sprintf(`SET LOCAL lock_timeout = %v; SET LOCAL idle_in_transaction_session_timeout = %v;`, LockTimeout.Milliseconds(), IdleInTxSessionTimeout.Milliseconds()))
		if err != nil {
			return errors.Wrap(err, "error setting transaction timeouts")
		}

		panicked = true
		err = fc(tx)
		panicked = false

		if err == nil {
			err = errors.WithStack(tx.Commit())
		}
	default:
		err = errors.Errorf("invalid db type")
	}

	return
}

func PrepareGet(q Queryer, sql string, dest interface{}, arg interface{}) error {
	stmt, err := q.PrepareNamed(sql)
	if err != nil {
		return errors.Wrap(err, "error preparing named statement")
	}
	return errors.Wrap(stmt.Get(dest, arg), "error in get query")
}

func PrepareQueryRowx(q Queryer, sql string, dest interface{}, arg interface{}) error {
	stmt, err := q.PrepareNamed(sql)
	if err != nil {
		return errors.Wrap(err, "error preparing named statement")
	}
	return errors.Wrap(stmt.QueryRowx(arg).Scan(dest), "error querying row")
}

type Q struct {
	Queryer
	Ctx                    context.Context
	DisableDefaultDeadline bool
}

func (q *Q) Context() (context.Context, context.CancelFunc) {
	if q.DisableDefaultDeadline && q.Ctx == nil {
		return context.Background(), func() {}
	} else if !q.DisableDefaultDeadline && q.Ctx == nil {
		return DefaultQueryCtx()
	} else if q.DisableDefaultDeadline {
		return q.Ctx, func() {}
	} else {
		return DefaultQueryCtxWithParent(q.Ctx)
	}

}

func GetQ(qs []Q, queryer Queryer) Q {
	if len(qs) == 0 {
		return Q{Queryer: queryer}
	}
	return qs[0]
}
