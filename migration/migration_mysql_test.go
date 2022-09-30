//go:build integration && mysql

package migration

import (
	"context"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukasdietrich/groundwork/noorm"
)

func TestMysqlMigration(t *testing.T) {
	db, err := noorm.Open("mysql", "noorm:noorm@/noorm?multiStatements=true")
	require.NoError(t, err)

	defer db.Close()

	db.Exec(`
		drop table if exists fruits ;
		drop table if exists database_changelog ;
	`)

	changesets := []Changeset{
		LiteralChangeset("1", `
			create table fruits (
				id   integer primary key auto_increment ,
				name varchar ( 64 ) not null
			) ;
		`),
	}

	ctx := noorm.WithDatabase(context.Background(), db)
	applied, err := Up(ctx, changesets, nil)
	require.NoError(t, err)
	assert.Equal(t, changesets, applied)

	changesets = append(changesets, LiteralChangeset("2", `
			insert into fruits (
				name
			) values (
				'Orange'
			) ;
		`))

	applied, err = Up(ctx, changesets, nil)
	require.NoError(t, err)
	assert.Equal(t, changesets[1:], applied)

	applied, err = Up(ctx, changesets, nil)
	require.NoError(t, err)
	assert.Empty(t, applied)

	changesets[1] = LiteralChangeset("2", "something else")
	applied, err = Up(ctx, changesets, nil)
	require.ErrorIs(t, err, ErrHashMismatch)
}
