# Zweig `neanderfunk`

Dieser Fork von [FreifunkBremen/yanic](https://codeberg.org/FreifunkBremen/yanic)
sammelt die respondd-Daten für die Karte `neander.map.freifunk.space` von
Freifunk Neanderland. Der Zweig `main` folgt unverändert dem Original, unsere
Änderungen liegen im Zweig `neanderfunk`, je Funktion ein Commit auf dem
Upstream-Stand `v1.9.0` (`91a8825`). Jeder Commit besteht die Tests für sich,
sie lassen sich also einzeln übernehmen oder bei einem Upstream-Update einzeln
nachziehen.

| Commit | Was |
| --- | --- |
| Zusätzliche Zieladressen (seeds) | fragt Adressen aus einer Datei per Unicast, für Netze ohne Rundruf |
| statistics.neanderfunk | Werte des Gluon-Pakets neanderfunk-respondd in die Zeitreihen |
| Clients der Accesspoints beim Router abziehen | Clients hinter einem AP zählen einmal, nicht beim AP und beim Router |
| Links der Accesspoints von beiden Seiten | Linie zwischen AP und Router mit voller Qualität, als Kabel |
| Links zwischen zwei Accesspoints als Funk | AP-AP (UniFi-Mesh) bleibt Funk, nur AP-Router wird Kabel |
| Aliase wie bei hopglass-server | `nodes.aliases_path`: Ort, Name usw. je node_id überschreiben, `"location": null` nimmt von der Landkarte; rücknehmbar |
| Offline-Knoten entfernen | `nodes.remove_dir`: Löschaufträge aus dem Service-Menü, nur für Knoten, die offline sind |

Die beiden letzten gehören zu
[Neanderfunk/unifi_respondd](https://github.com/Neanderfunk/unifi_respondd):
Accesspoints, die ein Stellvertreter meldet, hängen als LAN-Geräte hinter
einem Freifunk-Router. Erkannt werden sie an `software.firmware.base`:

```toml
[nodes]
# Vorgabe, wenn nicht gesetzt: ["UniFi"]; [] schaltet beides ab
accesspoint_firmware = ["UniFi"]
```

Alle neuen Konfigurationsschlüssel sind optional. Ohne Accesspoints im Netz
verhält sich der Zweig wie das Original.

Lizenz wie das Original: AGPL-3.0.

Aktualisieren auf einen neuen Upstream-Stand:

```bash
git fetch upstream
git rebase <neuer-upstream-tag> neanderfunk
go test ./...
```
