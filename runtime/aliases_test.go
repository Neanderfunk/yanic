package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/FreifunkBremen/yanic/data"
)

func meldung(t *testing.T, id, hostname string, lat, lon float64) *data.ResponseData {
	var ni data.Nodeinfo
	raw := `{"node_id":"` + id + `","hostname":"` + hostname + `","network":{"mac":"02:00:00:00:00:01"},` +
		`"location":{"latitude":` + jsonFloat(lat) + `,"longitude":` + jsonFloat(lon) + `}}`
	if err := json.Unmarshal([]byte(raw), &ni); err != nil {
		t.Fatal(err)
	}
	return &data.ResponseData{Nodeinfo: &ni}
}

func jsonFloat(f float64) string {
	b, _ := json.Marshal(f)
	return string(b)
}

func aliasDatei(t *testing.T, pfad, inhalt string, zeit time.Time) {
	if err := os.WriteFile(pfad, []byte(inhalt), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(pfad, zeit, zeit); err != nil {
		t.Fatal(err)
	}
}

func neuLaden(nodes *Nodes) {
	nodes.Lock()
	if nodes.loadAliases() {
		nodes.applyAliasesAll()
	}
	nodes.Unlock()
}

func TestMergePatch(t *testing.T) {
	assert := assert.New(t)
	ziel := map[string]interface{}{"a": 1.0, "b": map[string]interface{}{"c": 2.0, "d": 3.0}}
	patch := map[string]interface{}{"a": nil, "b": map[string]interface{}{"c": 9.0}, "e": "neu"}
	assert.Equal(map[string]interface{}{"b": map[string]interface{}{"c": 9.0, "d": 3.0}, "e": "neu"},
		mergePatch(ziel, patch))
}

func TestAliasOrtVerschieben(t *testing.T) {
	assert := assert.New(t)
	pfad := filepath.Join(t.TempDir(), "aliases.json")
	aliasDatei(t, pfad, `{"n1": {"nodeinfo": {"location": {"latitude": 51.25, "longitude": 6.97}}}}`, time.Unix(1000, 0))
	nodes := NewNodes(&NodesConfig{AliasesPath: pfad})
	neuLaden(nodes)

	nodes.Update("n1", meldung(t, "n1", "knoten", -1.5, 2.5))
	n := nodes.List["n1"]
	assert.Equal(51.25, n.Nodeinfo.Location.Latitude)
	assert.Equal(6.97, n.Nodeinfo.Location.Longitude)
	assert.Equal("knoten", n.Nodeinfo.Hostname)
	assert.Equal(-1.5, n.NodeinfoOriginal.Location.Latitude)
}

func TestAliasVonDerKarteNehmen(t *testing.T) {
	assert := assert.New(t)
	pfad := filepath.Join(t.TempDir(), "aliases.json")
	aliasDatei(t, pfad, `{"n1": {"nodeinfo": {"location": null}}}`, time.Unix(1000, 0))
	nodes := NewNodes(&NodesConfig{AliasesPath: pfad})
	neuLaden(nodes)
	nodes.Update("n1", meldung(t, "n1", "knoten", 51.0, 7.0))
	assert.Nil(nodes.List["n1"].Nodeinfo.Location)
	assert.Equal("knoten", nodes.List["n1"].Nodeinfo.Hostname)
}

func TestAliasWegfallGiltSofortAuchOffline(t *testing.T) {
	assert := assert.New(t)
	pfad := filepath.Join(t.TempDir(), "aliases.json")
	aliasDatei(t, pfad, `{"n1": {"nodeinfo": {"hostname": "umbenannt"}}}`, time.Unix(1000, 0))
	nodes := NewNodes(&NodesConfig{AliasesPath: pfad})
	neuLaden(nodes)
	nodes.Update("n1", meldung(t, "n1", "knoten", 51.0, 7.0))
	assert.Equal("umbenannt", nodes.List["n1"].Nodeinfo.Hostname)

	// Knoten meldet nichts mehr; der Alias faellt weg
	nodes.List["n1"].Online = false
	aliasDatei(t, pfad, `{}`, time.Unix(2000, 0))
	neuLaden(nodes)
	assert.Equal("knoten", nodes.List["n1"].Nodeinfo.Hostname)
	assert.Nil(nodes.List["n1"].NodeinfoOriginal)
}

func TestAliasNeuGeladenAufBestehendeKnoten(t *testing.T) {
	assert := assert.New(t)
	pfad := filepath.Join(t.TempDir(), "aliases.json")
	nodes := NewNodes(&NodesConfig{AliasesPath: pfad})
	nodes.Update("n1", meldung(t, "n1", "knoten", 51.0, 7.0))
	assert.Equal(51.0, nodes.List["n1"].Nodeinfo.Location.Latitude)

	aliasDatei(t, pfad, `{"n1": {"nodeinfo": {"location": {"latitude": 50.0}}}}`, time.Unix(1000, 0))
	neuLaden(nodes)
	assert.Equal(50.0, nodes.List["n1"].Nodeinfo.Location.Latitude)
	assert.Equal(7.0, nodes.List["n1"].Nodeinfo.Location.Longitude)

	// Neue Meldung: Grundlage ist die frische nodeinfo, der Alias liegt wieder darueber
	nodes.Update("n1", meldung(t, "n1", "knoten", 52.0, 8.0))
	assert.Equal(50.0, nodes.List["n1"].Nodeinfo.Location.Latitude)
	assert.Equal(8.0, nodes.List["n1"].Nodeinfo.Location.Longitude)
	assert.Equal(52.0, nodes.List["n1"].NodeinfoOriginal.Location.Latitude)
}

func TestAliasKaputteDateiBehaeltAlte(t *testing.T) {
	assert := assert.New(t)
	pfad := filepath.Join(t.TempDir(), "aliases.json")
	aliasDatei(t, pfad, `{"n1": {"nodeinfo": {"hostname": "umbenannt"}}}`, time.Unix(1000, 0))
	nodes := NewNodes(&NodesConfig{AliasesPath: pfad})
	neuLaden(nodes)
	aliasDatei(t, pfad, `{ kaputt`, time.Unix(2000, 0))
	neuLaden(nodes)
	nodes.Update("n1", meldung(t, "n1", "knoten", 51.0, 7.0))
	assert.Equal("umbenannt", nodes.List["n1"].Nodeinfo.Hostname)
}

func TestAliasOhneKonfigurationNichts(t *testing.T) {
	assert := assert.New(t)
	nodes := NewNodes(&NodesConfig{})
	neuLaden(nodes)
	nodes.Update("n1", meldung(t, "n1", "knoten", 51.0, 7.0))
	assert.Equal("knoten", nodes.List["n1"].Nodeinfo.Hostname)
	assert.Nil(nodes.List["n1"].NodeinfoOriginal)
}
