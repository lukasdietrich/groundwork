package noorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGuessDialect(t *testing.T) {
	for driverName, expectedDialect := range map[string]Dialect{
		"sqlite3":  sqliteDialect{},
		"postgres": postgresDialect{},
		"mysql":    mysqlDialect{},
		"sql":      defaultDialect{},
	} {
		actualDialect := guessDialect(driverName)
		assert.IsTypef(t, expectedDialect, actualDialect, "driverName=%q", driverName)
	}
}

func TestDialectPlaceholder(t *testing.T) {
	for dialect, expectedPlaceholders := range map[Dialect][]string{
		sqliteDialect{}:   {"?", "?"},
		postgresDialect{}: {"$1", "$2"},
		mysqlDialect{}:    {"?", "?"},
		defaultDialect{}:  {"?", "?"},
	} {
		actualPlaceholders := make([]string, 2)
		for i := range actualPlaceholders {
			actualPlaceholders[i] = dialect.Placeholder(i)
		}

		assert.Equal(t, expectedPlaceholders, actualPlaceholders)
	}
}

func TestDialectQuote(t *testing.T) {
	for dialect, expectedQuoted := range map[Dialect][]string{
		sqliteDialect{}:   {`"simple"`, `"with ""double"" quotes"`, "\"with `back` ticks\""},
		postgresDialect{}: {`"simple"`, `"with ""double"" quotes"`, "\"with `back` ticks\""},
		mysqlDialect{}:    {"`simple`", "`with \"double\" quotes`", "`with ``back`` ticks`"},
		defaultDialect{}:  {`"simple"`, `"with ""double"" quotes"`, "\"with `back` ticks\""},
	} {
		actualQuoted := []string{
			dialect.QuoteIdentifier("simple"),
			dialect.QuoteIdentifier(`with "double" quotes`),
			dialect.QuoteIdentifier("with `back` ticks"),
		}

		assert.Equal(t, expectedQuoted, actualQuoted)
	}
}
