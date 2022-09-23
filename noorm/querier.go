package noorm

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrNoQuerierInContext = errors.New("noorm: no querier in context")
)

type ctxQuerierKey struct{}

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

// WithQuerier is used to store a given querier in the current context.
// Calls to Query, and Exec expect a querier to be present in the context and will return an error
// otherwise.
//
// It is a good idea to start a transaction in a http middleware and store it in the request context
// to mimic the @Transactional annotation of Java.
func WithQuerier(ctx context.Context, querier Querier) context.Context {
	return context.WithValue(ctx, ctxQuerierKey{}, querier)
}

func getQuerier(ctx context.Context) (Querier, Dialect, error) {
	querier, ok := ctx.Value(ctxQuerierKey{}).(Querier)
	if !ok {
		return nil, nil, ErrNoQuerierInContext
	}

	return querier, dialect(querier), nil
}
