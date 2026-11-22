package influxdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Beispiel aus neanderfunk-respondd/README.md (Feed b99b4e6)
const dhSample = `{"node_id":"deadbeef","neanderfunk":{"domainhelper":{
 "bridge":{"state":"cut","where":"lan3","domain":"18_nefuk","node":"dceae7aeb483","note":"getrennt","age":100},
 "follow":{"domain":"18_nefuk","rule":"R1","age":600,"wait":3600},
 "island":{"age":300},
 "last_switch":{"from":"21_dias","to":"18_nefuk","how":"follow-R1","time":1795300000}}}}`

func TestDomainHelper(t *testing.T) {
	assert := assert.New(t)
	z := punkte(t, dhSample)

	assert.Len(z["nf_domainhelper_last_switch"], 1)
	ls := z["nf_domainhelper_last_switch"][0]
	f, _ := ls.Fields()
	assert.EqualValues(1795300000, f["time"])
	assert.Equal("21_dias", ls.Tags()["from"])
	assert.Equal("18_nefuk", ls.Tags()["to"])
	assert.Equal("follow-R1", ls.Tags()["how"])
	assert.Equal("deadbeef", ls.Tags()["nodeid"])

	b := z["nf_domainhelper_bridge"][0]
	f, _ = b.Fields()
	assert.EqualValues(100, f["age"])
	assert.Equal("cut", b.Tags()["state"])
	assert.Equal("18_nefuk", b.Tags()["peer_domain"])
	assert.Equal("dceae7aeb483", b.Tags()["peer_node"])
	assert.Equal("lan3", b.Tags()["where"])

	fo := z["nf_domainhelper_follow"][0]
	f, _ = fo.Fields()
	assert.EqualValues(600, f["age"])
	assert.EqualValues(3600, f["wait"])
	assert.Equal("R1", fo.Tags()["rule"])
	assert.Equal("18_nefuk", fo.Tags()["follow_domain"])

	f, _ = z["nf_domainhelper_island"][0].Fields()
	assert.EqualValues(300, f["age"])
}

// Normalfall: der Block fehlt, nichts wird geschrieben
func TestDomainHelperFehlt(t *testing.T) {
	z := punkte(t, nfSample)
	for _, m := range []string{"nf_domainhelper_last_switch", "nf_domainhelper_bridge",
		"nf_domainhelper_follow", "nf_domainhelper_island"} {
		assert.Empty(t, z[m], m)
	}
}

// Falscher Typ in einem Teil: nur der Teil faellt weg, der Rest der
// Statistik (loadavg, last_switch) kommt an
func TestDomainHelperTolerant(t *testing.T) {
	assert := assert.New(t)
	z := punkte(t, `{"node_id":"deadbeef","loadavg":0.5,"neanderfunk":{"domainhelper":{
 "bridge":[],"island":{"age":"lang"},
 "last_switch":{"from":"21_dias","to":"05_mon","how":"nodeplacer","time":1795300000}}}}`)
	f, _ := z["node"][0].Fields()
	assert.EqualValues(0.5, f["load"])
	assert.Len(z["nf_domainhelper_last_switch"], 1)
	assert.Empty(z["nf_domainhelper_bridge"])
	assert.Empty(z["nf_domainhelper_island"])
}
