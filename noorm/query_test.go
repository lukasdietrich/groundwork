package noorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsert(t *testing.T) {
	type TestStruct struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
	}

	testStruct := TestStruct{
		ID:   123,
		Name: "Tester",
	}

	insert := Insert[TestStruct]{
		Tablename: "table",
		Model:     testStruct,
	}

	for dialect, expectedQuery := range map[Dialect]string{
		sqliteDialect{}: `insert into "table" ("id", "name") values (?, ?) ;`,
		mysqlDialect{}:  "insert into `table` (`id`, `name`) values (?, ?) ;",
	} {
		actualQuery, params, err := insert.rebind(dialect)
		assert.NoError(t, err)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Equal(t, []any{int64(123), "Tester"}, params)
	}
}
