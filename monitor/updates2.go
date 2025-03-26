package monitor

import (
	"encoding/json"
	"errors"
	"fmt"
)

type RowUpdate2 struct {
	Initial json.RawMessage `json:"initial,omitempty"`
	Insert  json.RawMessage `json:"insert,omitempty"`
	Delete  json.RawMessage `json:"delete,omitempty"`
	Modify  json.RawMessage `json:"modify,omitempty"`
}

type TableUpdate2 map[string]RowUpdate2

type TableSetUpdate2 map[string]TableUpdate2

type Update2 struct {
	monName string
	update2 TableSetUpdate2
}

func (u *Update2) UnmarshalJSON(data []byte) error {
	var params []json.RawMessage
	if err := json.Unmarshal(data, &params); err != nil {
		return err
	}
	if len(params) != 2 {
		return errors.New("wrong number of param in Update2 notification")
	}
	if err := json.Unmarshal(params[0], &u.monName); err != nil {
		return fmt.Errorf("unmarshal monName: %w", err)
	}
	var upd2 TableSetUpdate2 = make(map[string]TableUpdate2)
	if err := json.Unmarshal(params[1], &upd2); err != nil {
		return fmt.Errorf("unmarshal update2: %w", err)
	}
	u.update2 = upd2
	return nil
}
