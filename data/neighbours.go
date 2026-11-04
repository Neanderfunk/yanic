package data

import "encoding/json"

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
	Wifi   WifiInterfaces `json:"wifi,omitempty"`
	NodeID string         `json:"node_id"`
}

// WifiInterfaces: Schluessel ist die MAC der WLAN-Schnittstelle.
type WifiInterfaces map[string]WifiInterface

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

// UnmarshalJSON nimmt wifi tolerant: aeltere Gluon-Versionen (etwa die 0.7.2
// von Freifunk Essen) schicken je Schnittstelle oder fuer wifi im Ganzen ein
// Array. Mit der strengen Struktur verwarf yanic deshalb die ganze
// neighbours-Antwort, und 13 Knoten standen ab 03.11.2026 als offline auf der
// Karte. Was kein Objekt ist, zaehlt jetzt als "keine Angabe".
func (w *WifiInterfaces) UnmarshalJSON(b []byte) error {
	var roh map[string]json.RawMessage
	if json.Unmarshal(b, &roh) != nil {
		*w = nil
		return nil
	}
	m := make(WifiInterfaces, len(roh))
	for mac, wert := range roh {
		var i WifiInterface
		_ = i.UnmarshalJSON(wert)
		m[mac] = i
	}
	*w = m
	return nil
}

// UnmarshalJSON liest nur die Frequenz und uebergeht alles andere.
func (w *WifiInterface) UnmarshalJSON(b []byte) error {
	var v struct {
		Frequency int `json:"frequency"`
	}
	if json.Unmarshal(b, &v) == nil {
		w.Frequency = v.Frequency
	} else {
		w.Frequency = 0
	}
	return nil
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
