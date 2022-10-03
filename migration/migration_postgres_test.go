//go:build integration && postgres

package migration

import (
	"context"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukasdietrich/groundwork/noorm"
)

func TestPostgresMigration(t *testing.T) {
	db, err := groundwork.Open("postgres", "user=groundwork dbname=groundwork password=groundwork sslmode=disable")
	require.NoError(t, err)

	defer db.Close()

	db.Exec(`
		drop table if exists "fruits" ;
		drop table if exists "database_changelog" ;
	`)

	changesets := []Changeset{
		LiteralChangeset("1", `
			create table "fruits" (
				"id"   serial primary key ,
				"name" varchar not null
			) ;
		`),
	}

	ctx := noorm.WithDatabase(context.Background(), db)
	applied, err := Up(ctx, changesets, nil)
	require.NoError(t, err)
	assert.Equal(t, changesets, applied)

	changesets = append(changesets, LiteralChangeset("2", `
			insert into "fruits" (
				"name"
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
