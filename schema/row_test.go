package schema

import (
	_ "embed"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

//go:embed testdata/bridge_row.json
var bridgeRow []byte

//go:embed testdata/bridge_row_array.json
var bridgeRowSet []byte

func TestRow_MarshalJSON(t *testing.T) {
}

func TestRow_TableName(t *testing.T) {

}

func TestRow_UnmarshalJSON(t *testing.T) {
	var dbs DbSchema
	_ = json.Unmarshal(ovsSchema, &dbs)
	r := dbs.Tables["Bridge"].NewRow()
	err := json.Unmarshal(bridgeRow, &r)
	require.NoError(t, err, "should be happy unmarshaled")

	rs := dbs.Tables["Bridge"].NewRows()
	err = json.Unmarshal(bridgeRowSet, &rs)
	require.NoError(t, err, "should be happy unmarshaled")
}
