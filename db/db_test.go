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

//go:embed testdata/initialA.json
var initialA []byte

//go:embed testdata/updatesA1.json
var updatesA1 []byte

//go:embed testdata/updatesA2.json
var updatesA2 []byte

//go:embed testdata/initialB.json
var initialB []byte

//go:embed testdata/updatesB1.json
var updatesB1 []byte

//go:embed testdata/updatesB2.json
var updatesB2 []byte

//go:embed testdata/updatesB3.json
var updatesB3 []byte

//go:embed testdata/updatesB4.json
var updatesB4 []byte
func Test_dbImpl_Update2(t *testing.T) {

	var dSch schema.DbSchema
	err := json.Unmarshal(ovsSchema, &dSch)
	require.NoError(t, err, "failed to unmarshal schema")

	t.Run("A", func(t *testing.T) {
		Db := NewDB(&dSch)

		var ini, upd1, upd2 monitor.Updates2
		err = json.Unmarshal(initialA, &ini)
		require.NoError(t, err, "failed to unmarshal initialA")
		err = json.Unmarshal(updatesA1, &upd1)
		require.NoError(t, err, "failed to unmarshal updatesA1")
		err = json.Unmarshal(updatesA2, &upd2)
		require.NoError(t, err, "failed to unmarshal updatesA2")

		err = Db.Update2(ini)
		require.NoError(t, err, "failed to initialA update")
		err = Db.Update2(upd1)
		require.NoError(t, err, "failed to updatesA1")
		err = Db.Update2(upd2)
		require.NoError(t, err, "failed to updatesA2")
	})

	t.Run("B", func(t *testing.T) {
		Db := NewDB(&dSch)

		var ini, upd1, upd2, upd3, upd4 monitor.Updates2
		err = json.Unmarshal(initialB, &ini)
		require.NoError(t, err, "failed to unmarshal initialB")
		err = json.Unmarshal(updatesB1, &upd1)
		require.NoError(t, err, "failed to unmarshal updatesB1")
		err = json.Unmarshal(updatesB2, &upd2)
		require.NoError(t, err, "failed to unmarshal updatesB2")
		err = json.Unmarshal(updatesB3, &upd3)
		require.NoError(t, err, "failed to unmarshal updatesB3")
		err = json.Unmarshal(updatesB4, &upd4)
		require.NoError(t, err, "failed to unmarshal updatesB4")

		err = Db.Update2(ini)
		require.NoError(t, err, "failed to initialB update")
		err = Db.Update2(upd1)
		require.NoError(t, err, "failed to updatesB1")
		err = Db.Update2(upd2)
		require.NoError(t, err, "failed to updatesB2")
		err = Db.Update2(upd3)
		require.NoError(t, err, "failed to updatesB3")
		err = Db.Update2(upd4)
		require.NoError(t, err, "failed to updatesB4")
	})
}
