package influxdb

import (
	"encoding/json"
	"testing"

	"github.com/influxdata/influxdb1-client/v2"
	"github.com/stretchr/testify/assert"

	"github.com/FreifunkBremen/yanic/data"
	"github.com/FreifunkBremen/yanic/runtime"
)

// Echtes Sample, dias-Testknoten 33f1, 22.09.2026, ergaenzt um die
// Temperatur eines MT7915-Knotens und einen Port ohne Werte
const nfSample = `{"node_id":"deadbeef","loadavg":0.5,"neanderfunk":{
 "system":{"refault_file":0,"mem_available":16668,"zram":{"ram":0,"data":0,"size":0},"forks":13118},
 "ethernet":{"wan":{"speed":1000,"possible":0,"duplex":"full","carrier":true},
             "lan1":{"speed":100,"possible":0,"duplex":"full","carrier":false},
             "leer":{}},
 "ssid_changer":{"gateway_losses":0,"offline":0,"switches":2},
 "wireless":{"radio0":{"txpower":23,"ssid":"Freifunk","channel":9,"mesh":true,"htmode":"HT20","country":"TW"}},
 "temperature":{"soc":51.5,"mt7915_phy0":48}}}`

func punkte(t *testing.T, stats string) map[string][]*client.Point {
	var s data.Statistics
	assert.NoError(t, json.Unmarshal([]byte(stats), &s))
	node := &runtime.Node{Online: true, Statistics: &s,
		Nodeinfo: &data.Nodeinfo{NodeID: "deadbeef", Hostname: "knoten"}}
	conn := &Connection{points: make(chan *client.Point, 100), config: Config{}}
	conn.InsertNode(node)
	close(conn.points)
	z := map[string][]*client.Point{}
	for p := range conn.points {
		z[p.Name()] = append(z[p.Name()], p)
	}
	return z
}

func TestNeanderfunk(t *testing.T) {
	assert := assert.New(t)
	z := punkte(t, nfSample)

	f, _ := z["node"][0].Fields()
	assert.EqualValues(13118, f["nf.forks"])
	assert.EqualValues(0, f["nf.refault_file"])
	assert.EqualValues(0, f["nf.zram.size"])
	assert.EqualValues(2, f["nf.ssid_changer.switches"])
	assert.NotContains(f, "nf.mem_available")

	assert.Len(z["nf_ethernet"], 2, "der leere Port faellt weg")
	for _, p := range z["nf_ethernet"] {
		f, _ := p.Fields()
		if p.Tags()["port"] == "wan" {
			assert.EqualValues(1, f["carrier"])
			assert.EqualValues(1000, f["speed"])
			assert.Equal("full", p.Tags()["duplex"])
			assert.Equal("deadbeef", p.Tags()["nodeid"])
		} else {
			assert.EqualValues(0, f["carrier"])
		}
	}

	assert.Len(z["nf_temperature"], 2)
	for _, p := range z["nf_temperature"] {
		f, _ := p.Fields()
		if p.Tags()["sensor"] == "soc" {
			assert.EqualValues(51.5, f["celsius"])
		}
	}

	assert.Len(z["nf_wireless"], 1)
	w := z["nf_wireless"][0]
	wf, _ := w.Fields()
	assert.EqualValues(23, wf["txpower"])
	assert.EqualValues(1, wf["mesh"])
	assert.Equal("HT20", w.Tags()["htmode"])
	assert.Equal("radio0", w.Tags()["radio"])
}

// Alt-Firmware ohne das Paket und Knoten ohne Sensor: nichts erfinden
func TestNeanderfunkFehlt(t *testing.T) {
	assert := assert.New(t)
	z := punkte(t, `{"node_id":"deadbeef","loadavg":0.5}`)
	f, _ := z["node"][0].Fields()
	assert.NotContains(f, "nf.forks")
	assert.Empty(z["nf_ethernet"])
	assert.Empty(z["nf_temperature"])

	z = punkte(t, `{"node_id":"deadbeef","neanderfunk":{"system":{"forks":5},"temperature":{}}}`)
	f, _ = z["node"][0].Fields()
	assert.EqualValues(5, f["nf.forks"])
	assert.NotContains(f, "nf.refault_file")
	assert.Empty(z["nf_temperature"])
}
