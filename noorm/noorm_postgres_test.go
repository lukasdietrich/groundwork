//go:build integration && postgres

package noorm

import (
	"context"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

type PostgresTestSuite struct {
	suite.Suite

	db  *Database
	ctx context.Context
}

func TestPostgresTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresTestSuite))
}

func (s *PostgresTestSuite) SetupTest() {
	db, err := Open("postgres", "user=noorm dbname=noorm password=noorm sslmode=disable")
	s.Require().NoError(err)

	_, err = db.Exec(`
		drop table if exists "users" ;

		create table "users" (
			"id"   serial primary key ,
			"name" varchar ( 64 ) not null
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

func (s *PostgresTestSuite) AfterTest(_, _ string) {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *PostgresTestSuite) TestExec_Named() {
	_, err := Exec(s.ctx,
		`
			insert into "users" ( "name" ) values ( @name ) ;
		`,
		Named(testStructUser{
			Name: "Tom",
		}))
	s.Require().NoError(err)
}

func (s *PostgresTestSuite) TestExec_NamedPointer() {
	_, err := Exec(s.ctx,
		`
			insert into "users" ( "name" ) values ( @name ) ;
		`,
		Named(&testStructUser{
			Name: "Jerry",
		}))
	s.Require().NoError(err)
}

func (s *PostgresTestSuite) TestExec_Positional() {
	_, err := Exec(s.ctx,
		`
			insert into "users" ( "name" ) values ( @0 ) ;
		`,
		Positional("Tom"))
	s.Require().NoError(err)
}

func (s *PostgresTestSuite) TestQuery() {
	users, err := Query[testStructUser](s.ctx,
		`
			select *
			from "users"
			where "name" in (@0) 
			order by "id" asc ;
		`,
		Positional([]string{"Foo", "Baz"}))

	s.Require().NoError(err)
	s.Equal([]testStructUser{{ID: 1, Name: "Foo"}, {ID: 3, Name: "Baz"}}, users)
}

func (s *PostgresTestSuite) TestQueryFirst() {
	user, err := QueryFirst[testStructUser](s.ctx,
		`
			select *
			from "users"
			where "name" = @name
		`,
		Named(testStructUser{Name: "Baz"}))

	s.Require().NoError(err)
	s.Equal(&testStructUser{ID: 3, Name: "Baz"}, user)
}

func (s *PostgresTestSuite) TestIterate() {
	iterator, err := Iterate[testStructUser](s.ctx,
		`
			select *
			from "users"
			where "name" in (@0) 
			order by "id" asc ;
		`,
		Positional([]string{"Foo", "Baz"}))

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
