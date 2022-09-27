package noorm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrNoDatabaseInContext = errors.New("noorm: no database in context")
	ErrNoQuerierInContext  = errors.New("noorm: no querier in context")
)

type (
	ctxDatabaseKey    struct{}
	ctxTransactionKey struct{}
)

type database struct {
	*sql.DB
	dialect Dialect
}

// Querier is an abstraction over both *sql.DB and *sql.Tx.
type Querier interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
}

var (
	_ Querier = &sql.DB{}
	_ Querier = &sql.Tx{}
)

func WithDatabase(ctx context.Context, db *sql.DB) context.Context {
	return context.WithValue(ctx, ctxDatabaseKey{}, &database{
		DB:      db,
		dialect: guessDialect(db),
	})
}

func Begin(ctx context.Context, opts *sql.TxOptions) (context.Context, *sql.Tx, error) {
	db, ok := ctx.Value(ctxDatabaseKey{}).(*database)
	if !ok {
		return ctx, nil, fmt.Errorf("%w: cannot begin transaction", ErrNoDatabaseInContext)
	}

	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return ctx, nil, err
	}

	return context.WithValue(ctx, ctxTransactionKey{}, tx), tx, nil
}

func getQuerierAndDialect(ctx context.Context) (Querier, Dialect, error) {
	db, ok := ctx.Value(ctxDatabaseKey{}).(*database)
	if !ok {
		return nil, nil, ErrNoQuerierInContext
	}

	tx, ok := ctx.Value(ctxTransactionKey{}).(*sql.Tx)
	if ok {
		return tx, db.dialect, nil
	}

	return db, db.dialect, nil
}
