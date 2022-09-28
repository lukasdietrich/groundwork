package noorm

import (
	"strconv"
	"strings"
)

// Dialect provides database specific sql query helpers.
type Dialect interface {
	// Placeholder returns a positional argument placeholder.
	// The parameter `position` is the index of the parameter starting at 0.
	Placeholder(position int) string
	// QuoteIdentifier quotes an identifier (eg. column or table name).
	QuoteIdentifier(identifier string) string
}

func guessDialect(driverName string) Dialect {
	switch strings.ToLower(driverName) {
	case "sqlite3":
		return sqliteDialect{}

	case "postgres":
		return postgresDialect{}

	case "mysql":
		return mysqlDialect{}

	default:
		return defaultDialect{}
	}
}

type defaultDialect struct{}

func (defaultDialect) Placeholder(int) string {
	return "?"
}

func (defaultDialect) QuoteIdentifier(identifier string) string {
	return `"` + strings.Replace(identifier, `"`, `""`, -1) + `"`
}

type sqliteDialect struct {
	defaultDialect
}

type postgresDialect struct {
	defaultDialect
}

func (postgresDialect) Placeholder(position int) string {
	return "$" + strconv.Itoa(position+1)
}

type mysqlDialect struct {
	defaultDialect
}

func (mysqlDialect) QuoteIdentifier(identifier string) string {
	return "`" + strings.Replace(identifier, "`", "``", -1) + "`"
}
