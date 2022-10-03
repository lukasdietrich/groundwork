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

	columnName = "name"
	columnHash = "hash"
	columnTime = "time"
)

type changelogEntry struct {
	Name string `db:"name"`
	Hash string `db:"hash"`
	Time string `db:"time"`
}

type changelogDao struct {
	dialect   noorm.Dialect
	tablename string
}

func newChangelogDao(ctx context.Context, tablename string) (*changelogDao, error) {
	_, dialect, err := noorm.QuerierFrom(ctx)
	if err != nil {
		return nil, err
	}

	dao := changelogDao{
		dialect:   dialect,
		tablename: tablename,
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
		c.dialect.QuoteIdentifier(c.tablename),
		c.dialect.QuoteIdentifier(columnName),
		c.dialect.QuoteIdentifier(columnHash),
		c.dialect.QuoteIdentifier(columnTime),

		nameSize,
		hashSize,
		timeSize,
	)

	_, err := noorm.Exec(ctx, noorm.SQL{Query: schema})
	return err
}

func (c *changelogDao) lookup(ctx context.Context, name string) (*changelogEntry, error) {
	query := fmt.Sprintf(`
		select *
		from %[1]s
		where %[2]s = @0 ;
	`,
		c.dialect.QuoteIdentifier(c.tablename),
		c.dialect.QuoteIdentifier(columnName),
	)

	return noorm.QueryFirst[changelogEntry](ctx, noorm.SQL{
		Query: query,
		Args:  noorm.Positional(name),
	})
}

func (c *changelogDao) insert(ctx context.Context, entry *changelogEntry) error {
	_, err := noorm.Exec(ctx, noorm.Insert[*changelogEntry]{
		Tablename: c.tablename,
		Model:     entry,
	})

	return err
}
