package noorm

import (
	"database/sql"
	"reflect"
)

type scanner struct {
	rows        *sql.Rows
	columnNames []string
	columnIndex fieldLookupMap
}

func newScanner[T Struct](rows *sql.Rows) (*scanner, error) {
	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	columnIndex, err := buildFieldLookupMap[T]()
	if err != nil {
		return nil, err
	}

	s := scanner{
		rows:        rows,
		columnNames: columnNames,
		columnIndex: columnIndex,
	}

	return &s, nil
}

func (s scanner) close() error { return s.rows.Close() }
func (s scanner) next() bool   { return s.rows.Next() }
func (s scanner) err() error   { return s.rows.Err() }

func (s scanner) scan(target any) error {
	value := reflect.Indirect(reflect.ValueOf(target))

	targetSlice, err := buildScanTargetSlice(s.columnIndex, s.columnNames, value)
	if err != nil {
		return err
	}

	return s.rows.Scan(targetSlice...)
}
