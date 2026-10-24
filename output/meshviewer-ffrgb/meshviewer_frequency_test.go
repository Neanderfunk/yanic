package meshviewerFFRGB

import (
	"encoding/json"
	"testing"

	"github.com/FreifunkBremen/yanic/data"
	"github.com/FreifunkBremen/yanic/runtime"
	"github.com/stretchr/testify/assert"
)

// Lokaler Zusatz: die Frequenz aus neighbours.wifi landet auf der Seite des
// Links, deren Schnittstelle sie meldet; ohne Angabe fehlt das Feld.
func TestTransformFrequency(t *testing.T) {
	assert := assert.New(t)

	knoten := func(id, mac string, wifi map[string]data.WifiInterface, nachbar string) *runtime.Node {
		return &runtime.Node{
			Online: true,
			Nodeinfo: &data.Nodeinfo{
				NodeID: id,
				Network: data.Network{
					Mac: mac,
					Mesh: map[string]*data.NetworkInterface{
						"bat0": {
							Interfaces: struct {
								Wireless []string `json:"wireless,omitempty"`
								Other    []string `json:"other,omitempty"`
								Tunnel   []string `json:"tunnel,omitempty"`
							}{
								Wireless: []string{mac + ":wifi"},
							},
						},
					},
				},
			},
			Neighbours: &data.Neighbours{
				NodeID: id,
				Wifi:   wifi,
				Batadv: map[string]data.BatadvNeighbours{
					mac + ":wifi": {
						Neighbours: map[string]data.BatmanLink{
							nachbar: {TQ: 204},
						},
					},
				},
			},
		}
	}

	nodes := runtime.NewNodes(&runtime.NodesConfig{})
	nodes.AddNode(knoten("node_a", "node:a:mac", map[string]data.WifiInterface{
		"node:a:mac:wifi": {Frequency: 5180},
	}, "node:b:mac:wifi"))
	nodes.AddNode(knoten("node_b", "node:b:mac", nil, "node:a:mac:wifi"))

	meshviewer := transform(nodes)
	assert.Len(meshviewer.Links, 1)
	link := meshviewer.Links[0]
	assert.Equal("wifi", link.Type)
	if link.SourceAddress == "node:a:mac:wifi" {
		assert.Equal(5180, link.SourceFrequency)
		assert.Equal(0, link.TargetFrequency)
	} else {
		assert.Equal(0, link.SourceFrequency)
		assert.Equal(5180, link.TargetFrequency)
	}

	// Ohne Angabe kein Feld im JSON
	roh, err := json.Marshal(&Link{Type: "wifi"})
	assert.NoError(err)
	assert.NotContains(string(roh), "frequency")

	// respondd-JSON: Nachbarwerte unter wifi werden uebergangen
	var n data.Neighbours
	assert.NoError(json.Unmarshal([]byte(`{"node_id":"x","wifi":{"aa:bb":{"frequency":2412,"neighbours":{"cc:dd":{"signal":-60,"noise":-95,"inactive":10}}}}}`), &n))
	assert.Equal(2412, n.Wifi["aa:bb"].Frequency)
}
