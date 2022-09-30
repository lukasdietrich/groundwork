package migration

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukasdietrich/groundwork/v2/noorm"
)

func TestSqliteMigration(t *testing.T) {
	db, err := noorm.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	defer db.Close()

	changesets := []Changeset{
		LiteralChangeset("1", `
			create table "users" (
				"id"   integer primary key autoincrement ,
				"name" varchar not null
			) ;
		`),
	}

	ctx := noorm.WithDatabase(context.Background(), db)
	applied, err := Up(ctx, changesets, nil)
	require.NoError(t, err)
	assert.Equal(t, changesets, applied)

	changesets = append(changesets, LiteralChangeset("2", `
			create table "posts" (
				"id"      integer primary key autoincrement ,
				"user_id" integer not null ,
				"text"    text    not null ,

				foreign key ( "user_id" ) references "users" ( "id" )
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
