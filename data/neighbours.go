package data

// Neighbours struct
type Neighbours struct {
	Batadv map[string]BatadvNeighbours `json:"batadv"`
	Babel  map[string]BabelNeighbours  `json:"babel"`
	LLDP   map[string][]string         `json:"lldp"`
	//WifiNeighbours map[string]WifiNeighbours   `json:"wifi"`
	// Lokaler Zusatz (Neanderfunk): je WLAN-Schnittstelle (Schluessel ist
	// ihre MAC) nur die Frequenz, damit die Karte das Band eines Links zeigen
	// kann. Die Nachbarwerte (Signal, Rauschen) bleiben draussen, sie wuerden
	// nur den gespeicherten Zustand aufblaehen.
	Wifi   map[string]WifiInterface `json:"wifi,omitempty"`
	NodeID string                   `json:"node_id"`
}

// WifiLink struct
type WifiLink struct {
	Inactive int `json:"inactive"`
	Noise    int `json:"nois"`
	Signal   int `json:"signal"`
}

// WifiInterface: Frequenz der WLAN-Schnittstelle in MHz. Optional, Gluon
// meldet sie nur mit dem Patch respondd-wifi-frequency (4830.org,
// Neanderfunk), aeltere Knoten gar nicht.
type WifiInterface struct {
	Frequency int `json:"frequency,omitempty"`
}

// BatmanLink struct
type BatmanLink struct {
	Lastseen float64 `json:"lastseen"`
	TQ       int     `json:"tq"`
}

// BabelLink struct
type BabelLink struct {
	// How need this:
	RXCost       int `json:"rxcost"`
	TXCost       int `json:"txcost"`
	Cost         int `json:"cost"`
	Reachability int `json:"reachability"`
}

// BatadvNeighbours struct
type BatadvNeighbours struct {
	Neighbours map[string]BatmanLink `json:"neighbours"`
}

// BabelNeighbours struct
type BabelNeighbours struct {
	Protocol         string               `json:"protocol"`
	LinkLocalAddress string               `json:"ll-addr"`
	Neighbours       map[string]BabelLink `json:"neighbours"`
}

// WifiNeighbours struct
type WifiNeighbours struct {
	Neighbours map[string]WifiLink `json:"neighbours"`
}
