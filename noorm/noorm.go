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
// Exec expects a Querier to be present in the context (see WithDatabase).
func Exec(ctx context.Context, query string, args ArgumentSource) (sql.Result, error) {
	if err := checkValidArgs(args); err != nil {
		return nil, err
	}

	querier, dialect, err := QuerierFrom(ctx)
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
// Iterate expects a Querier to be present in the context (see WithDatabase).
func Iterate[T Struct](ctx context.Context, query string, args ArgumentSource) (Iterator[T], error) {
	if err := checkValidArgs(args); err != nil {
		return nil, err
	}

	querier, dialect, err := QuerierFrom(ctx)
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
// Query expects a Querier to be present in the context (see WithDatabase).
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

// QueryFirst executed a query and returns the first result.
// If the query yields no rows, sql.ErrNoRows is returned.
// If the query yields more than one row, the remaining rows are discarded.
// QueryFirst expects a Querier to be present in the context (see WithDatabase).
func QueryFirst[T Struct](ctx context.Context, query string, args ArgumentSource) (*T, error) {
	iter, err := Iterate[T](ctx, query, args)
	if err != nil {
		return nil, err
	}

	defer iter.Close()

	if !iter.Next() {
		if err := iter.Err(); err != nil {
			return nil, err
		}

		return nil, sql.ErrNoRows
	}

	value, err := iter.Value()
	if err != nil {
		return nil, err
	}

	return &value, nil
}
