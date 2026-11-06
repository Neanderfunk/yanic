package influxdb

// Neanderfunk-Erweiterung: schreibt statistics.neanderfunk (Gluon-Paket
// neanderfunk-respondd) in die Datenbank. Lokaler Patch des
// Neanderfunk-Kartenservers, nicht upstream.
//
// Einzelwerte je Knoten landen als Felder in der Messung node und tragen
// damit alle Knoten-Tags (Modell, Firmware, site). Was es je Knoten mehrfach
// gibt, bekommt eine eigene Messung mit dem Namen als Tag:
//
//   node         nf.refault_file, nf.forks        Zaehler seit dem Start
//                nf.zram.ram/data/size            kB
//                nf.ssid_changer.gateway_losses/offline/switches   Zaehler
//   nf_ethernet  carrier (0/1), speed, possible   Tags port, duplex
//   nf_temperature  celsius                       Tag sensor
//   nf_wireless  txpower, channel, mesh (0/1)     Tags radio, ssid, htmode, country
//   node         nf.ddhcpd.running/learning (0/1)
//                nf.ddhcpd.blocks.*, .leases.*    Momentaufnahme
//                nf.ddhcpd.roaming.*, .dhcp.*     Zaehler seit Start des Daemons
//                nf.clients_ipv4.total/ddhcpd     Momentaufnahme
//
// Fehlt ein Wert, wird er nicht geschrieben (nicht als 0).
// system.mem_available doppelt memory.available und bleibt weg.

import (
	"time"

	models "github.com/influxdata/influxdb1-client/models"

	"github.com/FreifunkBremen/yanic/data"
)

const (
	MeasurementNFEthernet    = "nf_ethernet"
	MeasurementNFTemperature = "nf_temperature"
	MeasurementNFWireless    = "nf_wireless"
)

func setze(fields models.Fields, name string, wert *float64) {
	if wert != nil {
		fields[name] = *wert
	}
}

func bit(wert *bool) (int64, bool) {
	if wert == nil {
		return 0, false
	}
	if *wert {
		return 1, true
	}
	return 0, true
}

// nfNodeFields ergaenzt die Felder der Messung node
func nfNodeFields(nf *data.NeanderfunkStatistics, fields models.Fields) {
	if sys := nf.System; sys != nil {
		setze(fields, "nf.refault_file", sys.RefaultFile)
		setze(fields, "nf.forks", sys.Forks)
		if z := sys.Zram; z != nil {
			setze(fields, "nf.zram.ram", z.RAM)
			setze(fields, "nf.zram.data", z.Data)
			setze(fields, "nf.zram.size", z.Size)
		}
	}
	if sc := nf.SSIDChanger; sc != nil {
		setze(fields, "nf.ssid_changer.gateway_losses", sc.GatewayLosses)
		setze(fields, "nf.ssid_changer.offline", sc.Offline)
		setze(fields, "nf.ssid_changer.switches", sc.Switches)
	}
	if d := nf.DDHCPD; d != nil {
		if v, ok := bit(d.Running); ok {
			fields["nf.ddhcpd.running"] = v
		}
		if v, ok := bit(d.Learning); ok {
			fields["nf.ddhcpd.learning"] = v
		}
		// Schluessel wie vom Knoten gemeldet, neue Zaehler kommen ohne
		// Aenderung hier an
		for gruppe, werte := range map[string]map[string]any{
			"blocks": d.Blocks, "leases": d.Leases, "roaming": d.Roaming, "dhcp": d.DHCP} {
			for k, v := range werte {
				if zahl, ok := v.(float64); ok {
					fields["nf.ddhcpd."+gruppe+"."+k] = zahl
				}
			}
		}
	}
	if c := nf.ClientsIPv4; c != nil {
		setze(fields, "nf.clients_ipv4.total", c.Total)
		setze(fields, "nf.clients_ipv4.ddhcpd", c.DDHCPD)
	}
}

// nfPoints schreibt Ports, Sensoren und Radios als eigene Messungen
func (conn *Connection) nfPoints(nf *data.NeanderfunkStatistics, tags models.Tags, t time.Time) {
	for port, e := range nf.Ethernet {
		if e == nil {
			continue
		}
		f := models.Fields{}
		if v, ok := bit(e.Carrier); ok {
			f["carrier"] = v
		}
		setze(f, "speed", e.Speed)
		setze(f, "possible", e.Possible)
		if len(f) == 0 {
			continue
		}
		tg := tags.Clone()
		tg.SetString("port", port)
		if e.Duplex != "" {
			tg.SetString("duplex", e.Duplex)
		}
		conn.addPoint(MeasurementNFEthernet, tg, f, t)
	}

	for sensor, c := range nf.Temperature {
		if c == nil {
			continue
		}
		tg := tags.Clone()
		tg.SetString("sensor", sensor)
		conn.addPoint(MeasurementNFTemperature, tg, models.Fields{"celsius": *c}, t)
	}

	for radio, r := range nf.Wireless {
		if r == nil {
			continue
		}
		f := models.Fields{}
		setze(f, "txpower", r.TxPower)
		setze(f, "channel", r.Channel)
		if v, ok := bit(r.Mesh); ok {
			f["mesh"] = v
		}
		if len(f) == 0 {
			continue
		}
		tg := tags.Clone()
		tg.SetString("radio", radio)
		for k, v := range map[string]string{"ssid": r.SSID, "htmode": r.HTMode, "country": r.Country} {
			if v != "" {
				tg.SetString(k, v)
			}
		}
		conn.addPoint(MeasurementNFWireless, tg, f, t)
	}
}
