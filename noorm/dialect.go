package noorm

import (
	"database/sql"
	"reflect"
	"strconv"
	"strings"
)

type Dialect interface {
	Placeholder(position int) string
	QuoteIdentifier(identifier string) string
}

func guessDialect(db *sql.DB) Dialect {
	driver := db.Driver()
	driverType := reflect.TypeOf(driver)
	if driverType.Kind() == reflect.Pointer {
		driverType = driverType.Elem()
	}

	switch driverType.PkgPath() {
	case "github.com/mattn/go-sqlite3":
		return sqliteDialect{}

	case "github.com/lib/pq", "github.com/jackc/pgx":
		return postgresDialect{}

	case "github.com/go-sql-driver/mysql":
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
