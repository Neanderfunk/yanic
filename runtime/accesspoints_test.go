package runtime

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FreifunkBremen/yanic/data"
)

func apNodeinfo(t *testing.T, id, mac, base string) *data.Nodeinfo {
	var ni data.Nodeinfo
	raw := `{"node_id":"` + id + `","network":{"mac":"` + mac + `"},` +
		`"software":{"firmware":{"base":"` + base + `"}}}`
	if err := json.Unmarshal([]byte(raw), &ni); err != nil {
		t.Fatal(err)
	}
	return &ni
}

func routerResponse(t *testing.T, total, wifi24, wifi5 uint32) *data.ResponseData {
	return &data.ResponseData{
		Nodeinfo: apNodeinfo(t, "router", "aa:aa:aa:aa:aa:01", "gluon-v2023.2.5"),
		Statistics: &data.Statistics{Clients: data.Clients{
			Total: total, Wifi: wifi24 + wifi5, Wifi24: wifi24, Wifi5: wifi5,
		}},
	}
}

func apResponse(t *testing.T, id, mac, base, router string, clients uint32) *data.ResponseData {
	return &data.ResponseData{
		Nodeinfo: apNodeinfo(t, id, mac, base),
		Neighbours: &data.Neighbours{
			NodeID: id,
			Batadv: map[string]data.BatadvNeighbours{
				mac: {Neighbours: map[string]data.BatmanLink{router: {TQ: 255}}},
			},
		},
		Statistics: &data.Statistics{Clients: data.Clients{Total: clients, Wifi: clients}},
	}
}

func TestAccessPointClientsAbgezogen(t *testing.T) {
	assert := assert.New(t)
	nodes := NewNodes(&NodesConfig{})

	nodes.Update("router", routerResponse(t, 20, 2, 1))
	nodes.Update("ap1", apResponse(t, "ap1", "0c:ea:14:00:00:01", "UniFi", "aa:aa:aa:aa:aa:01", 9))
	nodes.Update("ap2", apResponse(t, "ap2", "0c:ea:14:00:00:02", "UniFi", "aa:aa:aa:aa:aa:01", 4))
	assert.Equal("router", nodes.apRouter["ap1"])

	// Beim naechsten Bericht des Routers sind die AP-Clients abgezogen
	nodes.Update("router", routerResponse(t, 20, 2, 1))
	assert.Equal(uint32(7), nodes.List["router"].Statistics.Clients.Total)
	// Der AP behaelt seine eigenen
	assert.Equal(uint32(9), nodes.List["ap1"].Statistics.Clients.Total)
}

func TestAccessPointClientsNieUnterEigenesWLAN(t *testing.T) {
	assert := assert.New(t)
	nodes := NewNodes(&NodesConfig{})

	nodes.Update("router", routerResponse(t, 10, 2, 1))
	nodes.Update("ap1", apResponse(t, "ap1", "0c:ea:14:00:00:01", "UniFi", "aa:aa:aa:aa:aa:01", 25))
	nodes.Update("router", routerResponse(t, 10, 2, 1))
	assert.Equal(uint32(3), nodes.List["router"].Statistics.Clients.Total)
}

func TestAccessPointOfflineZaehltNicht(t *testing.T) {
	assert := assert.New(t)
	nodes := NewNodes(&NodesConfig{})

	nodes.Update("router", routerResponse(t, 20, 0, 0))
	nodes.Update("ap1", apResponse(t, "ap1", "0c:ea:14:00:00:01", "UniFi", "aa:aa:aa:aa:aa:01", 9))
	nodes.List["ap1"].Online = false
	nodes.Update("router", routerResponse(t, 20, 0, 0))
	assert.Equal(uint32(20), nodes.List["router"].Statistics.Clients.Total)
}

func TestAccessPointNurBekannteFirmware(t *testing.T) {
	assert := assert.New(t)
	nodes := NewNodes(&NodesConfig{})

	// Ein Gluon-Knoten als Nachbar ist ein gewoehnlicher Mesh-Nachbar
	nodes.Update("router", routerResponse(t, 20, 0, 0))
	nodes.Update("mesh", apResponse(t, "mesh", "02:00:00:00:00:09", "gluon-v2023.2.5", "aa:aa:aa:aa:aa:01", 9))
	nodes.Update("router", routerResponse(t, 20, 0, 0))
	assert.Equal(uint32(20), nodes.List["router"].Statistics.Clients.Total)
	assert.False(nodes.IsAccessPoint(nodes.List["mesh"]))
}

func TestAccessPointAbschaltbar(t *testing.T) {
	assert := assert.New(t)
	nodes := NewNodes(&NodesConfig{AccessPointFirmware: []string{}})

	nodes.Update("router", routerResponse(t, 20, 0, 0))
	nodes.Update("ap1", apResponse(t, "ap1", "0c:ea:14:00:00:01", "UniFi", "aa:aa:aa:aa:aa:01", 9))
	nodes.Update("router", routerResponse(t, 20, 0, 0))
	assert.Equal(uint32(20), nodes.List["router"].Statistics.Clients.Total)
}

func TestAccessPointUmgezogen(t *testing.T) {
	assert := assert.New(t)
	nodes := NewNodes(&NodesConfig{})

	nodes.Update("router", routerResponse(t, 20, 0, 0))
	r2 := routerResponse(t, 15, 0, 0)
	r2.Nodeinfo = apNodeinfo(t, "router2", "aa:aa:aa:aa:aa:02", "gluon-v2023.2.5")
	nodes.Update("router2", r2)
	nodes.Update("ap1", apResponse(t, "ap1", "0c:ea:14:00:00:01", "UniFi", "aa:aa:aa:aa:aa:01", 9))
	nodes.Update("ap1", apResponse(t, "ap1", "0c:ea:14:00:00:01", "UniFi", "aa:aa:aa:aa:aa:02", 9))

	nodes.Update("router", routerResponse(t, 20, 0, 0))
	assert.Equal(uint32(20), nodes.List["router"].Statistics.Clients.Total)
	r2 = routerResponse(t, 15, 0, 0)
	r2.Nodeinfo = apNodeinfo(t, "router2", "aa:aa:aa:aa:aa:02", "gluon-v2023.2.5")
	nodes.Update("router2", r2)
	assert.Equal(uint32(6), nodes.List["router2"].Statistics.Clients.Total)
}

func TestAccessPointGesamtstatistikZaehltEinmal(t *testing.T) {
	assert := assert.New(t)
	nodes := NewNodes(&NodesConfig{})

	nodes.Update("router", routerResponse(t, 20, 2, 1))
	nodes.Update("ap1", apResponse(t, "ap1", "0c:ea:14:00:00:01", "UniFi", "aa:aa:aa:aa:aa:01", 9))
	nodes.Update("router", routerResponse(t, 20, 2, 1))
	stats := NewGlobalStats(nodes, map[string][]string{})
	assert.Equal(uint32(20), stats[GLOBAL_SITE][GLOBAL_DOMAIN].Clients)
}

// Meldet ein AP neben dem Router einen anderen AP als Nachbarn (UniFi-Mesh),
// ist sein Router trotzdem der Freifunk-Router.
func TestAccessPointRouterNichtDerNachbarAP(t *testing.T) {
	assert := assert.New(t)
	nodes := NewNodes(&NodesConfig{})

	nodes.Update("router", routerResponse(t, 20, 0, 0))
	nodes.Update("ap2", apResponse(t, "ap2", "0c:ea:14:00:00:02", "UniFi", "aa:aa:aa:aa:aa:01", 4))
	ap1 := apResponse(t, "ap1", "0c:ea:14:00:00:01", "UniFi", "aa:aa:aa:aa:aa:01", 9)
	ap1.Neighbours.Batadv["0c:ea:14:00:00:01"].Neighbours["0c:ea:14:00:00:02"] = data.BatmanLink{TQ: 255}
	nodes.Update("ap1", ap1)
	assert.Equal("router", nodes.apRouter["ap1"])

	nodes.Update("router", routerResponse(t, 20, 0, 0))
	assert.Equal(uint32(7), nodes.List["router"].Statistics.Clients.Total)
}
