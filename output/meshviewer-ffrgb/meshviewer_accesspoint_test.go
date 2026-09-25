package meshviewerFFRGB

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FreifunkBremen/yanic/data"
	"github.com/FreifunkBremen/yanic/runtime"
)

func nodeinfoAusJSON(t *testing.T, raw string) *data.Nodeinfo {
	var ni data.Nodeinfo
	if err := json.Unmarshal([]byte(raw), &ni); err != nil {
		t.Fatal(err)
	}
	return &ni
}

// Den Link eines Accesspoints zu seinem Router meldet nur der AP. Er soll
// trotzdem mit voller Qualitaet auf beiden Seiten erscheinen.
func TestTransformAccessPointLinkBeidseitig(t *testing.T) {
	assert := assert.New(t)

	nodes := runtime.NewNodes(&runtime.NodesConfig{})
	nodes.AddNode(&runtime.Node{
		Online: true,
		Nodeinfo: nodeinfoAusJSON(t, `{"node_id":"router","network":{"mac":"aa:aa:aa:aa:aa:01"},
			"software":{"firmware":{"base":"gluon-v2023.2.5"}}}`),
	})
	nodes.AddNode(&runtime.Node{
		Online: true,
		Nodeinfo: nodeinfoAusJSON(t, `{"node_id":"ap1","network":{"mac":"0c:ea:14:00:00:01",
			"mesh":{"bat0":{"interfaces":{"other":["0c:ea:14:00:00:01"]}}}},
			"software":{"firmware":{"base":"UniFi"}}}`),
		Neighbours: &data.Neighbours{
			NodeID: "ap1",
			Batadv: map[string]data.BatadvNeighbours{
				"0c:ea:14:00:00:01": {Neighbours: map[string]data.BatmanLink{
					"aa:aa:aa:aa:aa:01": {TQ: 255},
				}},
			},
		},
	})

	meshviewer := transform(nodes)
	assert.Len(meshviewer.Links, 1)
	link := meshviewer.Links[0]
	assert.Equal(float32(1), link.SourceTQ)
	assert.Equal(float32(1), link.TargetTQ)
	assert.Equal("other", link.Type)
	assert.ElementsMatch([]string{"router", "ap1"}, []string{link.Source, link.Target})
}
