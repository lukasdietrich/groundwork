//go:build integration && mysql

package noorm

import (
	"context"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/suite"
)

type MysqlTestSuite struct {
	suite.Suite

	db  *Database
	ctx context.Context
}

func TestMysqlTestSuite(t *testing.T) {
	suite.Run(t, new(MysqlTestSuite))
}

func (s *MysqlTestSuite) SetupTest() {
	db, err := Open("mysql", "noorm:noorm@/noorm?multiStatements=true")
	s.Require().NoError(err)

	_, err = db.Exec(`
		drop table if exists users ;

		create table users (
			id   integer primary key auto_increment ,
			name varchar ( 64 ) not null
		) ;

		insert into users
			( id, name )
		values
			( null, "Foo" ),
			( null, "Bar" ),
			( null, "Baz" )
		;
	`)
	s.Require().NoError(err)

	ctx := context.Background()
	ctx = WithDatabase(ctx, db)

	s.db = db
	s.ctx = ctx
}

func (s *MysqlTestSuite) AfterTest(_, _ string) {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *MysqlTestSuite) TestExec_Named() {
	_, err := Exec(s.ctx,
		`
			insert into users ( name ) values ( @name ) ;
		`,
		Named(testStructUser{
			Name: "Tom",
		}))
	s.Require().NoError(err)
}

func (s *MysqlTestSuite) TestExec_NamedPointer() {
	_, err := Exec(s.ctx,
		`
			insert into users ( name ) values ( @name ) ;
		`,
		Named(&testStructUser{
			Name: "Jerry",
		}))
	s.Require().NoError(err)
}

func (s *MysqlTestSuite) TestExec_Positional() {
	_, err := Exec(s.ctx,
		`
			insert into users ( name ) values ( @0 ) ;
		`,
		Positional("Tom"))
	s.Require().NoError(err)
}

func (s *MysqlTestSuite) TestQuery() {
	users, err := Query[testStructUser](s.ctx,
		`
			select *
			from users
			where name in (@0) 
			order by id asc ;
		`,
		Positional([]string{"Foo", "Baz"}))

	s.Require().NoError(err)
	s.Equal([]testStructUser{{ID: 1, Name: "Foo"}, {ID: 3, Name: "Baz"}}, users)
}

func (s *MysqlTestSuite) TestIterate() {
	iterator, err := Iterate[testStructUser](s.ctx,
		`
			select *
			from users
			where name in (@0) 
			order by id asc ;
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
