package noorm

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildFieldLookupMap(t *testing.T) {
	type EmbeddedTestStruct struct {
		Field4 int
	}

	type TestStructComplex struct {
		Field1 string
		Field2 string  `db:"field_2"`
		Field3 *string `db:"field_3,primary"`
		EmbeddedTestStruct
	}

	index, err := buildFieldLookupMap[TestStructComplex]()
	require.NoError(t, err)
	assert.Len(t, index, 4)
	assert.Equal(t, fieldLookupMap{
		"Field1":  []int{0},
		"field_2": []int{1},
		"field_3": []int{2},
		"Field4":  []int{3, 0},
	}, index)
}

func TestInitializeFieldPath(t *testing.T) {
	type TestStruct struct {
		Field1 struct {
			Field2 *struct {
				Field3 *struct {
					Field4 int
					Field5 *string
				}
			}
		}
	}

	var s TestStruct
	initializeFieldPath(reflect.ValueOf(&s), []int{0, 0, 0, 1})

	require.NotNil(t, s.Field1.Field2)
	require.NotNil(t, s.Field1.Field2.Field3)
}
