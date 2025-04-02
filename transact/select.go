package transact

import (
	"fmt"
	"github.com/kazmanavt/ovsdb/v2/schema"
	"github.com/kazmanavt/ovsdb/v2/types"
)

type selectOp struct {
	Op      string            `json:"op"`
	Table   string            `json:"table"`
	Where   []types.Condition `json:"where"`
	Columns []string          `json:"columns,omitempty"`
}

func (s *selectOp) Name() string {
	return s.Op
}

func (s *selectOp) Validate(dSch *schema.DbSchema) error {
	tName := s.Table
	tSch, ok := dSch.Tables[tName]
	if !ok {
		return fmt.Errorf("table %q not found", tName)
	}
	for _, cName := range s.Columns {
		if _, ok := tSch.Columns[cName]; !ok {
			return fmt.Errorf("column %q not found in table %q", cName, tName)
		}
	}
	for _, cond := range s.Where {
		cSch, ok := tSch.Columns[cond.GetColumn()]
		if !ok {
			return fmt.Errorf("column %q not found in table %q", cond.GetColumn(), tName)
		}
		if err := cSch.ValidateCond(cond.GetOp(), cond.GetValue()); err != nil {
			return fmt.Errorf("table %q: %w", tName, err)
		}
	}
	return nil
}

func (t *transaction) Select(tName string, where []types.Condition, columns []string) Transaction {
	t.txnSet = append(t.txnSet, &selectOp{
		Op:      "select",
		Table:   tName,
		Where:   where,
		Columns: columns,
	})
	tSch, ok := t.sch.Tables[tName]
	if ok {
		t.resp = append(t.resp, &Result{Rows: tSch.NewRows()})
	} else {
		t.resp = append(t.resp, &Result{})
	}
	return t
}
