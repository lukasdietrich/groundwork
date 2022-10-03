package migration

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/lukasdietrich/groundwork/noorm"
)

var (
	// ErrHashMismatch is returned when a changeset was already applied, but the queries changed.
	ErrHashMismatch = errors.New("migration: hash does not match already applied version")
	// ErrNameTooLong is returned when a changeset name exceeds the limit.
	ErrNameTooLong = errors.New("migration: changeset name too long (max. 256)")
)

type Options struct {
	// ChangelogTablename is the name of the table used to track already applied changesets.
	ChangelogTablename string
}

func fillDefaults(opts *Options) *Options {
	if opts == nil {
		opts = &Options{}
	}

	if opts.ChangelogTablename == "" {
		opts.ChangelogTablename = "database_changelog"
	}

	return opts
}

// Changeset represents a series of related queries to advance the database schema.
type Changeset interface {
	// Name returns the name of the changeset.
	// It is used to keep track of already applied changesets.
	Name() string
	// Queries returns the content of this changeset as raw sql.
	Queries() (string, error)
}

// Up applies all changesets in order, if they have not already been applied.
// All changesets are applied within a transaction.
// When an error occurs, the process will stop, but previously applied changesets won't be rolled
// back.
// When a changeset was already applied, but does not match in content, an errors is returned.
// Up expects a groundwork/noorm.Querier to be present in the context (see WithDatabase).
func Up(ctx context.Context, changesets []Changeset, opts *Options) ([]Changeset, error) {
	opts = fillDefaults(opts)

	dao, err := newChangelogDao(ctx, opts.ChangelogTablename)
	if err != nil {
		return nil, err
	}

	if err := dao.setupTable(ctx); err != nil {
		return nil, err
	}

	return (&migrator{dao}).up(ctx, changesets)
}

type migrator struct {
	dao *changelogDao
}

func (m *migrator) up(ctx context.Context, changesets []Changeset) ([]Changeset, error) {
	var applied []Changeset

	for _, changeset := range changesets {
		wasApplied, err := m.apply(ctx, changeset)
		if err != nil {
			return applied, fmt.Errorf("%w on changeset %q", err, changeset.Name())
		}

		if wasApplied {
			applied = append(applied, changeset)
		}
	}

	return applied, nil
}

func (m *migrator) apply(ctx context.Context, changeset Changeset) (bool, error) {
	queries, err := changeset.Queries()
	if err != nil {
		return false, err
	}

	name := changeset.Name()
	if len(name) > nameSize {
		return false, ErrNameTooLong
	}

	hash := calculateHash(queries)

	ctx, tx, err := noorm.Begin(ctx, nil)
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	shouldApply, err := m.check(ctx, changeset, hash)
	if err != nil {
		return false, err
	}

	if shouldApply {
		if _, err := noorm.Exec(ctx, noorm.SQL{Query: queries}); err != nil {
			return false, err
		}

		entry := changelogEntry{
			Name: changeset.Name(),
			Hash: hash,
			Time: time.Now().Format(timeFormat),
		}

		if err := m.dao.insert(ctx, &entry); err != nil {
			return false, err
		}
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	return shouldApply, nil
}

func (m *migrator) check(ctx context.Context, changeset Changeset, hash string) (bool, error) {
	entry, err := m.dao.lookup(ctx, changeset.Name())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true, nil
		}

		return false, err
	}

	if entry == nil {
		return true, nil
	}

	if hash != entry.Hash {
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
