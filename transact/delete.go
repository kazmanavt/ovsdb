package transact

import (
	"fmt"
	"github.com/kazmanavt/ovsdb/schema"
	"github.com/kazmanavt/ovsdb/types"
)

type deleteOp struct {
	Op    string            `json:"op"`
	Table string            `json:"table"`
	Where []types.Condition `json:"where"`
}

func (d *deleteOp) Name() string {
	return d.Op
}

func (d *deleteOp) Validate(dSch *schema.DbSchema) error {
	tName := d.Table
	tSch, ok := dSch.Tables[tName]
	if !ok {
		return fmt.Errorf("table %q not found", tName)
	}
	for _, cond := range d.Where {
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

func (t *transaction) Delete(tName string, where []types.Condition) Transaction {
	t.txnSet = append(t.txnSet, &deleteOp{
		Op:    "delete",
		Table: tName,
		Where: where,
	})
	t.resp = append(t.resp, &Result{})
	return t
}
