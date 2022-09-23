package noorm

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
	"unicode"
	"unicode/utf8"
)

// invalidArg is a pseudo ArgsSource implementation when a provided value is not permitted.
type invalidArg struct {
	error
}

func (a invalidArg) arg(string) (any, error) {
	return nil, a.error
}

type noneArgs struct{}

// None is used when you do not need to provide any arguments for a query.
func None() ArgumentSource {
	return noneArgs{}
}

func (noneArgs) arg(name string) (any, error) {
	return nil, fmt.Errorf("noorm: no args provided: %q", name)
}

type namedArgs struct {
	lookupMap fieldLookupMap
	value     reflect.Value
}

// Named uses the fields of a struct as named arguments for a query.
// Field names can be overwritten with struct tags.
func Named[T Struct](args T) ArgumentSource {
	lookupMap, err := buildFieldLookupMap[T]()
	if err != nil {
		return invalidArg{err}
	}

	return &namedArgs{
		lookupMap: lookupMap,
		value:     reflect.Indirect(reflect.ValueOf(&args)),
	}
}

func (a *namedArgs) arg(name string) (any, error) {
	index, ok := a.lookupMap[name]
	if !ok {
		return nil, fmt.Errorf("noorm: named argument is not in struct: %q", name)
	}

	return reflect.Indirect(a.value).FieldByIndex(index).Interface(), nil
}

type positionalArgs []any

// Positional uses the positional index of the the provided args as their name in a query.
// The index starts counting at 0.
func Positional(args ...any) ArgumentSource {
	return positionalArgs(args)
}

func (a positionalArgs) arg(name string) (any, error) {
	i, err := strconv.Atoi(name)
	if err != nil {
		return nil, fmt.Errorf("noorm: positional argument is not a number: %q", name)
	}

	if i < 0 || i >= len(a) {
		return nil, fmt.Errorf("noorm: positional argument is out of range: 0 <= %s < %d", name, len(a))
	}

	return a[i], nil
}

var _ driver.Valuer = nullableValue{}

type nullableValue struct {
	value reflect.Value
}

func (n nullableValue) Value() (driver.Value, error) {
	if n.value.Kind() == reflect.Pointer && n.value.IsNil() {
		return nil, nil
	}

	return n.value.Interface(), nil
}

// rebindQuery parses the query and replaces named paremeters with the database specific
// placeholder. Named parameters have the form `@name` where `name` is the actual name.
// A literal `@` can be written by doubling it `@@`. Only letters, numbers, dashes and underscores
// are permitted as names.
func rebindQuery(query, placeholder string, args ArgumentSource) (string, []any, error) {
	const at = '@'

	var (
		queryBuffer    bytes.Buffer
		nameBuffer     bytes.Buffer
		parameterSlice []any
		byteOffset     int
		inParam        bool
	)

	for _, r := range query {
		byteOffset += utf8.RuneLen(r)
		endOfQuery := byteOffset >= len(query)
		wasInParam := inParam

		if inParam {
			if !isParameterNameRune(r) {
				if nameBuffer.Len() == 1 {
					wasInParam = false

					if r != at {
						nameBuffer.WriteTo(&queryBuffer)
					}

					nameBuffer.Reset()
				}

				inParam = false
			}
		} else {
			if r == at {
				inParam = true
			}
		}

		if inParam {
			nameBuffer.WriteRune(r)
		}

		if wasInParam && (!inParam || endOfQuery) {
			name := nameBuffer.String()
			nameBuffer.Reset()

			arg, err := args.arg(name[1:])
			if err != nil {
				return query, nil, err
			}

			argValues := splitArg(arg)
			parameterSlice = append(parameterSlice, argValues...)
			repeatPlaceholder(&queryBuffer, placeholder, len(argValues))
		}

		if !inParam {
			queryBuffer.WriteRune(r)
		}
	}

	return queryBuffer.String(), parameterSlice, nil
}

func isParameterNameRune(r rune) bool {
	return unicode.IsDigit(r) || unicode.IsLetter(r) || r == '-' || r == '_'
}

func repeatPlaceholder(buffer *bytes.Buffer, placeholder string, n int) {
	for i := 0; i < n; i++ {
		if i > 0 {
			buffer.WriteString(", ")
		}

		buffer.WriteString(placeholder)
	}
}

func splitArg(arg any) []any {
	v := reflect.ValueOf(arg)

	if v.Kind() != reflect.Slice {
		return []any{arg}
	}

	argSlice := make([]any, v.Len())
	for i := 0; i < len(argSlice); i++ {
		argSlice[i] = v.Index(i).Interface()
	}

	return argSlice
}
