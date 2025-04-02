package transact

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kazmanavt/ovsdb/schema"
	"github.com/kazmanavt/ovsdb/types"
)

func NewTransaction(sch *schema.DbSchema) Transaction {
	return &transaction{
		sch:    sch,
		txnSet: make([]operation, 0),
		resp:   make([]*Result, 0),
	}
}

type Transaction interface {
	Insert(row schema.Row, uuid ...string) Transaction
	Select(tName string, where []types.Condition, columns []string) Transaction
	Update(where []types.Condition, row schema.Row) Transaction
	Mutate(tName string, where []types.Condition, mutt []types.Mutation) Transaction
	Delete(tName string, where []types.Condition) Transaction
	Wait(tName string, where []types.Condition, columns []string, until string, rows []schema.Row, timeout int) Transaction
	Commit(durable bool) Transaction
	Abort() Transaction
	Comment(comment string) Transaction
	Assert() Transaction
	Validate() error
	Operations() []operation
	Len() int
	DecodeResult(result json.RawMessage) error
	Result(idx int) *Result
	Error() error
	Clone() Transaction
}
type transaction struct {
	sch    *schema.DbSchema
	txnSet []operation
	resp   []*Result
}

func (t *transaction) Clone() Transaction {
	newT := *t
	return &newT
}

func (t *transaction) Validate() error {
	if t == nil {
		return fmt.Errorf("nil transaction")
	}
	for _, op := range t.txnSet {
		if err := op.Validate(t.sch); err != nil {
			return err
		}
	}
	return nil
}

func (t *transaction) Operations() []operation {
	return t.txnSet
}

func (t *transaction) Len() int {
	return len(t.txnSet)
}

func (t *transaction) DecodeResult(result json.RawMessage) error {
	return json.Unmarshal(result, &t.resp)
}

func (t *transaction) Result(idx int) *Result {
	return t.resp[idx]
}

func (t *transaction) Error() error {
	for i, r := range t.resp {
		if r.Error != nil {
			var errStr string
			if i >= len(t.txnSet) {
				errStr = fmt.Sprintf("general transaction error: %v", r.Error)
			} else {
				errStr = fmt.Sprintf("operation #%d(%s): %v", i, t.txnSet[i].Name(), r.Error)
			}
			if r.Details != nil {
				errStr += fmt.Sprintf(" (%v)", r.Details)
			}
			return errors.New(errStr)
		}
	}
	return nil
}
