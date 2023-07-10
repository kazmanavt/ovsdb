package schema

import (
	_ "embed"
	"encoding/json"
	"github.com/kazmanavt/ovsdb/types"
	"github.com/stretchr/testify/require"
	"testing"
)

//go:embed testdata/Open_vSwitch.json
var ovsSchemaC []byte

func TestColumnSchema_ValidateValue(t *testing.T) {
	var sch DbSchema
	err := json.Unmarshal(ovsSchema, &sch)
	require.NoError(t, err, "failed to unmarshal schema")

	t.Run("Validate bond_mode on Port table", func(t *testing.T) {
		cSch := sch.Tables["Port"].Columns["bond_mode"]
		require.NoError(t, err, "failed to unmarshal column")
		err = cSch.ValidateValue(types.Set[string]{"balance-slb"})
		require.NoError(t, err, "failed to validate value")
		err = cSch.ValidateValue([]string{"active-backup"})
		require.Error(t, err, "should be Set value")
		err = cSch.ValidateValue("balance-slb1")
		require.Error(t, err, "failed to validate value")
	})

}
