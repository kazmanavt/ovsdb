package schema

import (
	"fmt"
	"reflect"
	"sync"
)

// TableSchema represents a table schema, according to RFC7047.
// A JSON object with the following members:
//
//	"columns": {<id>: <column-schema>, ...}   required
//	"maxRows": <integer>                      optional
//	"isRoot": <boolean>                       optional
//	"indexes": [<column-set>*]                optional
type TableSchema struct {
	Name    string
	Columns map[string]*ColumnSchema `json:"columns"`           // columns in the table
	MaxRows int                      `json:"maxRows,omitempty"` // maximum number of rows in the table
	IsRoot  bool                     `json:"isRoot,omitempty"`  // true if the table rows are part of a root-set (they are not cleaned by garbage collection)
	Indexes [][]string               `json:"indexes,omitempty"` // indexes in the table
}

// NewRow creates a new row with the given values.
// The values are given as a list of key-value pairs.
// The keys are the column names and the values are the column values.
// The keys must be strings, NewRow will panic otherwise.
// The values must be pointers to the column type, NewRow will panic otherwise.
// For example:
//
//	row := table.NewRow("name", &"foo", "age", &42)
func (ts *TableSchema) NewRow(args ...any) Row {
	r := rowImpl{
		tSch: ts,
		row:  make(map[string]any),
		mu:   &sync.RWMutex{},
	}

	for i := 0; i+1 < len(args); i += 2 {
		cName, ok := args[i].(string)
		if !ok {
			panic("column name must be a string")
		}
		if _, ok := ts.Columns[cName]; !ok {
			panic(fmt.Sprintf("column %q does not exist in table %q", cName, ts.Name))
		}
		var val any
		switch {
		case i+1 >= len(args):
			val = ts.Columns[cName].GetDefaultValue()
		case args[i+1] == nil:
			val = nil
		case reflect.ValueOf(args[i+1]).Kind() == reflect.Ptr:
			val = reflect.ValueOf(args[i+1]).Elem().Interface()
			//panic("column value must not be a pointer")
		default:
			val = args[i+1]
		}
		if err := ts.Columns[cName].ValidateValue(val); err != nil {
			panic(fmt.Sprintf("column %q of table %q: %s", cName, ts.Name, err))
		}
		r.row[cName] = val
	}

	return &r
}

func (ts *TableSchema) NewRows() Rows {
	return Rows{
		tSch: ts,
		Rows: make([]Row, 0),
	}
}
