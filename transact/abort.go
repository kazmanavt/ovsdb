package transact

import "github.com/kazmanavt/ovsdb/schema"

type abortOp struct {
	Op string `json:"op"`
}

func (a *abortOp) Name() string {
	return a.Op
}

func (a *abortOp) Validate(_ *schema.DbSchema) error {
	return nil
}

func (t *transaction) Abort() Transaction {
	t.txnSet = append(t.txnSet, &abortOp{Op: "abort"})
	t.resp = append(t.resp, &Result{})
	return t
}
