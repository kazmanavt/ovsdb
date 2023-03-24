package monitor

import (
	"encoding/json"
	"fmt"
	"github.com/kazmanavt/ovsdb/schema"
	"github.com/kazmanavt/ovsdb/types"
)

func NewMonCondReqs(ds *schema.DbSchema) MonCondReqs {
	return &monCondReqs{
		monReqs: monReqs{
			sch: ds,
		},
		reqs: make(map[string][]MonCondReq),
	}
}

// monReqs is used to store the monitor requests
// It is used to pass the requests to the Monitor method of the Client
type monCondReqs struct {
	monReqs
	reqs map[string][]MonCondReq
}

func (mcrs *monCondReqs) MarshalJSON() ([]byte, error) {
	return json.Marshal(mcrs.reqs)
}

func (mcrs *monCondReqs) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &mcrs.reqs)
}

func (mcrs *monCondReqs) Add(tName string, reqs ...MonCondReq) MonCondReqs {
	i, u := setInitialAndUpdatesFlags(reqs)
	mcrs.hasInitial = mcrs.hasInitial || i
	mcrs.hasUpdates = mcrs.hasUpdates || u
	for _, r := range reqs {
		if r.Where == nil {
			r.Where = []types.Condition{}
		}
	}
	mcrs.reqs[tName] = append(mcrs.reqs[tName], reqs...)
	return mcrs
}

func (mcrs *monCondReqs) Validate() error {
	if mcrs == nil {
		return fmt.Errorf("nil monitor cond requests")
	}
	for tName, tReqs := range mcrs.reqs {
		tSch, ok := mcrs.sch.Tables[tName]
		if !ok {
			return fmt.Errorf("non-existent table %q", tName)
		}
		if err := validateRequestedColumns(tSch, tReqs); err != nil {
			return err
		}
		if err := validateRequestSelect(tReqs); err != nil {
			return err
		}
		for i, r := range tReqs {
			for _, cond := range r.Where {
				column := cond.GetColumn()
				cSch, ok := tSch.Columns[column]
				if !ok {
					return fmt.Errorf("req #%d: column %q not in table %q", i, column, tName)
				}
				if err := cSch.ValidateCond(cond.GetOp(), cond.GetValue()); err != nil {
					return fmt.Errorf("req #%d on table %q: %w", i, tName, err)
				}
			}
		}
	}
	return nil
}

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
