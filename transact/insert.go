package transact

import (
	"fmt"
	"github.com/kazmanavt/ovsdb/v2/schema"
)

type insertOp struct {
	Op    string     `json:"op"`
	Table string     `json:"table"`
	Row   schema.Row `json:"row"`
	Uuid  *string    `json:"uuid-name,omitempty"`
}

func (i *insertOp) Name() string {
	return i.Op
}

func (i *insertOp) Validate(_ *schema.DbSchema) error {
	if i.Row == nil {
		return fmt.Errorf("nil row to insert operation")
	}
	return nil
}

func (t *transaction) Insert(row schema.Row, uuid ...string) Transaction {
	uu := new(string)
	if len(uuid) > 0 {
		uu = &uuid[0]
	}

	t.txnSet = append(t.txnSet, &insertOp{
		Op:    "insert",
		Table: row.TableName(),
		Row:   row,
		Uuid:  uu,
	})
	t.resp = append(t.resp, &Result{})
	return t
}
