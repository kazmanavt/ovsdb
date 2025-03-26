package monitor

import (
	_ "embed"
	"encoding/json"
	"github.com/kazmanavt/ovsdb/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

//go:embed testdata/Open_vSwitch.json
var ovsSchema []byte

func TestDbSchema_NewMonReqs(t *testing.T) {
	var sch schema.DbSchema
	err := json.Unmarshal(ovsSchema, &sch)
	require.NoError(t, err, "fail to load schema")
	reqs := NewMonReqSet(&sch)
	assert.NotNil(t, reqs, "request should not be nil")
	assert.IsType(t, &monReqSet{}, reqs, "incorrect type")
	assert.Equal(t, &monReqSet{sch: &sch, reqs: make(map[string][]MonReq)}, reqs, "incorrect request")
}

func Test_monReqs_Add(t *testing.T) {
	var sch schema.DbSchema
	err := json.Unmarshal(ovsSchema, &sch)
	require.NoError(t, err, "fail to load schema")
	reqs := NewMonReqSet(&sch)
	t.Run("fail to add request to non-existing table", func(t *testing.T) {
		reqs.Add("NonExistingTable", MonReq{
			Columns: []string{"name"},
			Select: &Select{
				Initial: true,
			},
		})
		err = reqs.Validate()
		assert.Error(t, err)
	})
	reqs = NewMonReqSet(&sch)
	t.Run("fail to add request with non-existing column", func(t *testing.T) {
		reqs.Add("Bridge", MonReq{
			Columns: []string{"name", "non-existing"},
			Select: &Select{
				Initial: true,
			},
		})
		err = reqs.Validate()
		assert.Error(t, err)
	})
	reqs = NewMonReqSet(&sch)
	t.Run("fail to add request with duplicate column", func(t *testing.T) {
		reqs.Add("Bridge", MonReq{
			Columns: []string{"name", "name"},
			Select: &Select{
				Initial: true,
			},
		})
		err = reqs.Validate()
		assert.Error(t, err)
	})
	reqs = NewMonReqSet(&sch)
	t.Run("fail to add request with no select options", func(t *testing.T) {
		reqs.Add("Bridge", MonReq{
			Columns: []string{"name"},
			Select: &Select{
				Initial: false,
				Insert:  false,
				Delete:  false,
				Modify:  false,
			},
		})
		err = reqs.Validate()
		assert.Error(t, err)
	})
	reqs = NewMonReqSet(&sch)
	t.Run("add request with no columns and no select (true for all)", func(t *testing.T) {
		reqs.Add("Bridge", MonReq{})
		err = reqs.Validate()
		require.NoError(t, err, "fail to add request")
		assert.Equal(t, &monReqSet{
			sch:        &sch,
			hasInitial: true,
			hasUpdates: true,
			reqs: map[string][]MonReq{
				"Bridge": {
					{},
				},
			},
		}, reqs)
	})
	reqs = NewMonReqSet(&sch)
	t.Run("add two requests in row", func(t *testing.T) {
		reqs.Add("Bridge",
			MonReq{
				Columns: []string{"name"},
				Select: &Select{
					Initial: true,
					Modify:  true,
				},
			},
			MonReq{Columns: []string{"fail_mode", "ports"}})
		require.NoError(t, err, "fail to add more request to same table")
		err = reqs.Validate()
		assert.Equal(t, &monReqSet{
			sch:        &sch,
			hasInitial: true,
			hasUpdates: true,
			reqs: map[string][]MonReq{
				"Bridge": {
					{
						Columns: []string{"name"},
						Select: &Select{
							Initial: true,
							Modify:  true,
						},
					},
					{
						Columns: []string{"fail_mode", "ports"},
					},
				},
			},
		}, reqs)
	})
	t.Run("fail to add non-single all-column", func(t *testing.T) {
		reqs = NewMonReqSet(&sch)
		reqs.Add("Bridge", MonReq{})
		reqs.Add("Port", MonReq{Columns: []string{"name", "tag"}})
		reqs.Add("Bridge", MonReq{})
		err = reqs.Validate()
		assert.Error(t, err, "succeed to add two all-column request")
		reqs = NewMonReqSet(&sch)
		reqs.Add("Bridge", MonReq{})
		reqs.Add("Port", MonReq{Columns: []string{"name", "tag"}})
		reqs.Add("Bridge", MonReq{Columns: []string{"name", "tag"}})
		err = reqs.Validate()
		assert.Error(t, err, "succeed to mix column to all-column request")
		reqs = NewMonReqSet(&sch)
		reqs.Add("Bridge", MonReq{})
		reqs.Add("Port", MonReq{Columns: []string{"name", "tag"}})
		reqs.Add("Port", MonReq{})
		err = reqs.Validate()
		assert.Error(t, err, "succeed to mix all-column to column request")
		reqs = NewMonReqSet(&sch)
		reqs.Add("Bridge", MonReq{})
		reqs.Add("Port", MonReq{Columns: []string{"name", "tag"}})
		reqs = NewMonReqSet(&sch)
		reqs.Add("Bridge", MonReq{Columns: []string{"name", "tag"}}, MonReq{})
		err = reqs.Validate()
		assert.Error(t, err, "succeed to mix all-column and column request in one call (columns + all-columns)")
		reqs = NewMonReqSet(&sch)
		reqs.Add("Bridge", MonReq{})
		reqs.Add("Port", MonReq{Columns: []string{"name", "tag"}})
		reqs.Add("Bridge", MonReq{}, MonReq{Columns: []string{"name", "tag"}})
		err = reqs.Validate()
		assert.Error(t, err, "succeed to mix all-column and column request in one call (all-columns + columns)")
		reqs = NewMonReqSet(&sch)
		reqs.Add("Bridge", MonReq{})
		reqs.Add("Port", MonReq{Columns: []string{"name", "tag"}})
		reqs.Add("Bridge", MonReq{}, MonReq{})
		err = reqs.Validate()
		assert.Error(t, err, "succeed to mix 2 all-column request in one call (all-columns + all-columns)")
	})
	t.Run("add request with columns duplicating already added request for same table", func(t *testing.T) {
		reqs = NewMonReqSet(&sch)
		reqs.Add("Bridge", MonReq{Columns: []string{"name"}})
		reqs.Add("Bridge", MonReq{
			Columns: []string{"name", "controller"},
			Select: &Select{
				Initial: true,
			},
		})
		err = reqs.Validate()
		assert.Error(t, err)
		reqs = NewMonReqSet(&sch)
		reqs.Add("Bridge", MonReq{
			Columns: []string{"controller", "fail_mode"},
		}, MonReq{
			Columns: []string{"controller", "ports"},
		})
		err = reqs.Validate()
		assert.Error(t, err)
	})
}

func Test_monReqs_MarshalJSON(t *testing.T) {
	var sch schema.DbSchema
	err := json.Unmarshal(ovsSchema, &sch)
	require.NoError(t, err, "fail to load schema")
	reqs := NewMonReqSet(&sch)
	reqs.Add("Bridge", MonReq{
		Columns: []string{"name"},
		Select: &Select{
			Initial: true,
		},
	})
	require.NoError(t, reqs.Validate(), "fail to add request")
	reqs.Add("Port",
		MonReq{
			Columns: []string{"name", "tag"},
			Select: &Select{
				Modify: true,
				Delete: true,
			},
		},
		MonReq{
			Columns: []string{"trunks"},
			Select:  nil,
		})
	require.NoError(t, reqs.Validate(), "fail to add request")
	reqs.Add("Interface", MonReq{})
	require.NoError(t, reqs.Validate(), "fail to add request")
	b, err := json.Marshal(reqs)
	require.NoError(t, err, "fail to marshal")
	assert.JSONEq(t, `{
									"Bridge":[
										{
											"columns":["name"],
											"select":{"initial":true,"modify":false,"delete":false,"insert":false}
										}
									],
									"Port":[
										{
											"columns":["name","tag"],
											"select":{"modify":true,"delete":true,"initial":false,"insert":false}
										},
										{
											"columns":["trunks"]
										}
									],
									"Interface":[
										{}
									]		
								}`, string(b), "incorrect marshal")
}

func Test_monReqs_UnmarshalJSON(t *testing.T) {
}
