package monitor

import (
	"encoding/json"
	"fmt"
	"github.com/kazmanavt/ovsdb/schema"
)

type GenericMonReqs interface {
	json.Marshaler
	json.Unmarshaler
	HasInitial() bool
	HasUpdates() bool
	Validate() error
}

type MonReqs interface {
	GenericMonReqs
	Add(table string, req ...MonReq) MonReqs
}

type MonCondReqs interface {
	GenericMonReqs
	Add(table string, req ...MonCondReq) MonCondReqs
}

type req interface {
	GetColumns() []string
	GetSelect() *Select
}

type Select struct {
	Initial bool `json:"initial"`
	Insert  bool `json:"insert"`
	Delete  bool `json:"delete"`
	Modify  bool `json:"modify"`
}

func setInitialAndUpdatesFlags[T MonReq | MonCondReq](tReqs []T) (hasInitial, hasUpdates bool) {
	for _, _r := range tReqs {
		r := any(&_r).(req)
		sel := r.GetSelect()
		if sel == nil || sel.Initial {
			hasInitial = true
		}
		if sel == nil || sel.Insert || sel.Delete || sel.Modify {
			hasUpdates = true
		}
	}
	return
}

func validateRequestSelect[T MonReq | MonCondReq](tReqs []T) error {
	for _, _r := range tReqs {
		r := any(&_r).(req)
		sel := r.GetSelect()
		if sel != nil && !(sel.Initial || sel.Insert || sel.Delete || sel.Modify) {
			return fmt.Errorf("no select options")
		}
	}
	return nil
}

func validateRequestedColumns[T MonReq | MonCondReq](tSch *schema.TableSchema, tReqs []T) error {
	tName := tSch.Name
	numReqs := len(tReqs)
	colNames := make(map[string]int)
	for _, _r := range tReqs {

		r := any(&_r).(req)
		if numReqs > 1 && r.GetColumns() == nil {
			return fmt.Errorf("non-single all-column request for table %q", tName)
		}

		for _, cName := range r.GetColumns() {
			if _, ok := tSch.Columns[cName]; !ok {
				return fmt.Errorf("no column %q in table %q", cName, tName)
			}
			colNames[cName] += 1
			if colNames[cName] > 1 {
				return fmt.Errorf("duplicate column %q in table %q", cName, tName)
			}
		}
	}
	return nil
}
