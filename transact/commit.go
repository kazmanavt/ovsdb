package transact

import "github.com/kazmanavt/ovsdb/v2/schema"

type commitOp struct {
	Op      string `json:"op"`
	Durable bool   `json:"durable"`
}

func (c *commitOp) Name() string {
	return c.Op
}

func (c *commitOp) Validate(_ *schema.DbSchema) error {
	return nil
}

func (t *transaction) Commit(durable bool) Transaction {
	t.txnSet = append(t.txnSet, &commitOp{
		Op:      "commit",
		Durable: durable,
	})
	t.resp = append(t.resp, &Result{})
	return t
}
