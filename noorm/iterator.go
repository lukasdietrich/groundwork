package noorm

type iterator[T any] struct {
	s *scanner
}

func newIterator[T any](s *scanner) Iterator[T] {
	return iterator[T]{s: s}
}

func (i iterator[T]) Close() error {
	return i.s.close()
}

func (i iterator[T]) Next() bool {
	return i.s.next()
}

func (i iterator[T]) Err() error {
	return i.s.err()
}

func (i iterator[T]) Value() (T, error) {
	var value T
	err := i.s.scan(&value)
	return value, err
}
