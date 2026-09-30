package runtime

// Lokaler Zusatz (Neanderfunk): Ortsangaben mit einem Rahmen abgleichen.
//
// nodes.location_bbox = [sued, west, nord, ost]. Liegt der gemeldete Ort
// ausserhalb, der mit getauschter Breite und Laenge aber innerhalb, gilt der
// getauschte. In allen anderen Faellen, auch ohne Rahmen, bleibt der Ort,
// wie er ist. Aliase legen sich danach darueber und haben Vorrang.

import (
	"github.com/bdlm/log"

	"github.com/FreifunkBremen/yanic/data"
)

func imRahmen(rahmen []float64, breite, laenge float64) bool {
	return rahmen[0] <= breite && breite <= rahmen[2] && rahmen[1] <= laenge && laenge <= rahmen[3]
}

// ortImRahmen tauscht Breite und Laenge in nodeinfo, wenn nur die getauschte
// Lesart im Rahmen liegt. Rueckgabe: ob getauscht wurde.
func ortImRahmen(rahmen []float64, ni *data.Nodeinfo) bool {
	if len(rahmen) != 4 || ni == nil || ni.Location == nil {
		return false
	}
	ort := ni.Location
	if imRahmen(rahmen, ort.Latitude, ort.Longitude) || !imRahmen(rahmen, ort.Longitude, ort.Latitude) {
		return false
	}
	ort.Latitude, ort.Longitude = ort.Longitude, ort.Latitude
	return true
}

// locationUpdate gleicht die frisch gemeldete nodeinfo ab. Protokolliert wird
// nur, wenn sich der gespeicherte Ort dadurch aendert, nicht in jeder Runde.
func (nodes *Nodes) locationUpdate(nodeID string, vorher *data.Location, ni *data.Nodeinfo) {
	if nodes.config == nil || !ortImRahmen(nodes.config.LocationBBox, ni) {
		return
	}
	if vorher != nil && vorher.Latitude == ni.Location.Latitude && vorher.Longitude == ni.Location.Longitude {
		return
	}
	log.WithFields(map[string]interface{}{
		"node_id":   nodeID,
		"hostname":  ni.Hostname,
		"gemeldet":  []float64{ni.Location.Longitude, ni.Location.Latitude},
		"getauscht": []float64{ni.Location.Latitude, ni.Location.Longitude},
	}).Warn("Ort ausserhalb des Rahmens, Breite und Laenge getauscht")
}
