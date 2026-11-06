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

// Host-Test der Packages-Session (respondd gegen echten ddhcpd im Namespace),
// 06.11.2026, kein echter Knoten
const nfDDHCPD = `{"node_id":"deadbeef","neanderfunk":{
 "ddhcpd":{"running":true,"learning":false,
  "blocks":{"total":64,"free":62,"ours":1,"claiming":1,"foreign":0,"releasing":0},
  "leases":{"total":4,"leased":2,"offered":0,"free":2,"away":1},
  "roaming":{"in":0,"out":0,"ack_sent":0,"ack_received":0,"nak_sent":0,"nak_received":0},
  "dhcp":{"discover":2,"offer":2,"request":2,"ack":2,"nak":0,"release":0,"inform":0},
  "conflicts":{"won":0,"lost":1,"handover_ack":3,"handover_nak":2}},
 "clients_ipv4":{"total":2,"ddhcpd":1}}}`

func TestNeanderfunkDDHCPD(t *testing.T) {
	assert := assert.New(t)
	f, _ := punkte(t, nfDDHCPD)["node"][0].Fields()
	assert.EqualValues(1, f["nf.ddhcpd.running"])
	assert.EqualValues(0, f["nf.ddhcpd.learning"])
	assert.EqualValues(64, f["nf.ddhcpd.blocks.total"])
	assert.EqualValues(0, f["nf.ddhcpd.blocks.foreign"])
	assert.EqualValues(1, f["nf.ddhcpd.leases.away"])
	assert.EqualValues(0, f["nf.ddhcpd.roaming.in"])
	assert.EqualValues(2, f["nf.ddhcpd.dhcp.ack"])
	assert.EqualValues(0, f["nf.ddhcpd.blocks.releasing"])
	assert.EqualValues(2, f["nf.ddhcpd.conflicts.handover_nak"])
	assert.EqualValues(2, f["nf.clients_ipv4.total"])
	assert.EqualValues(1, f["nf.clients_ipv4.ddhcpd"])

	// Daemon antwortet nicht: nur running, keine erfundenen Nullen
	f, _ = punkte(t, `{"node_id":"deadbeef","neanderfunk":{"ddhcpd":{"running":false},
	 "clients_ipv4":{"total":3,"ddhcpd":0}}}`)["node"][0].Fields()
	assert.EqualValues(0, f["nf.ddhcpd.running"])
	assert.NotContains(f, "nf.ddhcpd.learning")
	assert.NotContains(f, "nf.ddhcpd.blocks.total")
	assert.EqualValues(0, f["nf.clients_ipv4.ddhcpd"])

	// fremde Werte in einer Gruppe: der Rest kommt trotzdem an
	z := punkte(t, `{"node_id":"deadbeef","loadavg":0.5,"neanderfunk":{"ddhcpd":{"running":true,
	 "blocks":{"total":64,"neu":"text","tief":{"x":1}}}}}`)
	f, _ = z["node"][0].Fields()
	assert.EqualValues(64, f["nf.ddhcpd.blocks.total"])
	assert.NotContains(f, "nf.ddhcpd.blocks.neu")
	assert.NotContains(f, "nf.ddhcpd.blocks.tief")
	assert.EqualValues(0.5, f["load"])

	// kaputte Teile: running als Text, Gruppe als Zahl. Der Rest bleibt.
	z = punkte(t, `{"node_id":"deadbeef","loadavg":0.5,"neanderfunk":{"ddhcpd":{"running":"ja",
	 "learning":true,"dhcp":5,"blocks":{"total":64,"ours":0}}}}`)
	f, _ = z["node"][0].Fields()
	assert.NotContains(f, "nf.ddhcpd.running")
	assert.EqualValues(1, f["nf.ddhcpd.learning"])
	assert.EqualValues(64, f["nf.ddhcpd.blocks.total"])
	assert.EqualValues(0.5, f["load"])

	// state-Datei: Rundreise ohne Verlust
	var s data.Statistics
	assert.NoError(json.Unmarshal([]byte(nfDDHCPD), &s))
	b, err := json.Marshal(&s)
	assert.NoError(err)
	var s2 data.Statistics
	assert.NoError(json.Unmarshal(b, &s2))
	assert.Equal(s.Neanderfunk.DDHCPD, s2.Neanderfunk.DDHCPD)

	// ohne Paket
	f, _ = punkte(t, `{"node_id":"deadbeef","neanderfunk":{"system":{"forks":5}}}`)["node"][0].Fields()
	assert.NotContains(f, "nf.ddhcpd.running")
	assert.NotContains(f, "nf.clients_ipv4.total")
}
