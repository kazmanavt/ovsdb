package monitor

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Update3 struct {
	monName   string
	lastTsnID string
	update2   TableSetUpdate2
}

func (u *Update3) UnmarshalJSON(data []byte) error {
	var params []json.RawMessage
	if err := json.Unmarshal(data, &params); err != nil {
		return err
	}
	if len(params) != 3 {
		return errors.New("wrong number of param in Update3 notification")
	}
	if err := json.Unmarshal(params[0], &u.monName); err != nil {
		return fmt.Errorf("unmarshal monName: %w", err)
	}
	if err := json.Unmarshal(params[1], &u.lastTsnID); err != nil {
		return fmt.Errorf("unmarshal lastTsnID: %w", err)
	}
	var upd2 TableSetUpdate2 = make(map[string]TableUpdate2)
	if err := json.Unmarshal(params[1], &upd2); err != nil {
		return fmt.Errorf("unmarshal update2: %w", err)
	}
	u.update2 = upd2
	return nil
}
