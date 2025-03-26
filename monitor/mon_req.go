package monitor

// MonReq is a request for monitoring a table.
// It denotes the columns to monitor and the types of updates to monitor.
type MonReq struct {
	Columns []string `json:"columns,omitempty"`
	Select  *Select  `json:"select,omitempty"`
}

func (mr *MonReq) GetColumns() []string {
	return mr.Columns
}
func (mr *MonReq) GetSelect() *Select {
	return mr.Select
}
func (mr *MonReq) withoutInitial() MonReq {
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
