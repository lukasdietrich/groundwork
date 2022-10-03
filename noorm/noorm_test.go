package noorm

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/suite"
)

type testStructUser struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type testStructPost struct {
	ID     int    `db:"id"`
	UserID int    `db:"user_id"`
	Text   string `db:"text"`
}

type SqliteTestSuite struct {
	suite.Suite

	db  *Database
	ctx context.Context
}

func TestSqliteTestSuite(t *testing.T) {
	suite.Run(t, new(SqliteTestSuite))
}

func (s *SqliteTestSuite) SetupTest() {
	db, err := Open("sqlite3", ":memory:")
	s.Require().NoError(err)

	_, err = db.Exec(`
		drop table if exists "users" ;

		create table "users" (
			"id"   integer primary key ,
			"name" varchar not null
		) ;

		insert into "users"
			( "name" )
		values
			( 'Foo' ),
			( 'Bar' ),
			( 'Baz' )
		;
	`)
	s.Require().NoError(err)

	ctx := context.Background()
	ctx = WithDatabase(ctx, db)

	s.db = db
	s.ctx = ctx
}

func (s *SqliteTestSuite) AfterTest() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *SqliteTestSuite) TestExec_Named() {
	_, err := Exec(s.ctx, SQL{
		Query: `insert into "users" ( "name" ) values ( @name ) ;`,
		Args:  Named(testStructUser{Name: "Tom"}),
	})

	s.Require().NoError(err)
}

func (s *SqliteTestSuite) TestExec_NamedPointer() {
	_, err := Exec(s.ctx, SQL{
		Query: `insert into "users" ( "name" ) values ( @name ) ;`,
		Args:  Named(&testStructUser{Name: "Jerry"}),
	})

	s.Require().NoError(err)
}

func (s *SqliteTestSuite) TestExec_Positional() {
	_, err := Exec(s.ctx, SQL{
		Query: `insert into "users" ( "name" ) values ( @0 ) ;`,
		Args:  Positional("Tom"),
	})

	s.Require().NoError(err)
}

func (s *SqliteTestSuite) TestQuery() {
	users, err := Query[testStructUser](s.ctx, SQL{
		Query: `
			select *
			from "users"
			where "name" in (@0) 
			order by "id" asc ;
		`,
		Args: Positional([]string{"Foo", "Baz"}),
	})

	s.Require().NoError(err)
	s.Equal([]testStructUser{{ID: 1, Name: "Foo"}, {ID: 3, Name: "Baz"}}, users)
}

func (s *SqliteTestSuite) TestQueryFirst() {
	user, err := QueryFirst[testStructUser](s.ctx, SQL{
		Query: `
			select *
			from "users"
			where "name" = @name
		`,
		Args: Named(testStructUser{Name: "Baz"}),
	})

	s.Require().NoError(err)
	s.Equal(&testStructUser{ID: 3, Name: "Baz"}, user)
}

func (s *SqliteTestSuite) TestIterate() {
	iterator, err := Iterate[testStructUser](s.ctx, SQL{
		Query: `
			select *
			from "users"
			where "name" in (@0) 
			order by "id" asc ;
		`,
		Args: Positional([]string{"Foo", "Baz"}),
	})

	s.Require().NoError(err)

	defer iterator.Close()

	s.Require().True(iterator.Next())
	s.Require().NoError(iterator.Err())

	user1, err := iterator.Value()
	s.Require().NoError(err)
	s.Equal(testStructUser{ID: 1, Name: "Foo"}, user1)

	s.Require().True(iterator.Next())
	s.Require().NoError(iterator.Err())

	user2, err := iterator.Value()
	s.Require().NoError(err)
	s.Equal(testStructUser{ID: 3, Name: "Baz"}, user2)

	s.False(iterator.Next())
}
