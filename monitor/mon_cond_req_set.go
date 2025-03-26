package monitor

import (
	"encoding/json"
	"fmt"
	"github.com/kazmanavt/ovsdb/schema"
	"github.com/kazmanavt/ovsdb/types"
)

func NewMonCondReqSet(ds *schema.DbSchema) MonCondReqSet {
	return &monCondReqSet{
		monReqSet: monReqSet{
			sch: ds,
		},
		reqs: make(map[string][]MonCondReq),
	}
}

// monReqSet is used to store the monitor requests
// It is used to pass the requests to the Monitor method of the Client
type monCondReqSet struct {
	monReqSet
	reqs map[string][]MonCondReq
}

func (mm *monCondReqSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(mm.reqs)
}

func (mm *monCondReqSet) Add(tName string, reqs ...MonCondReq) MonCondReqSet {
	i, u := setInitialAndUpdatesFlags(reqs)
	mm.hasInitial = mm.hasInitial || i
	mm.hasUpdates = mm.hasUpdates || u
	for _, r := range reqs {
		if r.Where == nil {
			r.Where = []types.Condition{}
		}
	}
	mm.reqs[tName] = append(mm.reqs[tName], reqs...)
	return mm
}

func (mm *monCondReqSet) WithoutInitial() GenericMonReqSet {
	if mm.hasInitial {
		clone := NewMonCondReqSet(mm.sch).(*monCondReqSet)
		clone.hasUpdates = mm.hasUpdates
		clone.hasInitial = false
		for tName, tReqs := range mm.reqs {
			for _, r := range tReqs {
				clone.Add(tName, r.withoutInitial())
			}
		}
		return clone
	}
	return mm
}

func (mm *monCondReqSet) Validate() error {
	if mm == nil {
		return fmt.Errorf("nil monitor cond requests")
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
