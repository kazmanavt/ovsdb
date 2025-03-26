package monitor

import (
	"encoding/json"
	"errors"
	"fmt"
)

type RowUpdate struct {
	New json.RawMessage `json:"new,omitempty"`
	Old json.RawMessage `json:"old,omitempty"`
}

type TableUpdate map[string]RowUpdate

type TableSetUpdate map[string]TableUpdate

type Update struct {
	monName string
	update  TableSetUpdate
}

func (u *Update) UnmarshalJSON(data []byte) error {
	var params []json.RawMessage
	if err := json.Unmarshal(data, &params); err != nil {
		return err
	}
	if len(params) != 2 {
		return errors.New("wrong number of param in Update notification")
	}
	if err := json.Unmarshal(params[0], &u.monName); err != nil {
		return fmt.Errorf("unmarshal monName: %w", err)
	}
	var upd TableSetUpdate = make(map[string]TableUpdate)
	if err := json.Unmarshal(params[1], &upd); err != nil {
		return fmt.Errorf("unmarshal update: %w", err)
	}
	u.update = upd
	return nil
}
