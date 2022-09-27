package noorm

import (
	"context"
	"database/sql"
)

// Struct must be a struct.
type Struct any

// Iterator is a typed wrapper for *sql.Rows, which scans rows into T.
type Iterator[T Struct] interface {
	// Next proceeds with the next row.
	// Next must be called before the first row can be scanned.
	Next() bool
	// Err returns the latest iteration error andshould be checked whenever Next returns false.
	Err() error
	// Value scans the current row into a new T.
	Value() (T, error)
	// Close closes the underlying *sql.Rows.
	Close() error
}

// ArgumentSource captures the provided named or positional arguments as a single type.
// See `Named`, `Positional` and `None` for implementations.
type ArgumentSource interface {
	arg(name string) (any, error)
}

// Exec executed a query without returning rows.
// Exec expects a Querier to be present in the context (see WithQuerier).
func Exec(ctx context.Context, query string, args ArgumentSource) (sql.Result, error) {
	if err := checkValidArgs(args); err != nil {
		return nil, err
	}

	querier, dialect, err := getQuerierAndDialect(ctx)
	if err != nil {
		return nil, err
	}

	rebound, params, err := rebindQuery(query, dialect, args)
	if err != nil {
		return nil, err
	}

	return querier.ExecContext(ctx, rebound, params...)
}

// Iterate executed a query and returns an iterator of the rows.
// Iterate expects a Querier to be present in the context (see WithQuerier).
func Iterate[T Struct](ctx context.Context, query string, args ArgumentSource) (Iterator[T], error) {
	if err := checkValidArgs(args); err != nil {
		return nil, err
	}

	querier, dialect, err := getQuerierAndDialect(ctx)
	if err != nil {
		return nil, err
	}

	rebound, params, err := rebindQuery(query, dialect, args)
	if err != nil {
		return nil, err
	}

	rows, err := querier.QueryContext(ctx, rebound, params...)
	if err != nil {
		return nil, err
	}

	iter, err := newIterator[T](rows)
	if err != nil {
		rows.Close() // close rows early, because we do not return a reference to it
		return nil, err
	}

	return iter, nil
}

// Query executed a query and returns a slice of T.
// Query expects a Querier to be present in the context (see WithQuerier).
func Query[T Struct](ctx context.Context, query string, args ArgumentSource) ([]T, error) {
	iter, err := Iterate[T](ctx, query, args)
	if err != nil {
		return nil, err
	}

	defer iter.Close()

	var valueSlice []T

	for iter.Next() {
		value, err := iter.Value()
		if err != nil {
			return nil, err
		}

		valueSlice = append(valueSlice, value)
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return valueSlice, nil
}
