package db

import (
	_ "embed"
	"encoding/json"
	"github.com/kazmanavt/ovsdb/monitor"
	"github.com/kazmanavt/ovsdb/schema"
	"github.com/kazmanavt/ovsdb/types"
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

//go:embed testdata/initialC.json
var initialC []byte

//go:embed testdata/updatesC1.json
var updatesC1 []byte

func Test_dbImpl_Update2(t *testing.T) {

	var dSch schema.DbSchema
	err := json.Unmarshal(ovsSchema, &dSch)
	require.NoError(t, err, "failed to unmarshal schema")

	t.Run("A", func(t *testing.T) {
		Db := NewDB(&dSch)

		var ini, upd1, upd2 monitor.RawTableSetUpdate2
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

		var ini, upd1, upd2, upd3, upd4 monitor.RawTableSetUpdate2
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

		//exp := `{
		//  "name": "nu1",
		//  "interfaces": [
		//    "set",
		//    [
		//      [
		//        "uuid",
		//        "8ef674bf-0ce0-4c24-bd2a-15dfc16fb678"
		//      ],
		//      [
		//        "uuid",
		//        "d401c3c5-e892-4b1c-853f-34bec442fc17"
		//      ]
		//    ]
		//  ],
		//  "other_config": [
		//    "map",
		//    [
		//      [
		//        "bond-detect-mode",
		//        "carrier"
		//      ],
		//      [
		//        "bond-rebalance-interval",
		//        "5000"
		//      ]
		//    ]
		//  ],
		//  "external_ids": [
		//    "map",
		//    [
		//      [
		//        "umb_network",
		//        "SysX sys0 1"
		//      ],
		//      [
		//        "umb_switch",
		//        "uSwitch0"
		//      ],
		//      [
		//        "umb_type",
		//        "uplink_bond"
		//      ],
		//      [
		//        "umb_version",
		//        "1.0.0"
		//      ]
		//    ]
		//  ],
		//  "bond_mode": "balance-slb"
		//}`
		//r := dSch.Tables["Port"].NewRow()
		//err = json.Unmarshal([]byte(exp), r)
		//require.NoError(t, err, "failed to unmarshal port exp")
		//exp2, err2 := json.Marshal(r)
		//require.NoError(t, err2, "failed to marshal port exp")
		//t.Log(Db.String())
		//act, err3 := json.Marshal(Db.TableRowS("Port", "8b8d8aaa-2ba7-411a-bbc5-557245943d5b"))
		//require.NoError(t, err3, "failed to marshal port")
		//require.JSONEqf(t, string(exp2), string(act), "port mismatch")

		err = Db.Update2(upd2)
		require.NoError(t, err, "failed to updatesB2")
		err = Db.Update2(upd3)
		require.NoError(t, err, "failed to updatesB3")
		err = Db.Update2(upd4)
		require.NoError(t, err, "failed to updatesB4")
	})

	t.Run("C", func(t *testing.T) {
		Db := NewDB(&dSch)

		var ini, upd1 monitor.RawTableSetUpdate2
		err = json.Unmarshal(initialC, &ini)
		require.NoError(t, err, "failed to unmarshal initialA")
		err = json.Unmarshal(updatesC1, &upd1)
		require.NoError(t, err, "failed to unmarshal updatesA1")

		uuid := "165f8f88-f073-41bc-8301-864050532dab"
		err = Db.Update2(ini)
		require.NoError(t, err, "failed to initialA update")

		eidsAny := Db.GetS("Bridge", uuid, "external_ids")
		require.IsType(t, (types.Map[string, string])(nil), eidsAny, "external_ids not a Map[str,str]")
		eids := eidsAny.(types.Map[string, string])
		nOptsStr, ok := eids["umb_net_opts"]
		require.True(t, ok, "missing umb_net_opts")
		var nOpts map[string]any
		err = json.Unmarshal([]byte(nOptsStr), &nOpts)
		require.NoError(t, err, "failed to unmarshal umb_net_opts")
		vlan, ok := nOpts["vlan"]
		require.True(t, ok, "missing vlan")
		require.Equal(t, 103.0, vlan, "vlan mismatch")

		err = Db.Update2(upd1)
		require.NoError(t, err, "failed to updatesA1")

		eidsAny = Db.GetS("Bridge", uuid, "external_ids")
		require.IsType(t, (types.Map[string, string])(nil), eidsAny, "external_ids not a Map[str,str]")
		eids = eidsAny.(types.Map[string, string])
		nOptsStr, ok = eids["umb_net_opts"]
		require.True(t, ok, "missing umb_net_opts")
		nOpts = make(map[string]any)
		err = json.Unmarshal([]byte(nOptsStr), &nOpts)
		require.NoError(t, err, "failed to unmarshal umb_net_opts")
		vlan, ok = nOpts["vlan"]
		require.True(t, ok, "missing vlan")
		require.Equal(t, 107.0, vlan, "vlan mismatch")
	})

}
