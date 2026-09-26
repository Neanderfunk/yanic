package runtime

import "github.com/FreifunkBremen/yanic/lib/duration"

type NodesConfig struct {
	StatePath    string            `toml:"state_path"`
	SaveInterval duration.Duration `toml:"save_interval"` // Save nodes periodically
	OfflineAfter duration.Duration `toml:"offline_after"` // Set node to offline if not seen within this period
	PruneAfter   duration.Duration `toml:"prune_after"`   // Remove nodes after n days of inactivity
	Output       map[string]interface{}

	// Lokaler Zusatz (Neanderfunk): software.firmware.base von Accesspoints
	// hinter Freifunk-Routern, siehe accesspoints.go. Nicht gesetzt: "UniFi".
	AccessPointFirmware []string `toml:"accesspoint_firmware"`

	// Lokaler Zusatz (Neanderfunk): Aliase wie bei hopglass-server, siehe
	// aliases.go. Leer: keine.
	AliasesPath string `toml:"aliases_path"`

	// Lokaler Zusatz (Neanderfunk): Verzeichnis mit Loeschauftraegen fuer
	// Offline-Knoten, siehe removals.go. Leer: keine.
	RemoveDir string `toml:"remove_dir"`
}
