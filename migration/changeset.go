package migration

import (
	"io/fs"
	"sort"
	"strings"
)

type literalChangeset struct {
	name    string
	queries string
}

func LiteralChangeset(name, queries string) Changeset {
	return literalChangeset{name, queries}
}

func (c literalChangeset) Name() string {
	return c.name
}

func (c literalChangeset) Queries() (string, error) {
	return c.queries, nil
}

type fileChangeset struct {
	filesystem fs.FS
	name       string
}

func (c fileChangeset) Name() string {
	return c.name
}

func (c fileChangeset) Queries() (string, error) {
	queries, err := fs.ReadFile(c.filesystem, c.name)
	return string(queries), err
}

func FileChangeset(filesystem fs.FS, name string) Changeset {
	return fileChangeset{
		filesystem: filesystem,
		name:       name,
	}
}

func ChangesetsFromFolder(filesystem fs.FS) ([]Changeset, error) {
	files, err := fs.ReadDir(filesystem, ".")
	if err != nil {
		return nil, err
	}

	var changesets []Changeset

	for _, file := range files {
		if isSqlFile(file) {
			changesets = append(changesets, FileChangeset(filesystem, file.Name()))
		}
	}

	sortChangesets(changesets)
	return changesets, nil
}

func isSqlFile(file fs.DirEntry) bool {
	return file.Type().IsRegular() && strings.HasSuffix(file.Name(), ".sql")
}

func sortChangesets(changeset []Changeset) {
	sort.Slice(changeset, func(i, j int) bool {
		return changeset[i].Name() < changeset[j].Name()
	})
}
