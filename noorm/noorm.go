package noorm

import (
	"context"
	"database/sql"
)

// Struct must be a struct.
type Struct any

type Iterator[T Struct] interface {
	// Next proceeds with the next row.
	// Next must be called before the first row can be scanned.
	Next() bool
	Err() error
	Value() (T, error)
	Close() error
}

// ArgumentSource captures the provided named or positional arguments as a single type.
// See `Named` and `Positional` for implementations.
type ArgumentSource interface {
	arg(name string) (any, error)
}

// Exec executed a query without returning rows.
// Exec expects a Querier to be present in the context (see WithQuerier).
func Exec(ctx context.Context, query string, args ArgumentSource) (sql.Result, error) {
	querier, dialect, err := getQuerier(ctx)
	if err != nil {
		return nil, err
	}

	rebound, params, err := rebindQuery(query, dialect.Placeholder(), args)
	if err != nil {
		return nil, err
	}

	return querier.ExecContext(ctx, rebound, params...)
}

// Iterate executed a query and returns an iterator of the rows.
// Iterate expects a Querier to be present in the context (see WithQuerier).
func Iterate[T Struct](ctx context.Context, query string, args ArgumentSource) (Iterator[T], error) {
	scanner, err := queryScanner[T](ctx, query, args)
	if err != nil {
		return nil, err
	}

	iterator := newIterator[T](scanner)
	return iterator, nil
}

// Query executed a query and returns a slice of T.
// Query expects a Querier to be present in the context (see WithQuerier).
func Query[T Struct](ctx context.Context, query string, args ArgumentSource) ([]T, error) {
	scanner, err := queryScanner[T](ctx, query, args)
	if err != nil {
		return nil, err
	}

	defer scanner.close()

	var valueSlice []T

	for scanner.next() {
		var value T
		if err := scanner.scan(&value); err != nil {
			return nil, err
		}

		valueSlice = append(valueSlice, value)
	}

	if err := scanner.err(); err != nil {
		return nil, err
	}

	return valueSlice, nil
}

func queryScanner[T Struct](ctx context.Context, query string, args ArgumentSource) (*scanner, error) {
	querier, dialect, err := getQuerier(ctx)
	if err != nil {
		return nil, err
	}

	rebound, params, err := rebindQuery(query, dialect.Placeholder(), args)
	if err != nil {
		return nil, err
	}

	rows, err := querier.QueryContext(ctx, rebound, params...)
	if err != nil {
		return nil, err
	}

	scanner, err := newScanner[T](rows)
	if err != nil {
		rows.Close()
		return nil, err
	}

	return scanner, nil
}
