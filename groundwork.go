package groundwork

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrHashMismatch is returned when a changeset was already applied, but the queries changed.
	ErrHashMismatch = errors.New("hash mismatches already applied changelog entry")
)

// Changeset represents a series of related queries to advance the database schema.
type Changeset interface {
	// Name returns the name of the changeset.
	// It is used to keep track of already applied changesets.
	Name() string
	// Queries returns the content of this changeset as raw sql.
	Queries() (string, error)
}

// Changelog is an entity to keep track of already applied changesets.
type Changelog struct {
	// Name is the name of the changeset.
	Name string
	// Hash is the result of sha256(changeset.Queries())
	Hash string
	// Time is the time when a changeset was applied.
	Time time.Time
}

// Tx is database transaction related to a specific dialect.
type Tx interface {
	// Commit commits the transaction.
	Commit() error
	// Rollback rollbacks the transaction.
	Rollback() error

	// Exec executes an sql query without arguments.
	Exec(context.Context, string) error
	// Lookup returns the changelog for a given changeset.
	Lookup(context.Context, Changeset) (*Changelog, error)
	// Insert writes a new changelog entry to the database.
	Insert(context.Context, Changelog) error
}

// Dialect is a database agnostic adapter to apply changesets.
type Dialect interface {
	// Setup creates the changelog table.
	Setup(context.Context) error
	// Begin startes a new dialect-specific transaction.
	Begin(context.Context) (Tx, error)
}

// Up is a shorthand for UpContext using the background context.
func Up(dialect Dialect, changesets []Changeset) error {
	return UpContext(context.Background(), dialect, changesets)
}

// UpContext applies all changesets in order, if they have not already been applied.
// All changesets are applied within a transaction.
// When an error occurs, the process will stop, but previously applied changesets won't be rolled
// back.
// When a changeset was already applied, but does not match in content, an errors is returned.
func UpContext(ctx context.Context, dialect Dialect, changesets []Changeset) error {
	if err := dialect.Setup(ctx); err != nil {
		return err
	}

	for _, changeset := range changesets {
		if err := applyChangeset(ctx, dialect, changeset); err != nil {
			return fmt.Errorf("changeset %q: %w", changeset.Name(), err)
		}
	}

	return nil
}

func applyChangeset(ctx context.Context, dialect Dialect, changeset Changeset) error {
	queries, err := changeset.Queries()
	if err != nil {
		return err
	}

	hash := calculateHash(queries)

	tx, err := dialect.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	shouldApply, err := checkChangelog(ctx, tx, changeset, hash)
	if err != nil {
		return err
	}

	if shouldApply {
		if err := tx.Exec(ctx, queries); err != nil {
			return err
		}

		changelog := Changelog{
			Name: changeset.Name(),
			Hash: hash,
			Time: time.Now(),
		}

		if err := tx.Insert(ctx, changelog); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func checkChangelog(ctx context.Context, tx Tx, changeset Changeset, hash string) (bool, error) {
	changelog, err := tx.Lookup(ctx, changeset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true, nil
		}

		return false, err
	}

	if changelog == nil {
		return true, nil
	}

	if hash != changelog.Hash {
		return false, ErrHashMismatch
	}

	return false, nil
}

func calculateHash(s string) string {
	m := sha256.New()
	m.Write([]byte(s))

	h := m.Sum(nil)
	return hex.EncodeToString(h)
}
