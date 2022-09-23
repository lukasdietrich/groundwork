package noorm

type Dialect interface {
	Placeholder() string
}

type defaultDialect struct{}

func (defaultDialect) Placeholder() string {
	return "?"
}

type querierWithDialect struct {
	Querier
	dialect Dialect
}

func (q querierWithDialect) Dialect() Dialect {
	return q.dialect
}

func WithDialect(querier Querier, dialect Dialect) Querier {
	return querierWithDialect{
		Querier: querier,
		dialect: dialect,
	}
}

func dialect(querier Querier) Dialect {
	if qd, ok := querier.(interface{ Dialect() Dialect }); ok {
		return qd.Dialect()
	}

	return defaultDialect{}
}
