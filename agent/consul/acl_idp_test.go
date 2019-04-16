package consul

import (
	"testing"

	"github.com/hashicorp/consul/agent/structs"
	"github.com/stretchr/testify/require"
)

func TestDoesBindingRuleMatch(t *testing.T) {
	type matchable struct {
		A string `bexpr:"a"`
		C string `bexpr:"c"`
	}

	for _, test := range []struct {
		name     string
		selector string
		details  interface{}
		ok       bool
	}{
		{"no fields",
			"a==b", nil, false},
		{"1 term ok",
			"a==b", &matchable{A: "b"}, true},
		{"1 term no field",
			"a==b", &matchable{C: "d"}, false},
		{"1 term wrong value",
			"a==b", &matchable{A: "z"}, false},
		{"2 terms ok",
			"a==b and c==d", &matchable{A: "b", C: "d"}, true},
		{"2 terms one missing field",
			"a==b and c==d", &matchable{A: "b"}, false},
		{"2 terms one wrong value",
			"a==b and c==d", &matchable{A: "z", C: "d"}, false},
		///////////////////////////////
		{"no fields (no selectors)",
			"", nil, true},
		{"1 term ok (no selectors)",
			"", &matchable{A: "b"}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			rule := structs.ACLBindingRule{Selector: test.selector}
			ok := doesBindingRuleMatch(&rule, test.details)
			require.Equal(t, test.ok, ok)
		})
	}
}

// mapify takes pairwise string fields and inserts them into a map
//
// example: mapify("a","b","c","d") => {"a":"b", "c":"d"}
func mapify(kvs ...string) map[string]string {
	if len(kvs)%2 != 0 {
		panic("function takes an even number of args")
	}
	out := make(map[string]string)
	for i := 0; i < len(kvs); i += 2 {
		out[kvs[i]] = kvs[i+1]
	}
	return out
}
