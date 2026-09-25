package data

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
	Temperature map[string]*float64 `json:"temperature,omitempty"`
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
