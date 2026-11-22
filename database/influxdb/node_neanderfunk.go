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
//                nf.ddhcpd.roaming.*, .dhcp.*, .conflicts.*
//                                                 Zaehler seit Start des Daemons
//                nf.clients_ipv4.total/ddhcpd     Momentaufnahme
//   nf_domainhelper_last_switch  time (Unix-Zeit vom Knoten)  Tags from, to, how
//   nf_domainhelper_bridge  age    Tags state, where, peer_domain, peer_node, note
//   nf_domainhelper_follow  age, wait                       Tags follow_domain, rule
//   nf_domainhelper_island  age
//                (domainhelper: nur solange der Knoten den Teil meldet)
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
	MeasurementNFDHSwitch    = "nf_domainhelper_last_switch"
	MeasurementNFDHBridge    = "nf_domainhelper_bridge"
	MeasurementNFDHFollow    = "nf_domainhelper_follow"
	MeasurementNFDHIsland    = "nf_domainhelper_island"
)

// mitTags haengt nicht-leere Werte als Tags an eine Kopie
func mitTags(tags models.Tags, werte map[string]string) models.Tags {
	tg := tags.Clone()
	for k, v := range werte {
		if v != "" {
			tg.SetString(k, v)
		}
	}
	return tg
}

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
		// Gruppen und Schluessel wie vom Knoten gemeldet
		for gruppe, werte := range d.Gruppen {
			for k, v := range werte {
				fields["nf.ddhcpd."+gruppe+"."+k] = v
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

	if dh := nf.DomainHelper; dh != nil {
		conn.nfDomainHelper(dh, tags, t)
	}
}

// nfDomainHelper schreibt die Teile von statistics.neanderfunk.domainhelper.
// Die Domain des Nachbarn und des follow-Ziels heissen peer_domain und
// follow_domain, damit sie nicht wie die Domain des Knotens aussehen.
func (conn *Connection) nfDomainHelper(dh *data.NeanderfunkDomainHelper, tags models.Tags, t time.Time) {
	if ls := dh.LastSwitch; ls != nil && ls.Time != nil {
		conn.addPoint(MeasurementNFDHSwitch,
			mitTags(tags, map[string]string{"from": ls.From, "to": ls.To, "how": ls.How}),
			models.Fields{"time": *ls.Time}, t)
	}
	if b := dh.Bridge; b != nil && b.Age != nil {
		conn.addPoint(MeasurementNFDHBridge,
			mitTags(tags, map[string]string{"state": b.State, "where": b.Where,
				"peer_domain": b.Domain, "peer_node": b.Node, "note": b.Note}),
			models.Fields{"age": *b.Age}, t)
	}
	if f := dh.Follow; f != nil {
		fl := models.Fields{}
		setze(fl, "age", f.Age)
		setze(fl, "wait", f.Wait)
		if len(fl) > 0 {
			conn.addPoint(MeasurementNFDHFollow,
				mitTags(tags, map[string]string{"follow_domain": f.Domain, "rule": f.Rule}), fl, t)
		}
	}
	if i := dh.Island; i != nil && i.Age != nil {
		conn.addPoint(MeasurementNFDHIsland, tags.Clone(), models.Fields{"age": *i.Age}, t)
	}
}
