package noorm

import (
	"database/sql"
	"reflect"
)

type iterator[T Struct] struct {
	*sql.Rows
	columnNames []string
	columnIndex fieldLookupMap
}

func newIterator[T Struct](rows *sql.Rows) (Iterator[T], error) {
	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	columnIndex, err := buildFieldLookupMap[T]()
	if err != nil {
		return nil, err
	}

	iter := iterator[T]{
		Rows:        rows,
		columnNames: columnNames,
		columnIndex: columnIndex,
	}

	return &iter, nil
}

func (i iterator[T]) Value() (T, error) {
	var value T
	err := i.scanInto(&value)
	return value, err
}

func (i iterator[T]) scanInto(target *T) error {
	value := reflect.Indirect(reflect.ValueOf(target))

	targetSlice, err := buildScanTargetSlice(i.columnIndex, i.columnNames, value)
	if err != nil {
		return err
	}

	return i.Scan(targetSlice...)

}
