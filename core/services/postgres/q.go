package postgres

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/smartcontractkit/sqlx"
)

var _ Queryer = Q{}

// Q wraps an underlying queryer (either a *sqlx.DB or a *sqlx.Tx)
//
// It is designed to make handling *sqlx.Tx or *sqlx.DB a little bit safer by
// preventing footguns such as having no deadline on contexts.
//
// It automatically adds the default context deadline to all non-context
// queries (if you _really_ want to issue a query without a context, use the
// underlying Queryer)
//
// This is not the prettiest construct but without macros its about the best we
// can do.
type Q struct {
	Queryer
	ParentCtx              context.Context
	DisableDefaultDeadline bool
}

func (q Q) Context() (context.Context, context.CancelFunc) {
	if q.DisableDefaultDeadline && q.ParentCtx == nil {
		return context.Background(), func() {}
	} else if !q.DisableDefaultDeadline && q.ParentCtx == nil {
		return DefaultQueryCtx()
	} else if q.DisableDefaultDeadline {
		return q.ParentCtx, func() {}
	} else {
		return DefaultQueryCtxWithParent(q.ParentCtx)
	}
}

func NewQ(queryer Queryer) Q {
	return Q{Queryer: queryer}
}

// NewQFromOpts is intended to be used in ORMs where the caller may wish to use
// either the default DB or pass an explicit Tx
func NewQFromOpts(qs []Q, queryer Queryer) Q {
	if len(qs) == 0 {
		return Q{Queryer: queryer}
	} else if len(qs) > 1 {
		panic("too many args")
	}
	return qs[0]
}

func NewQWithParentCtx(parentCtx context.Context, queryer Queryer) Q {
	return Q{ParentCtx: parentCtx, Queryer: queryer}
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

func (q Q) Transaction(fc func(tx *sqlx.Tx) error) error {
	ctx, cancel := q.Context()
	defer cancel()
	return SqlxTransaction(ctx, q.Queryer, fc)
}
func (q Q) Query(query string, args ...interface{}) (*sql.Rows, error) {
	ctx, cancel := q.Context()
	defer cancel()
	return q.Queryer.QueryContext(ctx, query, args...)
}
func (q Q) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	ctx, cancel := q.Context()
	defer cancel()
	return q.Queryer.QueryxContext(ctx, query, args...)
}
func (q Q) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	ctx, cancel := q.Context()
	defer cancel()
	return q.Queryer.QueryRowxContext(ctx, query, args...)
}
func (q Q) Exec(query string, args ...interface{}) (sql.Result, error) {
	ctx, cancel := q.Context()
	defer cancel()
	return q.Queryer.ExecContext(ctx, query, args...)
}
func (q Q) Select(dest interface{}, query string, args ...interface{}) error {
	ctx, cancel := q.Context()
	defer cancel()
	return q.Queryer.SelectContext(ctx, dest, query, args...)
}
