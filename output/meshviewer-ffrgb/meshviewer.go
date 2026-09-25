package meshviewerFFRGB

import (
	"fmt"
	"strings"

	"github.com/FreifunkBremen/yanic/lib/jsontime"
	"github.com/FreifunkBremen/yanic/runtime"
)

func transform(nodes *runtime.Nodes) *Meshviewer {

	meshviewer := &Meshviewer{
		Timestamp: jsontime.Now(),
		Nodes:     make([]*Node, 0),
		Links:     make([]*Link, 0),
	}

	links := make(map[string]*Link)

	nodes.RLock()
	defer nodes.RUnlock()

	for _, nodeOrigin := range nodes.List {
		node := NewNode(nodes, nodeOrigin)
		meshviewer.Nodes = append(meshviewer.Nodes, node)

		if !nodeOrigin.Online {
			continue
		}

		// Lokaler Zusatz (Neanderfunk): den Link eines Accesspoints zu seinem
		// Router meldet nur der AP, fuer den Router ist er ein LAN-Geraet.
		// Die fehlende Seite bekaeme TQ 0, und meshviewer faerbte die Linie
		// nach dem Mittel beider Seiten wie eine halbe Verbindung. Der Typ
		// bliebe "unknown", weil die Router-MAC keine Mesh-Schnittstelle ist;
		// es ist aber immer ein Kabel am LAN des Routers.
		accessPoint := nodes.IsAccessPoint(nodeOrigin)

		for _, linkOrigin := range nodes.NodeLinks(nodeOrigin) {
			var key string
			// keep source and target in the same order
			switchSourceTarget := strings.Compare(linkOrigin.SourceAddress, linkOrigin.TargetAddress) > 0
			if switchSourceTarget {
				key = fmt.Sprintf("%s-%s", linkOrigin.SourceAddress, linkOrigin.TargetAddress)
			} else {
				key = fmt.Sprintf("%s-%s", linkOrigin.TargetAddress, linkOrigin.SourceAddress)
			}

			if link := links[key]; link != nil {
				if switchSourceTarget {
					link.TargetTQ = linkOrigin.TQ
				} else {
					link.SourceTQ = linkOrigin.TQ
				}
				continue
			}
			link := &Link{
				Source:        linkOrigin.SourceID,
				SourceAddress: linkOrigin.SourceAddress,
				Target:        linkOrigin.TargetID,
				TargetAddress: linkOrigin.TargetAddress,
				SourceTQ:      linkOrigin.TQ,
				TargetTQ:      0,
				Type:          linkOrigin.Type.String(),
			}

			if switchSourceTarget {
				link.SourceTQ = 0
				link.Source = linkOrigin.TargetID
				link.SourceAddress = linkOrigin.TargetAddress
				link.TargetTQ = linkOrigin.TQ
				link.Target = linkOrigin.SourceID
				link.TargetAddress = linkOrigin.SourceAddress
			}

			if accessPoint {
				link.SourceTQ = linkOrigin.TQ
				link.TargetTQ = linkOrigin.TQ
				link.Type = runtime.OtherLinkType.String()
			}

			links[key] = link
		}
	}
	for _, link := range links {
		meshviewer.Links = append(meshviewer.Links, link)
	}
	return meshviewer
}
