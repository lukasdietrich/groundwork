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

type Database struct {
	*sql.DB
	dialect Dialect
}

type transaction struct {
	*sql.Tx

	// save pointer to db to avoid traversing the context twice
	db *Database
}

func New(db *sql.DB, dialect Dialect) *Database {
	return &Database{
		DB:      db,
		dialect: dialect,
	}
}

func Open(driverName, dataSourceName string) (*Database, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	return New(db, guessDialect(driverName)), nil
}

func WithDatabase(ctx context.Context, db *Database) context.Context {
	return context.WithValue(ctx, ctxDatabaseKey{}, db)
}

func Begin(ctx context.Context, opts *sql.TxOptions) (context.Context, *sql.Tx, error) {
	db, ok := ctx.Value(ctxDatabaseKey{}).(*Database)
	if !ok {
		return ctx, nil, fmt.Errorf("%w: cannot begin transaction", ErrNoDatabaseInContext)
	}

	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return ctx, nil, err
	}

	ctx = context.WithValue(ctx, ctxTransactionKey{}, &transaction{
		Tx: tx,
		db: db,
	})

	return ctx, tx, nil
}

func QuerierFrom(ctx context.Context) (Querier, Dialect, error) {
	tx, ok := ctx.Value(ctxTransactionKey{}).(*transaction)
	if ok {
		return tx, tx.db.dialect, nil
	}

	db, ok := ctx.Value(ctxDatabaseKey{}).(*Database)
	if ok {
		return db, db.dialect, nil
	}

	return nil, nil, ErrNoQuerierInContext
}
