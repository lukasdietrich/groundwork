package noorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPositional(t *testing.T) {
	args := Positional(1, 2, 3, 4, 5)
	assert.NotNil(t, args)

	for _, tc := range []struct {
		name string
		err  bool
		arg  any
	}{
		{name: "0", arg: 1},
		{name: "4", arg: 5},
		{name: "-1", err: true},
		{name: "5", err: true},
		{name: "NaN", err: true},
	} {
		arg, err := args.arg(tc.name)
		if tc.err {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.arg, arg)
		}
	}
}

func TestNamed(t *testing.T) {
	type TestStruct struct {
		Field1 string
		Field2 int `db:"field_2"`
	}

	args := Named(&TestStruct{
		Field1: "test",
		Field2: 420,
	})
	assert.NotNil(t, args)

	for _, tc := range []struct {
		name string
		err  bool
		arg  any
	}{
		{name: "Field1", arg: "test"},
		{name: "field_2", arg: 420},
		{name: "Field3", err: true},
	} {
		arg, err := args.arg(tc.name)
		if tc.err {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.arg, arg)
		}
	}
}

func TestRebind(t *testing.T) {
	type expected struct {
		query      string
		parameters []any
	}

	type input struct {
		query string
		args  ArgumentSource
	}

	for _, tc := range []struct {
		input
		expected
	}{
		{
			input{
				query: `select * from t where a = @0 and b in (@1) ;`,
				args:  Positional(1, []int{1, 2, 3}),
			},
			expected{
				query:      `select * from t where a = ? and b in (?, ?, ?) ;`,
				parameters: []any{1, 1, 2, 3},
			},
		},
		{
			input{
				query: `select * from t where a = @param_a and b in (@ParamB) ;`,
				args: Named(struct {
					ParamA string `db:"param_a"`
					ParamB []int
				}{
					ParamA: "a",
					ParamB: []int{1, 2, 3},
				}),
			},
			expected{
				query:      `select * from t where a = ? and b in (?, ?, ?) ;`,
				parameters: []any{"a", 1, 2, 3},
			},
		},
		{
			input{
				query: `select * from t where a = @0 and b like @1`,
				args:  Positional(1, 2),
			},
			expected{
				query:      `select * from t where a = ? and b like ?`,
				parameters: []any{1, 2},
			},
		},
		{
			input{
				query: `select * from t where a = @2 and b = @0`,
				args:  Positional(1, 2, 3, 4),
			},
			expected{
				query:      `select * from t where a = ? and b = ?`,
				parameters: []any{3, 1},
			},
		},
		{
			input{
				query: `select * from t where a = '@@' and b = '@ hello' ;`,
				args:  Positional(1, 2, 3, 4),
			},
			expected{
				query:      `select * from t where a = '@' and b = '@ hello' ;`,
				parameters: nil,
			},
		},
	} {
		query, args, err := rebindQuery(tc.input.query, defaultDialect{}, tc.input.args)

		require.NoError(t, err)
		assert.Equal(t, tc.expected.query, query)
		assert.Equal(t, tc.expected.parameters, args)
	}
}
