package monitor

import (
	"encoding/json"
	"fmt"
	"github.com/kazmanavt/ovsdb/schema"
)

func NewMonReqs(ds *schema.DbSchema) MonReqs {
	return &monReqs{
		sch: ds,
		//tColNames: make(map[string]map[string]int),
		reqs: make(map[string][]MonReq),
	}
}

// monReqs is used to store the monitor requests
// It is used to pass the requests to the Monitor method of the Client
type monReqs struct {
	sch                    *schema.DbSchema // DbSchema is used to validate the requests in Validate function
	reqs                   map[string][]MonReq
	hasInitial, hasUpdates bool
}

func (mrs *monReqs) MarshalJSON() ([]byte, error) {
	return json.Marshal(mrs.reqs)
}

func (mrs *monReqs) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &mrs.reqs)
}

func (mrs *monReqs) Add(tName string, reqs ...MonReq) MonReqs {
	i, u := setInitialAndUpdatesFlags(reqs)
	mrs.hasInitial = mrs.hasInitial || i
	mrs.hasUpdates = mrs.hasUpdates || u
	mrs.reqs[tName] = append(mrs.reqs[tName], reqs...)
	return mrs
}

func (mrs *monReqs) Validate() error {
	if mrs == nil {
		return fmt.Errorf("nil monitor requests")
	}
	for tName, tReqs := range mrs.reqs {
		tSch, ok := mrs.sch.Tables[tName]
		if !ok {
			return fmt.Errorf("non-existent table %q", tName)
		}
		if err := validateRequestedColumns(tSch, tReqs); err != nil {
			return err
		}
		if err := validateRequestSelect(tReqs); err != nil {
			return err
		}
	}
	return nil
}

func (mrs *monReqs) HasInitial() bool {
	return mrs.hasInitial
}

func (mrs *monReqs) HasUpdates() bool {
	return mrs.hasUpdates
}

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
