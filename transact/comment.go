package transact

import "github.com/kazmanavt/ovsdb/schema"

type commentOp struct {
	Op      string `json:"op"`
	Comment string `json:"comment"`
}

func (c *commentOp) Name() string {
	return c.Op
}

func (c *commentOp) Validate(dSch *schema.DbSchema) error {
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
