package monitor

import (
	"github.com/kazmanavt/ovsdb/v2/types"
)

// MonCondReq is a request for monitoring a table.
// It denotes the columns to monitor and the types of updates to monitor.
// additionally it specifies condition to filter the rows to monitor
type MonCondReq struct {
	Columns []string          `json:"columns,omitempty"`
	Where   []types.Condition `json:"where,omitempty"`
	Select  *Select           `json:"select,omitempty"`
}

func (mr *MonCondReq) GetColumns() []string {
	return mr.Columns
}

func (mr *MonCondReq) GetSelect() *Select {
	return mr.Select
}

func (mr *MonCondReq) clone() MonCondReq {
	clone := *mr
	sel := *mr.Select
	clone.Select = &sel
	return clone
}

func (mr *MonCondReq) withoutInitial() MonCondReq {
	clone := *mr
	if mr.Select == nil {
		clone.Select = &Select{
			Initial: false,
			Insert:  true,
			Delete:  true,
			Modify:  true,
		}
		return clone
	}
	sel := *mr.Select
	sel.Initial = false
	clone.Select = &sel
	return clone
}
