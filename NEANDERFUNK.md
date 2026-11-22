# Zweig `neanderfunk`

Dieser Fork von [FreifunkBremen/yanic](https://codeberg.org/FreifunkBremen/yanic)
sammelt die respondd-Daten für die Karte `neander.map.freifunk.space` von
Freifunk Neanderland. Der Zweig `main` folgt unverändert dem Original, unsere
Änderungen liegen im Zweig `neanderfunk`, je Funktion ein Commit, ursprünglich auf dem
Upstream-Stand `v1.9.0` (`91a8825`), seitdem per Merge nachgezogen (zuletzt
`89bd94c`, 04.05.2026). Jeder Commit besteht die Tests für sich,
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
| Frequenz der WLAN-Links | `neighbours.wifi[<mac>].frequency` (Gluon-Patch respondd-wifi-frequency, wie 4830.org) als `source_frequency`/`target_frequency` je Link in meshviewer.json; optional, ohne Angabe fehlt das Feld. `wifi` wird tolerant gelesen: was kein Objekt ist (aeltere Gluon-Versionen schicken Arrays), zaehlt als keine Angabe, statt die Antwort zu verwerfen |
| CPU-steal | `statistics.stat.cpu.steal` (Supernodes, mesh-announce) als `stat.cpu.steal`; fehlt es, wie bei Gluon, wird nichts geschrieben |
| domainhelper | `statistics.neanderfunk.domainhelper` (neanderfunk-domainhelper 0.12) als eigene Messungen `nf_domainhelper_last_switch` (Feld `time`, Tags `from`/`to`/`how`), `_bridge`, `_follow`, `_island`; nur geschrieben, solange der Knoten den Teil meldet. Tolerant gelesen: ein Teil mit falschem Typ faellt einzeln weg |
| Filter site_or_domain | Liste von Codes: durch kommt, wessen site_code darin steht, oder wer einen site_code hat und dessen domain_code darin steht. Alte Single-Domain-Knoten (`nef-10_wlf`) und Multi-Domain-Knoten (`ffnef` + `10_wlf`) landen so in derselben Ortsausgabe, Supernodes (nur domain_code) nicht. Veraendert den Knoten nicht, anders als domain_as_site |

Die drei Commits zu den Accesspoints gehören zu
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

Aktualisieren auf einen neuen Upstream-Stand, per Merge, nicht per Rebase:
auf die Commits dieses Zweigs verweisen veröffentlichte Links (etwa aus
Forenbeiträgen), ein Rebase würde ihre IDs ändern und einen Force-Push
brauchen. Seit 28.09.2026 (Upstream `89bd94c`) wird zusammengeführt.

```bash
git fetch upstream
git merge upstream/main
go test ./...
```

Welche Commits unsere sind, zeigt dann
`git log --no-merges --first-parent v1.9.0..neanderfunk`.
