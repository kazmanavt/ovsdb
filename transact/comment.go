package transact

import "github.com/kazmanavt/ovsdb/v2/schema"

type commentOp struct {
	Op      string `json:"op"`
	Comment string `json:"comment"`
}

func (c *commentOp) Name() string {
	return c.Op
}

func (c *commentOp) Validate(_ *schema.DbSchema) error {
	return nil
}

func (t *transaction) Comment(comment string) Transaction {
	t.txnSet = append(t.txnSet, &commentOp{
		Op:      "comment",
		Comment: comment,
	})
	t.resp = append(t.resp, &Result{})
	return t
}
