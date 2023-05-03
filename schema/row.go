package schema

import (
	"encoding/json"
	"fmt"
	"github.com/kazmanavt/ovsdb/types"
	"reflect"
	"sync"
)

type Rows struct {
	tSch *TableSchema
	Rows []Row
}

func (rs *Rows) UnmarshalJSON(data []byte) error {
	var _rows []json.RawMessage
	if err := json.Unmarshal(data, &_rows); err != nil {
		return err
	}
	for _, _row := range _rows {
		row := rs.tSch.NewRow()
		if err := row.UnmarshalJSON(_row); err != nil {
			return err
		}
		rs.Rows = append(rs.Rows, row)
	}
	return nil
}

type Row interface {
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
	TableName() string
	TableSchema() *TableSchema
	// Get returns the value of the column cName of table Row.
	// If the column is not set, the default value is returned.
	// If the column is not in the table, it panics.
	Get(cName string) any
	// Set sets the value of the column cName of table Row to value.
	// If the column is not in the table, or type of value violate schema, Set will panic.
	Set(cName string, value any)
	// Update2 updates the row with the modify part of tables-update2 response.
	Update2(diff Row) error
	// Match returns true if the row matches the conditions.
	Match(where []types.Condition) bool
}

type rowImpl struct {
	tSch *TableSchema
	mu   *sync.RWMutex
	row  map[string]any
}

func (r rowImpl) MarshalJSON() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return json.Marshal(r.row)
}

func (r *rowImpl) UnmarshalJSON(data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	var row map[string]json.RawMessage
	if err := json.Unmarshal(data, &row); err != nil {
		return err
	}
	for cName, jsonValue := range row {
		cSch, ok := r.tSch.Columns[cName]
		if !ok {
			return fmt.Errorf("column %q not in table %q", cName, r.tSch.Name)
		}
		value := cSch.GetDefaultValue()
		ptr := reflect.New(reflect.TypeOf(value))
		if err := json.Unmarshal(jsonValue, ptr.Interface()); err != nil {
			return err
		}
		r.row[cName] = ptr.Elem().Interface()
	}
	return nil
}

func (r *rowImpl) TableName() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tSch.Name
}

func (r *rowImpl) TableSchema() *TableSchema {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tSch
}

// Get implements Row.Get.
// Get will panic on schema violation
func (r *rowImpl) Get(cName string) any {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.get(cName)
}

// get is the internal implementation of Get. (unlocked)
func (r *rowImpl) get(cName string) any {
	value, ok := r.row[cName]
	if !ok {
		if _, ok := r.tSch.Columns[cName]; !ok {
			// early panic
			panic(fmt.Errorf("schema violated: table %q doesn't have column %q", r.tSch.Name, cName))
		}
		r.row[cName] = r.tSch.Columns[cName].GetDefaultValue()
		value = r.row[cName]
	}
	return value
}

func (r *rowImpl) Set(cName string, value any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.set(cName, value)
}

// set is the internal implementation of Set. (unlocked)
func (r *rowImpl) set(cName string, value any) {
	if reflect.ValueOf(value).Kind() == reflect.Ptr {
		value = reflect.ValueOf(value).Elem().Interface()
	}
	cSch, ok := r.tSch.Columns[cName]
	if !ok {
		panic(fmt.Errorf("schema violated: table %q doesn't have column %q", r.tSch.Name, cName))
	}
	if err := cSch.ValidateValue(value); err != nil {
		panic(fmt.Errorf("schema violated: %v", err))
	}
	r.row[cName] = value
}

func (r *rowImpl) Update2(_diff Row) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	diff, ok := _diff.(*rowImpl)
	if !ok {
		return fmt.Errorf("diff bad Row implementation")
	}
	for cName, ucVal := range diff.row {
		// special case for set of 0/1
		cSch, ok := r.tSch.Columns[cName]
		if ok && *cSch.Type.Min == 0 && *cSch.Type.Max.(*int) == 1 {
			r.set(cName, ucVal)
			continue
		}
		// usual path
		cVal, ok := r.get(cName).(types.Updater2)
		if !ok {
			r.set(cName, ucVal)
			continue
		}
		if updatedVal, err := cVal.Update2(ucVal); err != nil {
			return err
		} else {
			r.set(cName, updatedVal)
		}
	}
	return nil
}

func (r *rowImpl) Match(where []types.Condition) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, cond := range where {
		cSch, ok := r.tSch.Columns[cond.GetColumn()]
		if !ok {
			return false
		}
		if err := cSch.ValidateCond(cond.GetOp(), cond.GetValue()); err != nil {
			return false
		}
		if !cond.Check(r.get(cond.GetColumn())) {
			return false
		}
	}
	return true
}
