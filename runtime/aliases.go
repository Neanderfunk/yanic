package runtime

// Lokaler Zusatz (Neanderfunk): Aliase wie bei hopglass-server.
//
// Eine Datei (nodes.aliases_path) legt je node_id Teile der nodeinfo ueber das,
// was der Knoten meldet, im Format von hopglass-server:
//
//	{"e8ab2f35b3af": {"nodeinfo": {"location": {"latitude": 51.2, "longitude": 6.9}}}}
//
// Gemischt wird nach RFC 7396 (JSON Merge Patch): Werte ersetzen, Objekte
// werden rekursiv gemischt, null entfernt das Feld. "location": null nimmt
// einen falsch platzierten Knoten von der Landkarte, ohne ihn aus Liste und
// Statistik zu nehmen.
//
// Die gemeldete nodeinfo bleibt in NodeinfoOriginal erhalten. Faellt ein
// Alias weg, gilt sofort wieder die gemeldete, auch bei einem Knoten, der
// offline ist und nichts Neues meldet.

import (
	"encoding/json"
	"os"
	"time"

	"github.com/bdlm/log"

	"github.com/FreifunkBremen/yanic/data"
)

type aliasEntry struct {
	Nodeinfo json.RawMessage `json:"nodeinfo"`
}

// loadAliases liest die Datei neu, wenn sie sich geaendert hat. Gibt true
// zurueck, wenn sich die Aliase geaendert haben. Aufruf mit nodes.Lock().
func (nodes *Nodes) loadAliases() bool {
	if nodes.config == nil || nodes.config.AliasesPath == "" {
		return false
	}
	info, err := os.Stat(nodes.config.AliasesPath)
	if err != nil {
		if nodes.aliases != nil {
			log.WithError(err).Warn("aliases: Datei nicht lesbar, alle Aliase aufgehoben")
			nodes.aliases = nil
			nodes.aliasesTime = time.Time{}
			return true
		}
		return false
	}
	if info.ModTime().Equal(nodes.aliasesTime) {
		return false
	}
	raw, err := os.ReadFile(nodes.config.AliasesPath)
	if err != nil {
		log.WithError(err).Error("aliases: lesen")
		return false
	}
	var entries map[string]aliasEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		// Eine kaputte Datei aendert nichts; die bisherigen Aliase bleiben.
		log.WithError(err).Error("aliases: kein gueltiges JSON, bisherige Aliase bleiben")
		return false
	}
	patches := make(map[string]json.RawMessage, len(entries))
	for id, e := range entries {
		if len(e.Nodeinfo) > 0 {
			patches[id] = e.Nodeinfo
		}
	}
	nodes.aliases = patches
	nodes.aliasesTime = info.ModTime()
	log.WithField("count", len(patches)).Info("aliases geladen")
	return true
}

// mergePatch nach RFC 7396.
func mergePatch(target, patch interface{}) interface{} {
	p, ok := patch.(map[string]interface{})
	if !ok {
		return patch
	}
	t, ok := target.(map[string]interface{})
	if !ok {
		t = map[string]interface{}{}
	}
	for k, v := range p {
		if v == nil {
			delete(t, k)
		} else {
			t[k] = mergePatch(t[k], v)
		}
	}
	return t
}

// applyAlias legt den Alias ueber die gemeldete nodeinfo oder nimmt ihn
// zurueck. Aufruf mit nodes.Lock().
func (nodes *Nodes) applyAlias(nodeID string, node *Node) {
	patch, ok := nodes.aliases[nodeID]
	if !ok {
		if node.NodeinfoOriginal != nil {
			node.Nodeinfo = node.NodeinfoOriginal
			node.NodeinfoOriginal = nil
		}
		return
	}
	base := node.NodeinfoOriginal
	if base == nil {
		base = node.Nodeinfo
	}
	if base == nil {
		return
	}
	raw, err := json.Marshal(base)
	if err != nil {
		return
	}
	var ziel, p interface{}
	if json.Unmarshal(raw, &ziel) != nil || json.Unmarshal(patch, &p) != nil {
		return
	}
	neu, err := json.Marshal(mergePatch(ziel, p))
	if err != nil {
		return
	}
	var ni data.Nodeinfo
	if err := json.Unmarshal(neu, &ni); err != nil {
		log.WithError(err).WithField("node_id", nodeID).Warn("aliases: Ergebnis ist keine nodeinfo")
		return
	}
	node.NodeinfoOriginal = base
	node.Nodeinfo = &ni
}

// applyAliasesAll nach dem (Neu-)Laden der Datei. Aufruf mit nodes.Lock().
func (nodes *Nodes) applyAliasesAll() {
	for id, node := range nodes.List {
		nodes.applyAlias(id, node)
	}
}
