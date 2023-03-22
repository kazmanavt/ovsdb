package transact

import (
	"fmt"
	"github.com/kazmanavt/ovsdb/schema"
	"github.com/kazmanavt/ovsdb/types"
)

type updateOp struct {
	Op    string            `json:"op"`
	Table string            `json:"table"`
	Where []types.Condition `json:"where"`
	Row   schema.Row        `json:"row"`
}

func (u *updateOp) Name() string {
	return u.Op
}

func (u *updateOp) Validate(_ *schema.DbSchema) error {
	if u.Row == nil {
		return fmt.Errorf("nil row to insert operation")
	}

	for _, cond := range u.Where {
		cName := cond.GetColumn()
		cSch, ok := u.Row.TableSchema().Columns[cName]
		if !ok {
			return fmt.Errorf("column %q not found in table %q", cName, u.Table)
		}
		if err := cSch.ValidateCond(cond.GetOp(), cond.GetValue()); err != nil {
			return fmt.Errorf("table %q: %w", u.Table, err)
		}
	}
	return nil
}

func (t *transaction) Update(where []types.Condition, row schema.Row) Transaction {
	t.txnSet = append(t.txnSet, &updateOp{
		Op:    "update",
		Table: row.TableName(),
		Where: where,
		Row:   row,
	})
	t.resp = append(t.resp, &Result{})
	return t
}
