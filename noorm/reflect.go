package noorm

import (
	"fmt"
	"reflect"
	"strings"
)

func typeOfGeneric[T any]() reflect.Type {
	var t *T
	return reflect.TypeOf(t).Elem()
}

type fieldLookupMap map[string][]int // name -> path of field indices

func buildFieldLookupMap[T Struct]() (fieldLookupMap, error) {
	t := typeOfGeneric[T]()
	lookup := make(fieldLookupMap)

	return lookup, analyzeStructFields(lookup, nil, t)
}

func analyzeStructFields(lookup fieldLookupMap, prefix []int, t reflect.Type) error {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return fmt.Errorf("cannot traverse %q, expected a struct", t)
	}

	for i := t.NumField() - 1; i >= 0; i-- {
		field := t.Field(i)

		if !field.IsExported() {
			// skip unexported fields
			continue
		}

		index := appendFieldIndex(prefix, i)

		if field.Anonymous {
			if err := analyzeStructFields(lookup, index, field.Type); err != nil {
				return err
			}
		} else {
			name := fieldName(field)

			if _, ok := lookup[name]; ok {
				return fmt.Errorf("duplicate struct field %q", field.Name)
			}

			lookup[name] = index
		}
	}

	return nil
}

func appendFieldIndex(prefix []int, i int) []int {
	path := make([]int, len(prefix)+1)
	copy(path, prefix)
	path[len(path)-1] = i
	return path
}

func fieldName(field reflect.StructField) string {
	if tag, ok := field.Tag.Lookup("db"); ok {
		if comma := strings.IndexByte(tag, ','); comma > -1 {
			tag = tag[:comma]
		}

		return tag
	}

	return field.Name
}

func initializeFieldPath(v reflect.Value, index []int) {
	for _, i := range index {
		field := reflect.Indirect(v).Field(i)

		if field.Kind() == reflect.Pointer && field.IsNil() {
			t := field.Type().Elem()
			field.Set(reflect.New(t))
		}

		v = field
	}
}

func buildScanTargetSlice(lookup fieldLookupMap, columns []string, v reflect.Value) ([]any, error) {
	targetSlice := make([]any, len(columns))

	for i, column := range columns {
		if index, ok := lookup[column]; ok {
			initializeFieldPath(v, index)
			field := v.FieldByIndex(index)
			targetSlice[i] = field.Addr().Interface()
		} else {
			var discard any
			targetSlice[i] = &discard
		}
	}

	return targetSlice, nil
}
