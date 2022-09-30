package migration

import (
	"embed"
	"io/fs"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/*
var testdata embed.FS

func TestFolderChangeset(t *testing.T) {
	files, err := fs.Sub(testdata, "testdata")
	require.NoError(t, err)

	changesets, err := ChangesetsFromFolder(files)
	require.NoError(t, err)
	assert.Len(t, changesets, 2)

	for i, expected := range []string{
		"01_first_file.sql",
		"02_second_file.sql",
	} {
		actual := changesets[i].Name()
		assert.Equal(t, expected, actual)
	}

	for i, expected := range []string{
		`create table "first_table" ( "first_column" integer ) ;`,
		`create table "second_table" ( "second_column" varchar ) ;`,
	} {
		actual, err := changesets[i].Queries()
		assert.NoError(t, err)
		assert.Equal(t, expected, strings.TrimSpace(actual))
	}
}
