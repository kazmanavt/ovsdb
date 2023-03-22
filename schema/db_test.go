package schema

import (
	_ "embed"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

//go:embed testdata/Open_vSwitch.json
var ovsSchema []byte

func TestDbSchema(t *testing.T) {
	t.Run("unmarshal", func(t *testing.T) {
		var db DbSchema
		err := json.Unmarshal(ovsSchema, &db)
		for tName, table := range db.Tables {
			assert.NotEmptyf(t, table.Name, "table['%s'] name is empty", tName)
			for cName, column := range table.Columns {
				require.NotEmptyf(t, column.Name, "table['%s'] column['%s'] name is empty", tName, cName)
				if column.Type.Key.Type == "map" {
					t.Logf("%s::%s -> map[%s,%s]", tName, cName, column.Type.Key.Type, column.Type.Value.Type)
				}
			}
		}
		assert.NoError(t, err)
		assert.Equal(t, "string", db.Tables["Bridge"].Columns["name"].Type.Key.Type)
		assert.Equal(t, "uuid", db.Tables["Bridge"].Columns["ports"].Type.Key.Type)
		assert.Equal(t, "string", db.Tables["Bridge"].Columns["other_config"].Type.Key.Type)
		assert.Equal(t, "string", db.Tables["Bridge"].Columns["other_config"].Type.Value.Type)

	})
}
