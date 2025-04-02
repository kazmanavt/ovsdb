package monitor

import (
	"encoding/json"
	"fmt"
	"github.com/kazmanavt/ovsdb/v2/schema"
)

func NewMonReqSet(ds *schema.DbSchema) MonReqSet {
	return &monReqSet{
		sch: ds,
		//tColNames: make(map[string]map[string]int),
		reqs: make(map[string][]MonReq),
	}
}

// monReqSet is used to store the monitor requests
// It is used to pass the requests to the Monitor method of the Client
type monReqSet struct {
	sch                    *schema.DbSchema // DbSchema is used to validate the requests in Validate function
	reqs                   map[string][]MonReq
	hasInitial, hasUpdates bool
}

func (mm *monReqSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(mm.reqs)
}

func (mm *monReqSet) Add(tName string, reqs ...MonReq) MonReqSet {
	i, u := setInitialAndUpdatesFlags(reqs)
	mm.hasInitial = mm.hasInitial || i
	mm.hasUpdates = mm.hasUpdates || u
	mm.reqs[tName] = append(mm.reqs[tName], reqs...)
	return mm
}

func (mm *monReqSet) WithoutInitial() GenericMonReqSet {
	if mm.hasInitial {
		clone := NewMonReqSet(mm.sch).(*monReqSet)
		clone.hasUpdates = mm.hasUpdates
		clone.hasInitial = false
		for tName, reqs := range mm.reqs {
			for _, req := range reqs {
				clone.Add(tName, req.withoutInitial())
			}
		}
		return clone
	}
	return mm
}

func (mm *monReqSet) Validate() error {
	if mm == nil {
		return fmt.Errorf("nil monitor requests")
	}
	for tName, tReqs := range mm.reqs {
		tSch, ok := mm.sch.Tables[tName]
		if !ok {
			return fmt.Errorf("non-existent table %q", tName)
		}
		if err := validateRequestColumns(tSch, tReqs); err != nil {
			return err
		}
		if err := validateRequestSelect(tReqs); err != nil {
			return err
		}
	}
	return nil
}

func (mm *monReqSet) HasInitial() bool {
	return mm.hasInitial
}

func (mm *monReqSet) HasUpdates() bool {
	return mm.hasUpdates
}
