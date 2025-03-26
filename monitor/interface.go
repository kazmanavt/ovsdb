package monitor

import "encoding/json"

type GenericMonReqSet interface {
	json.Marshaler
	WithoutInitial() GenericMonReqSet
	HasUpdates() bool
	Validate() error
}

type MonReqSet interface {
	GenericMonReqSet
	Add(table string, req ...MonReq) MonReqSet
}

type MonCondReqSet interface {
	GenericMonReqSet
	Add(table string, req ...MonCondReq) MonCondReqSet
}
