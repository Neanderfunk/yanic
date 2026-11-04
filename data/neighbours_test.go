package data

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Aeltere Gluon-Versionen schicken wifi als Array oder je Schnittstelle ein
// Array; das darf die Antwort nicht verwerfen (Freifunk Essen, 03.11.2026).
func TestNeighboursWifiTolerant(t *testing.T) {
	assert := assert.New(t)
	faelle := map[string]int{
		`{"node_id":"a","wifi":{"aa:bb:cc:dd:ee:ff":{"frequency":2412,"neighbours":{}}}}`: 2412,
		`{"node_id":"a","wifi":{"aa:bb:cc:dd:ee:ff":[]}}`:                                 0,
		`{"node_id":"a","wifi":{"aa:bb:cc:dd:ee:ff":{"neighbours":[]}}}`:                  0,
		`{"node_id":"a","wifi":{"aa:bb:cc:dd:ee:ff":{"frequency":"x"}}}`:                  0,
	}
	for roh, freq := range faelle {
		var n Neighbours
		assert.NoError(json.Unmarshal([]byte(roh), &n), roh)
		assert.Equal("a", n.NodeID)
		assert.Equal(freq, n.Wifi["aa:bb:cc:dd:ee:ff"].Frequency, roh)
	}
	var n Neighbours
	assert.NoError(json.Unmarshal([]byte(`{"node_id":"b","wifi":[],"batadv":{}}`), &n))
	assert.Equal("b", n.NodeID)
	assert.Empty(n.Wifi)
}
