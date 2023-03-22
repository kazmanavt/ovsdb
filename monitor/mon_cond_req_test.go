package monitor

import (
	"github.com/kazmanavt/ovsdb/schema"
	"github.com/kazmanavt/ovsdb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_monCondReqs_Add(t *testing.T) {
	var sch schema.DbSchema
	err := sch.UnmarshalJSON(ovsSchema)
	require.NoError(t, err, "fail to load schema")

	failTests := []struct {
		msg   string
		tName string
		cond  types.Condition
	}{
		{
			msg:   "succeed to add request with wrong columns in condition",
			tName: "Bridge",
			cond:  types.Equal("non-existing-column", "br0"),
		},
		{
			msg:   "succeed to add request with wrong value type in condition",
			tName: "Bridge",
			cond:  types.Equal("name", 1),
		},
		{
			msg:   "succeed to add request with wrong operator in condition",
			tName: "Bridge",
			cond:  types.LessEqual("name", "br0"),
		},
		{
			msg:   "succeed to add request with value out of range in condition",
			tName: "Port",
			cond:  types.Equal("tag", 65536),
		},
		{
			msg:   "succeed to add request violating column constraint on value in condition",
			tName: "Port",
			cond:  types.Equal("tag", types.Set[int]{44, 33}),
		},
		{
			msg:   "succeed to add request violating column constraint on value in condition 2",
			tName: "NetFlow",
			cond:  types.Equal("active_timeout", -2),
		},
		{
			msg:   "succeed to add request with cond value of wrong type",
			tName: "Flow_Sample_Collector_Set",
			cond:  types.Equal("bridge", "non-uuid"),
		},
		{
			msg:   "succeed to add request with cond value of wrong type 2",
			tName: "Bridge",
			cond:  types.Equal("fail_mode", types.Set[string]{"wrong", "value"}),
		},
		{
			msg:   "succeed to add request violating column constraint on value in condition 3",
			tName: "Bridge",
			cond:  types.Equal("flow_tables", types.Map[int, types.UUID]{1: "uuid", 2000: "uuid"}),
		},
	}

	for _, test := range failTests {
		t.Run("should fail", func(t *testing.T) {
			reqs := NewMonCondReqs(&sch)
			reqs.Add(test.tName, MonCondReq{Where: []types.Condition{test.cond}})
			err = reqs.Validate()
			assert.Error(t, err, test.msg)
		})
	}

	okTests := []struct {
		msg   string
		tName string
		cond  types.Condition
	}{
		{
			msg:   "fail to add request with scalar for Set of 0 or 1",
			tName: "Port",
			cond:  types.Equal("tag", 65),
		},
		{
			msg:   "fail to add request with Set for Set of 0 or 1",
			tName: "Port",
			cond:  types.Equal("tag", types.Set[int]{65}),
		},
		{
			msg:   "fail to add request integer",
			tName: "NetFlow",
			cond:  types.Equal("active_timeout", -1),
		},
		{
			msg:   "fail to add request with Set of 0 or 1",
			tName: "Port",
			cond:  types.Includes("interfaces", types.Set[types.UUID]{"iii", "jjj"}),
		},
		{
			msg:   "fail to add request with Map",
			tName: "Port",
			cond:  types.Equal("external_ids", types.Map[string, string]{"key": "value"}),
		},
		{
			msg:   "fail to add request with Map",
			tName: "Bridge",
			cond:  types.Equal("flow_tables", types.Map[int, types.UUID]{22: "uuid", 2: "uuid"}),
		},
	}
	for _, test := range okTests {
		t.Run("should succeed", func(t *testing.T) {
			reqs := NewMonCondReqs(&sch)
			reqs.Add(test.tName, MonCondReq{Where: []types.Condition{test.cond}})
			err = reqs.Validate()
			assert.NoError(t, err, test.msg)
		})
	}

}

func Test_monCondReqs_MarshalJSON(t *testing.T) {
	var sch schema.DbSchema
	err := sch.UnmarshalJSON(ovsSchema)
	require.NoError(t, err, "fail to load schema")

	reqs := NewMonCondReqs(&sch)
	reqs.Add("Bridge",
		MonCondReq{Columns: []string{"name", "ports"}, Where: []types.Condition{types.Equal("name", "br0")}, Select: &Select{Modify: true}},
		MonCondReq{Columns: []string{"other_config"}, Where: []types.Condition{types.Excludes("external_ids", types.Map[string, string]{"key": "value"})}, Select: &Select{Initial: true}},
	)
	err = reqs.Validate()
	require.NoError(t, err, "fail to add request")

	b, err := reqs.MarshalJSON()
	require.NoError(t, err, "fail to marshal requests")
	assert.JSONEq(t, `{
			"Bridge":[
				{
					"columns":["name","ports"],
					"where":[["name", "==","br0"]],
					"select":{"modify":true, "initial":false, "insert":false, "delete":false}
				},
				{
					"columns":["other_config"],
					"where":[
						[
							"external_ids",
							"excludes",
							[
								"map",
								[
									["key", "value"]
								]
							]
						]
					],
					"select":{"initial":true,"insert":false,"modify":false,"delete":false}
				}
			]
		}`, string(b))

	//{"columns":["name","ports"],"where":[{"op":"==","column":"name","value":"br0"}]}`, string(b))
}
