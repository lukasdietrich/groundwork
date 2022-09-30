package migration

import (
	"context"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukasdietrich/groundwork/v2/noorm"
)

func TestMysqlMigration(t *testing.T) {
	db, err := noorm.Open("mysql", "noorm:noorm@/noorm?multiStatements=true")
	require.NoError(t, err)

	defer db.Close()

	db.Exec(`
		drop table if exists posts ;
		drop table if exists users ;
		drop table if exists database_changelog ;
	`)

	changesets := []Changeset{
		LiteralChangeset("1", `
			create table users (
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
			create table posts (
				id      integer primary key auto_increment ,
				user_id integer not null ,
				text    text    not null ,

				foreign key ( user_id ) references users ( id )
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
