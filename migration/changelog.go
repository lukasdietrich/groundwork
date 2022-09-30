package migration

import (
	"context"
	"fmt"
	"time"

	"github.com/lukasdietrich/groundwork/noorm"
)

const (
	nameSize = 256
	// hex(sha256)
	hashSize = 64
	// we store the time as text, because a varchar works in all major databases the same way.
	timeFormat = time.RFC3339Nano
	timeSize   = len(timeFormat)
)

type changelogEntry struct {
	Name string `db:"name"`
	Hash string `db:"hash"`
	Time string `db:"time"`
}

type changelogDao struct {
	identTable string
	identName  string
	identHash  string
	identTime  string
}

func newChangelogDao(ctx context.Context, tablename string) (*changelogDao, error) {
	_, dialect, err := noorm.QuerierFrom(ctx)
	if err != nil {
		return nil, err
	}

	dao := changelogDao{
		identTable: dialect.QuoteIdentifier(tablename),
		identName:  dialect.QuoteIdentifier("name"),
		identHash:  dialect.QuoteIdentifier("hash"),
		identTime:  dialect.QuoteIdentifier("time"),
	}

	return &dao, nil
}

func (c *changelogDao) setupTable(ctx context.Context) error {
	schema := fmt.Sprintf(`
		create table if not exists %[1]s (
			%[2]s varchar ( %[5]d ) not null ,
			%[3]s varchar ( %[6]d ) not null ,
			%[4]s varchar ( %[7]d ) not null ,

			primary key ( %[2]s )
		) ;
	`,
		c.identTable,
		c.identName,
		c.identHash,
		c.identTime,

		nameSize,
		hashSize,
		timeSize,
	)

	_, err := noorm.Exec(ctx, schema, noorm.None())
	return err
}

func (c *changelogDao) lookup(ctx context.Context, name string) (*changelogEntry, error) {
	query := fmt.Sprintf(`
		select *
		from %[1]s
		where %[2]s = @0 ;
	`, c.identTable, c.identName)

	return noorm.QueryFirst[changelogEntry](ctx, query, noorm.Positional(name))
}

func (c *changelogDao) insert(ctx context.Context, entry *changelogEntry) error {
	query := fmt.Sprintf(`
		insert into %[1]s (
			%[2]s ,
			%[3]s ,
			%[4]s
		) values (
			@name ,
			@hash ,
			@time
		) ;
	`, c.identTable, c.identName, c.identHash, c.identTime)

	_, err := noorm.Exec(ctx, query, noorm.Named(entry))
	return err
}
