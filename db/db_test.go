package db

import (
	_ "embed"
	"encoding/json"
	"github.com/kazmanavt/ovsdb/monitor"
	"github.com/kazmanavt/ovsdb/schema"
	"github.com/stretchr/testify/require"
	"testing"
)

//go:embed testdata/Open_vSwitch.json
var ovsSchema []byte

//go:embed testdata/initial.json
var initial []byte

//go:embed testdata/updates1.json
var updates1 []byte

//go:embed testdata/updates2.json
var updates2 []byte
func Test_dbImpl_Update2(t *testing.T) {

	var dSch schema.DbSchema
	err := json.Unmarshal(ovsSchema, &dSch)
	require.NoError(t, err, "failed to unmarshal schema")

	Db := NewDB(&dSch)

	var ini, upd1, upd2 monitor.Updates2
	err = json.Unmarshal(initial, &ini)
	require.NoError(t, err, "failed to unmarshal initial")
	err = json.Unmarshal(updates1, &upd1)
	require.NoError(t, err, "failed to unmarshal updates1")
	err = json.Unmarshal(updates2, &upd2)
	require.NoError(t, err, "failed to unmarshal updates2")

	err = Db.Update2(ini)
	require.NoError(t, err, "failed to initial update")
	err = Db.Update2(upd1)
	require.NoError(t, err, "failed to update1")
	err = Db.Update2(upd2)
	require.NoError(t, err, "failed to update2")

}
