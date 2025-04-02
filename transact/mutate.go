package transact

import (
	"fmt"
	"github.com/kazmanavt/ovsdb/v2/schema"
	"github.com/kazmanavt/ovsdb/v2/types"
)

type mutateOp struct {
	Op        string            `json:"op"`
	Table     string            `json:"table"`
	Where     []types.Condition `json:"where"`
	Mutations []types.Mutation  `json:"mutations"`
}

func (m *mutateOp) Name() string {
	return m.Op
}

func (m *mutateOp) Validate(dSch *schema.DbSchema) error {
	tName := m.Table
	tSch, ok := dSch.Tables[tName]
	if !ok {
		return fmt.Errorf("table %q not found", tName)
	}
	for _, cond := range m.Where {
		cSch, ok := tSch.Columns[cond.GetColumn()]
		if !ok {
			return fmt.Errorf("column %q not found in table %q", cond.GetColumn(), tName)
		}
		if err := cSch.ValidateCond(cond.GetOp(), cond.GetValue()); err != nil {
			return fmt.Errorf("table %q: %w", tName, err)
		}
	}
	for _, m := range m.Mutations {
		cSch, ok := tSch.Columns[m.GetColumn()]
		if !ok {
			return fmt.Errorf("column %q not found in table %q", m.GetColumn(), tName)
		}
		if err := cSch.ValidateMutation(m.GetOp(), m.GetValue()); err != nil {
			return fmt.Errorf("table %q: %w", tName, err)
		}
	}
	return nil
}

func (t *transaction) Mutate(tName string, where []types.Condition, mutt []types.Mutation) Transaction {
	t.txnSet = append(t.txnSet, &mutateOp{
		Op:        "mutate",
		Table:     tName,
		Where:     where,
		Mutations: mutt,
	})
	t.resp = append(t.resp, &Result{})
	return t
}
