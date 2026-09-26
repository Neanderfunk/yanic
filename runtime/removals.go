package runtime

// Lokaler Zusatz (Neanderfunk): Offline-Knoten auf Anforderung entfernen.
//
// Ein Knoten, der dauerhaft weg ist, steht sonst bis prune_after (bei uns 90
// Tage) als offline in den Listen. Ueber das Service-Menue der Karte kann
// ihn jemand sofort entfernen: der Dienst legt dafuer eine leere Datei mit
// der node_id als Namen in nodes.remove_dir. yanic arbeitet die Auftraege
// bei jedem Speichern ab und entfernt einen Knoten nur, wenn er offline ist;
// einer, der gerade online ist, bleibt stehen. Meldet sich ein entfernter
// Knoten spaeter wieder, erscheint er ganz normal neu.

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/bdlm/log"
)

var removalName = regexp.MustCompile(`^[0-9A-Za-z_-]{1,64}$`)

// processRemovals arbeitet die Auftraege ab und loescht die Auftragsdateien.
// Gibt die Zahl der entfernten Knoten zurueck. Aufruf mit nodes.Lock().
func (nodes *Nodes) processRemovals() int {
	if nodes.config == nil || nodes.config.RemoveDir == "" {
		return 0
	}
	eintraege, err := os.ReadDir(nodes.config.RemoveDir)
	if err != nil {
		return 0
	}
	entfernt := 0
	for _, e := range eintraege {
		name := e.Name()
		if e.IsDir() || !removalName.MatchString(name) {
			continue
		}
		pfad := filepath.Join(nodes.config.RemoveDir, name)
		node, ok := nodes.List[name]
		switch {
		case !ok:
			log.WithField("node_id", name).Info("remove: unbekannt")
		case node.Online:
			log.WithField("node_id", name).Warn("remove: online, bleibt stehen")
		default:
			delete(nodes.List, name)
			delete(nodes.apRouter, name)
			entfernt++
			log.WithField("node_id", name).Info("remove: offline-Knoten entfernt")
		}
		if err := os.Remove(pfad); err != nil {
			log.WithError(err).WithField("node_id", name).Error("remove: Auftrag nicht loeschbar")
		}
	}
	return entfernt
}
