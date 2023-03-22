package schema

import (
	"encoding/json"
)

// DbSchema is the schema of a database (RFC 7047).
// "name": <id>                            required
// "version": <version>                    required
// "cksum": <string>                       optional
// "tables": {<id>: <table-schema>, ...}   required
type DbSchema struct {
	Name    string                  `json:"name"`            // Name of the database
	Version string                  `json:"version"`         // version of the database schema
	Cksum   string                  `json:"cksum,omitempty"` // checksum of the database schema
	Tables  map[string]*TableSchema `json:"tables"`          // tables in the database
}

func (ds *DbSchema) UnmarshalJSON(data []byte) error {
	type DS DbSchema
	if err := json.Unmarshal(data, (*DS)(ds)); err != nil {
		return err
	}
	for name, tbl := range ds.Tables {
		tbl.Name = name
		addUUID(tbl)
		addVersion(tbl)
		for name, col := range tbl.Columns {
			col.Name = name
		}
	}
	return nil
}

var one = 1

func addUUID(tbl *TableSchema) {
	tbl.Columns["_uuid"] = &ColumnSchema{
		Name: "_uuid",
		Type: ColumnType{
			kind: "uuid",
			Key: BaseType{
				Type: "uuid",
			},
			Min: &one,
			Max: &one,
		},
		Ephemeral: false,
		Mutable:   false,
	}
}
func addVersion(tbl *TableSchema) {
	tbl.Columns["_version"] = &ColumnSchema{
		Name: "_version",
		Type: ColumnType{
			kind: "uuid",
			Key: BaseType{
				Type: "uuid",
			},
			Min: &one,
			Max: &one,
		},
		Ephemeral: false,
		Mutable:   true,
	}
}
