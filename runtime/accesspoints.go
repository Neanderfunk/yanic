package runtime

import "sort"

// Lokaler Zusatz (Neanderfunk): Accesspoints hinter Freifunk-Routern.
//
// Einen Accesspoint, den ein Stellvertreter wie unifi_respondd meldet, sieht
// batman nicht als Nachbarn: er haengt als gewoehnliches Geraet am LAN eines
// Freifunk-Routers, und seine WLAN-Clients stehen in der Uebersetzungstabelle
// als Clients dieses Routers. Ohne Korrektur zaehlen sie doppelt, einmal beim
// AP und einmal beim Router. Gemessen beim LVR am 25.09.2026: 836 von 837
// AP-Clients in der Tabelle hinter dem Router ihres APs, an 105 Routern
// 1007 von 1291 gemeldeten Clients.

// defaultAccessPointFirmware gilt, wenn accesspoint_firmware nicht gesetzt ist.
var defaultAccessPointFirmware = []string{"UniFi"}

func (nodes *Nodes) accessPointFirmware() []string {
	if nodes.config == nil || nodes.config.AccessPointFirmware == nil {
		return defaultAccessPointFirmware
	}
	return nodes.config.AccessPointFirmware
}

// IsAccessPoint sagt, ob der Knoten ein Accesspoint hinter einem Router ist,
// erkannt an software.firmware.base. Nimmt keine Sperre.
func (nodes *Nodes) IsAccessPoint(node *Node) bool {
	if node == nil || node.Nodeinfo == nil || node.Nodeinfo.Software.Firmware == nil {
		return false
	}
	base := node.Nodeinfo.Software.Firmware.Base
	for _, b := range nodes.accessPointFirmware() {
		if b != "" && b == base {
			return true
		}
	}
	return false
}

// routerOf liefert den Router, den ein AP als batman-Nachbarn meldet.
func (nodes *Nodes) routerOf(apID string, ap *Node) string {
	if ap.Neighbours == nil {
		return ""
	}
	var macs []string
	for _, batadv := range ap.Neighbours.Batadv {
		for mac := range batadv.Neighbours {
			macs = append(macs, mac)
		}
	}
	sort.Strings(macs)
	for _, mac := range macs {
		id := nodes.ifaceToNodeID[mac]
		if id == "" || id == apID {
			continue
		}
		if other := nodes.List[id]; other != nil && nodes.IsAccessPoint(other) {
			continue
		}
		return id
	}
	return ""
}

// accessPointUpdate merkt sich bei einem AP seinen Router und zieht bei einem
// Router die Clients seiner APs von clients.total ab. Die eigenen WLAN-Clients
// des Routers bleiben immer stehen: melden die APs mehr, als der Router ueber
// Kabel zaehlt (Clients, die gerade nichts senden, fehlen in der
// Uebersetzungstabelle), faellt der Wert nicht darunter.
// Aufruf mit nodes.Lock().
func (nodes *Nodes) accessPointUpdate(nodeID string, node *Node) {
	if nodes.apRouter == nil {
		nodes.apRouter = make(map[string]string)
	}
	if nodes.IsAccessPoint(node) {
		if router := nodes.routerOf(nodeID, node); router != "" {
			nodes.apRouter[nodeID] = router
		} else {
			delete(nodes.apRouter, nodeID)
		}
		return
	}
	stats := node.Statistics
	if stats == nil {
		return
	}
	var behind uint32
	for apID, router := range nodes.apRouter {
		if router != nodeID {
			continue
		}
		ap := nodes.List[apID]
		if ap == nil {
			delete(nodes.apRouter, apID)
			continue
		}
		if !ap.Online || ap.Statistics == nil {
			continue
		}
		behind += ap.Statistics.Clients.Total
	}
	if behind == 0 {
		return
	}
	c := &stats.Clients
	own := c.Wifi
	if own == 0 {
		own = c.Wifi24 + c.Wifi5
	}
	if c.Total > behind && c.Total-behind > own {
		own = c.Total - behind
	}
	if own < c.Total {
		c.Total = own
	}
}
