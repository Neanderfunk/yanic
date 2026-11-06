package data

import "encoding/json"

// Neanderfunk-Erweiterung: statistics.neanderfunk aus dem Gluon-Paket
// neanderfunk-respondd (Feed Neanderfunk/packages, ab bd33007, Temperatur ab
// c5d5275). Lokaler Patch des Neanderfunk-Kartenservers, nicht upstream.
//
// Alles ist optional. Alt-Firmware liefert den Baum gar nicht, Knoten ohne
// Sensor keine Temperatur, Knoten ohne zram keine zram-Werte. Deshalb Zeiger:
// ein fehlender Wert bleibt nil und wird nicht als 0 gespeichert, sonst sahe
// ein Knoten ohne Sensor aus wie einer bei null Grad.
//
// Zahlen als float64 aus demselben Grund wie im Rest von yanic: Lua-basierte
// respondd-Module kennen keinen Ganzzahltyp.

// NeanderfunkStatistics ist statistics.neanderfunk
type NeanderfunkStatistics struct {
	System      *NeanderfunkSystem              `json:"system,omitempty"`
	Ethernet    map[string]*NeanderfunkEthernet `json:"ethernet,omitempty"`
	SSIDChanger *NeanderfunkSSIDChanger         `json:"ssid_changer,omitempty"`
	Wireless    map[string]*NeanderfunkRadio    `json:"wireless,omitempty"`
	// Schluessel ist der Sensor (soc, mt7915_phy0, ...), Wert in Grad Celsius
	Temperature map[string]*float64     `json:"temperature,omitempty"`
	DDHCPD      *NeanderfunkDDHCPD      `json:"ddhcpd,omitempty"`
	ClientsIPv4 *NeanderfunkClientsIPv4 `json:"clients_ipv4,omitempty"`
}

// NeanderfunkSystem ist statistics.neanderfunk.system
type NeanderfunkSystem struct {
	// Zaehler seit dem Start: Pagecache-Refaults (workingset_refault_file)
	RefaultFile *float64 `json:"refault_file,omitempty"`
	// doppelt memory.available, nicht gespeichert
	MemAvailable *float64         `json:"mem_available,omitempty"`
	Zram         *NeanderfunkZram `json:"zram,omitempty"`
	// Zaehler seit dem Start
	Forks *float64 `json:"forks,omitempty"`
}

// NeanderfunkZram in kB
type NeanderfunkZram struct {
	RAM  *float64 `json:"ram,omitempty"`
	Data *float64 `json:"data,omitempty"`
	Size *float64 `json:"size,omitempty"`
}

// NeanderfunkEthernet ist ein Port unter statistics.neanderfunk.ethernet
type NeanderfunkEthernet struct {
	Speed *float64 `json:"speed,omitempty"`
	// Hoechste Rate in Mbit/s, die beide Seiten anbieten (advertising und
	// lp_advertising verundet). 0 = nicht ermittelbar, etwa bei Link down.
	// Auf dem Knoten gecacht, aendert sich nur mit speed. Alarmsignal ist
	// possible > speed: ein 2,5G-Port, der nicht hochkam (RTL8221B).
	Possible *float64 `json:"possible,omitempty"`
	Duplex   string   `json:"duplex,omitempty"`
	Carrier  *bool    `json:"carrier,omitempty"`
}

// NeanderfunkSSIDChanger sind Zaehler seit dem Start
type NeanderfunkSSIDChanger struct {
	GatewayLosses *float64 `json:"gateway_losses,omitempty"`
	Offline       *float64 `json:"offline,omitempty"`
	Switches      *float64 `json:"switches,omitempty"`
}

// NeanderfunkRadio ist Konfiguration, kein Messwert
type NeanderfunkRadio struct {
	TxPower *float64 `json:"txpower,omitempty"`
	SSID    string   `json:"ssid,omitempty"`
	Channel *float64 `json:"channel,omitempty"`
	Mesh    *bool    `json:"mesh,omitempty"`
	HTMode  string   `json:"htmode,omitempty"`
	Country string   `json:"country,omitempty"`
}

// NeanderfunkDDHCPD ist statistics.neanderfunk.ddhcpd (verteilter DHCP, ab
// Feed v2025.1.x 1977778). Fehlt ganz ohne das Paket. Antwortet der Daemon
// nicht, kommt nur running=false, ohne learning und ohne die Gruppen.
//
// Gruppen (Stand Feed 7bb9700): blocks und leases sind Momentaufnahmen,
// roaming, dhcp und conflicts Zaehler seit dem Start des Daemons (Reset bei
// dessen Neustart). Die Blocktabelle ist auf allen Knoten einer Domain gleich:
// total = free + ours + claiming + foreign + releasing. In der Lernphase
// fehlen free, claiming, foreign und releasing.
//
// Jedes Objekt neben running/learning gilt als Gruppe, jede Zahl darin als
// Wert: neue Gruppen und Schluessel kommen ohne Aenderung an yanic an. Was
// keine Zahl ist, faellt einzeln weg und laesst nicht das Einlesen der ganzen
// Statistik scheitern, sonst fehlte der Knoten mit allen Werten (vgl.
// neighbours.wifi, Essen 11/2026).
type NeanderfunkDDHCPD struct {
	Running  *bool
	Learning *bool
	Gruppen  map[string]map[string]float64
}

// UnmarshalJSON liest tolerant, siehe NeanderfunkDDHCPD
func (d *NeanderfunkDDHCPD) UnmarshalJSON(b []byte) error {
	var roh map[string]json.RawMessage
	if json.Unmarshal(b, &roh) != nil {
		return nil
	}
	for name, wert := range roh {
		switch name {
		case "running", "learning":
			var x bool
			if json.Unmarshal(wert, &x) != nil {
				continue
			}
			if name == "running" {
				d.Running = &x
			} else {
				d.Learning = &x
			}
		default:
			var gruppe map[string]any
			if json.Unmarshal(wert, &gruppe) != nil {
				continue
			}
			for k, v := range gruppe {
				zahl, ok := v.(float64)
				if !ok {
					continue
				}
				if d.Gruppen == nil {
					d.Gruppen = map[string]map[string]float64{}
				}
				if d.Gruppen[name] == nil {
					d.Gruppen[name] = map[string]float64{}
				}
				d.Gruppen[name][k] = zahl
			}
		}
	}
	return nil
}

// MarshalJSON schreibt die Form vom Knoten zurueck (state-Datei von yanic)
func (d NeanderfunkDDHCPD) MarshalJSON() ([]byte, error) {
	aus := map[string]any{}
	for name, gruppe := range d.Gruppen {
		aus[name] = gruppe
	}
	if d.Running != nil {
		aus["running"] = *d.Running
	}
	if d.Learning != nil {
		aus["learning"] = *d.Learning
	}
	return json.Marshal(aus)
}

// NeanderfunkClientsIPv4 sind die lokalen Clients mit IPv4 laut ARP an
// local-node; ddhcpd = davon aus ddhcpd.range der Domain, der Rest vom Kea
// des Supernodes. Nicht vergleichbar mit clients.total (andere Quelle).
type NeanderfunkClientsIPv4 struct {
	Total  *float64 `json:"total,omitempty"`
	DDHCPD *float64 `json:"ddhcpd,omitempty"`
}
