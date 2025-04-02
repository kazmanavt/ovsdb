package transact

import (
	"fmt"
	"github.com/kazmanavt/ovsdb/v2/schema"
	"github.com/kazmanavt/ovsdb/v2/types"
	"strings"
)

type waitOp struct {
	Op      string            `json:"op"`
	Timeout int               `json:"timeout,omitempty"`
	Table   string            `json:"table"`
	Where   []types.Condition `json:"where"`
	Columns []string          `json:"columns"`
	Until   string            `json:"until"`
	Rows    []schema.Row      `json:"rows"`
}

func (w *waitOp) Name() string {
	return w.Op
}

func (w *waitOp) Validate(dSch *schema.DbSchema) error {
	tName := w.Table
	tSch, ok := dSch.Tables[tName]
	if !ok {
		return fmt.Errorf("table %q not found", tName)
	}

	if !strings.Contains("==!=", w.Until) {
		return fmt.Errorf("invalid until %q", w.Until)
	}

	for _, cName := range w.Columns {
		if _, ok := tSch.Columns[cName]; !ok {
			return fmt.Errorf("column %q not found in table %q", cName, tName)
		}
	}

	for _, cond := range w.Where {
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

func (t *transaction) Wait(tName string, where []types.Condition, columns []string, until string, rows []schema.Row, timeout int) Transaction {
	t.txnSet = append(t.txnSet, &waitOp{
		Op:      "wait",
		Table:   tName,
		Where:   where,
		Columns: columns,
		Until:   until,
		Rows:    rows,
		Timeout: timeout,
	})

	t.resp = append(t.resp, &Result{})
	return t
}
