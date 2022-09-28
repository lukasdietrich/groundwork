package noorm

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/suite"
)

type QuerierTestSuite struct {
	suite.Suite

	ctx context.Context
	db  *Database
}

func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(QuerierTestSuite))
}

func (s *QuerierTestSuite) SetupTest() {
	db, err := Open("sqlite3", ":memory:")
	s.Require().NoError(err)

	s.ctx = context.Background()
	s.db = db
}

func (s *QuerierTestSuite) AfterTest(_, _ string) {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *QuerierTestSuite) TestNoQuerier() {
	querier, dialect, err := QuerierFrom(s.ctx)
	s.Nil(querier)
	s.Nil(dialect)
	s.ErrorIs(err, ErrNoQuerierInContext)
}

func (s *QuerierTestSuite) TestBeginNoDatabase() {
	ctx, tx, err := Begin(s.ctx, nil)
	s.ErrorIs(err, ErrNoDatabaseInContext)
	s.Equal(s.ctx, ctx)
	s.Nil(tx)
}

func (s *QuerierTestSuite) TestWithDatabase() {
	ctx := WithDatabase(s.ctx, s.db)

	querier, dialect, err := QuerierFrom(ctx)
	s.NoError(err)
	s.Equal(s.db, querier)
	s.NotNil(dialect)
}

func (s *QuerierTestSuite) TestBegin() {
	ctx := WithDatabase(s.ctx, s.db)

	ctx, tx, err := Begin(ctx, nil)
	s.NoError(err)
	defer tx.Rollback()
	s.NotNil(tx)

	querier, dialect, err := QuerierFrom(ctx)
	s.NotEqual(s.db, querier)
	s.IsType(&transaction{}, querier)
	s.NotNil(dialect)
	s.NoError(err)
}
