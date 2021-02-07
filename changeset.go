package groundwork

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
