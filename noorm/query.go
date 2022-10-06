package noorm

import (
	"bytes"
	"sort"
)

type requireExplicitFields struct{}

type SQL struct {
	requireExplicitFields

	Query string
	Args  ArgumentSource
}

func (s SQL) rebind(dialect Dialect) (string, []any, error) {
	if s.Args == nil {
		s.Args = None()
	}

	if err := checkValidArgs(s.Args); err != nil {
		return "", nil, err
	}

	return rebindQuery(dialect, s.Query, s.Args)
}

type Insert struct {
	requireExplicitFields

	Tablename string
	Model     Struct
	Returning bool
}

func (i Insert) rebind(dialect Dialect) (string, []any, error) {
	switch args := Named(i.Model).(type) {
	case invalidArg:
		return "", nil, args.error

	case *namedArgs:
		query, err := i.generateInsertQuery(dialect, args)
		if err != nil {
			return "", nil, err
		}

		return rebindQuery(dialect, query, args)

	}

	return "", nil, ErrInvalidArg
}

func (i *Insert) generateInsertQuery(dialect Dialect, args *namedArgs) (string, error) {
	var buffer bytes.Buffer

	columns := make([]string, 0, len(args.lookupMap))
	for name := range args.lookupMap {
		columns = append(columns, name)
	}

	sort.Strings(columns)

	buffer.WriteString("insert into ")
	buffer.WriteString(dialect.QuoteIdentifier(i.Tablename))
	buffer.WriteString(" (")

	for i, column := range columns {
		if i > 0 {
			buffer.WriteString(", ")
		}

		buffer.WriteString(dialect.QuoteIdentifier(column))
	}

	buffer.WriteString(") values (")

	for i, column := range columns {
		if i > 0 {
			buffer.WriteString(", ")
		}

		buffer.WriteString("@")
		buffer.WriteString(column)
	}

	buffer.WriteString(")")

	if i.Returning {
		buffer.WriteString(" returning *")
	}

	buffer.WriteString(" ;")
	return buffer.String(), nil
}
