package sqlite3

import (
	"fmt"
	"reflect"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Results struct {
	st      *Statement
	err     error
	cols    []interface{}
	rowid   int64
	changes int
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	typeText = reflect.TypeOf("")
	typeBlob = reflect.TypeOf([]byte{})
	typeTime = reflect.TypeOf(time.Time{})
)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r *Results) String() string {
	str := "<results"
	if r.rowid != 0 {
		str += fmt.Sprintf(" lastinsertid=%v", r.rowid)
	}
	if r.changes != 0 {
		str += fmt.Sprintf(" rowsaffected=%v", r.changes)
	}
	if r.st != nil {
		str += " " + r.st.String()
	}
	if r.err != nil && r.err != SQLITE_ROW {
		str += fmt.Sprintf(" err=%q", r.err.Error())
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

// Return a new results object
func results(st *Statement, err error) *Results {
	r := new(Results)
	r.st = st
	r.err = err
	r.cols = make([]interface{}, 0, st.ColumnCount())
	r.rowid = st.Conn().LastInsertId()
	r.changes = st.Conn().Changes()
	return r
}

// Return next row of values, or nil if there are no more rows.
// If arguments t are provided, then the values will be
// cast to the types in t if that is possible
func (r *Results) Next(t ...reflect.Type) []interface{} {
	// If no more results, return nil,io.EOF
	if r.err == SQLITE_DONE {
		r.st.Reset()
		r.st = nil
		r.cols = nil
		return nil
	}

	// Check for SQLITE_ROW result, abort result if error occurred
	if r.err != SQLITE_ROW {
		r.st.Reset()
		r.st = nil
		r.cols = nil
		return nil
	}

	// Adjust size of columns
	n := r.st.ColumnCount()
	r.cols = r.cols[:n]

	// Cast values into columns. If type t is defined also cast
	// value to type t
	for i := 0; i < n; i++ {
		if i < len(t) && t[i] != nil {
			if v, err := r.castvalue(i, t[i]); err == nil {
				r.cols[i] = v
				continue
			}
		}
		r.cols[i] = r.value(i)
	}

	// Call step to next row
	r.err = r.st.Step()

	// Return result
	return r.cols
}

func (r *Results) LastInsertId() int64 {
	return r.rowid
}

func (r *Results) RowsAffected() int {
	return r.changes
}

// Return the expanded SQL statement
func (r *Results) ExpandedSQL() string {
	if r.st == nil {
		return ""
	} else {
		return r.st.ExpandedSQL()
	}
}

// Return column count
func (r *Results) ColumnCount() int {
	return r.st.ColumnCount()
}

// Return column name
func (r *Results) ColumnName(i int) string {
	if r.st == nil {
		return ""
	}
	return r.st.ColumnName(i)
}

// Return column type
func (r *Results) ColumnType(i int) Type {
	if r.st == nil {
		return SQLITE_NULL
	}
	return r.st.ColumnType(i)
}

// Return column decltype
func (r *Results) ColumnDeclType(i int) string {
	if r.st == nil {
		return ""
	}
	return r.st.ColumnDeclType(i)
}

// Return the source database schema name
func (r *Results) ColumnDatabaseName(i int) string {
	if r.st == nil {
		return ""
	}
	return r.st.ColumnDatabaseName(i)
}

// Return the source table name
func (r *Results) ColumnTableName(i int) string {
	if r.st == nil {
		return ""
	}
	return r.st.ColumnTableName(i)
}

// Return the origin
func (r *Results) ColumnOriginName(i int) string {
	if r.st == nil {
		return ""
	}
	return r.st.ColumnOriginName(i)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (r *Results) value(index int) interface{} {
	return r.st.ColumnInterface(index)
}

func (r *Results) castvalue(index int, t reflect.Type) (interface{}, error) {
	st := r.st.ColumnType(index)

	// Do NULL cases
	if st == SQLITE_NULL {
		return reflect.Zero(t).Interface(), nil
	}

	// Do simple cases first
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		rv := reflect.ValueOf(r.st.ColumnInt64(index))
		if rv.CanConvert(t) {
			return rv.Convert(t).Interface(), nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		rv := reflect.ValueOf(r.st.ColumnInt64(index))
		if rv.CanConvert(t) {
			return rv.Convert(t).Interface(), nil
		}
	case reflect.Bool:
		if r.st.ColumnInt64(index) == 0 {
			return false, nil
		} else {
			return true, nil
		}
	case reflect.String:
		return r.st.ColumnText(index), nil
	case reflect.Float32, reflect.Float64:
		rv := reflect.ValueOf(r.st.ColumnDouble(index))
		if rv.CanConvert(t) {
			return rv.Convert(t).Interface(), nil
		}
	}
	// Do types
	switch t {
	case typeTime:
		if st == SQLITE_TEXT {
			return time.Parse(time.RFC3339, r.st.ColumnText(index))
		} else if st == SQLITE_FLOAT {
			return nil, fmt.Errorf("Cannot convert julian day number to time (at this time)")
		} else if st == SQLITE_INTEGER {
			return time.Unix(r.st.ColumnInt64(index), 0), nil
		}
	case typeBlob:
		if st == SQLITE_BLOB {
			return r.st.ColumnBlob(index, true), nil
		} else if st == SQLITE_TEXT {
			return []byte(r.st.ColumnText(index)), nil
		}
	}

	// No conversion possible
	return nil, fmt.Errorf("Cannot convert %q to %q", r.st.ColumnType(index), t)
}
