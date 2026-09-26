package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func auftrag(t *testing.T, dir, name string) {
	if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRemovalNurOffline(t *testing.T) {
	assert := assert.New(t)
	dir := t.TempDir()
	nodes := NewNodes(&NodesConfig{RemoveDir: dir})
	nodes.Update("weg", meldung(t, "weg", "weg", 51.0, 7.0))
	nodes.Update("da", meldung(t, "da", "da", 51.0, 7.0))
	nodes.List["weg"].Online = false

	auftrag(t, dir, "weg")
	auftrag(t, dir, "da")
	auftrag(t, dir, "gibtsnicht")
	nodes.Lock()
	n := nodes.processRemovals()
	nodes.Unlock()

	assert.Equal(1, n)
	assert.Nil(nodes.List["weg"])
	assert.NotNil(nodes.List["da"])
	rest, _ := os.ReadDir(dir)
	assert.Len(rest, 0, "alle Auftraege abgearbeitet")
}

func TestRemovalUngueltigeNamenBleibenLiegen(t *testing.T) {
	assert := assert.New(t)
	dir := t.TempDir()
	nodes := NewNodes(&NodesConfig{RemoveDir: dir})
	auftrag(t, dir, "mit.punkt")
	nodes.Lock()
	assert.Equal(0, nodes.processRemovals())
	nodes.Unlock()
	rest, _ := os.ReadDir(dir)
	assert.Len(rest, 1)
}

func TestRemovalOhneKonfiguration(t *testing.T) {
	nodes := NewNodes(&NodesConfig{})
	nodes.Lock()
	assert.Equal(t, 0, nodes.processRemovals())
	nodes.Unlock()
}

func TestRemovalGemeldetKommtNeu(t *testing.T) {
	assert := assert.New(t)
	dir := t.TempDir()
	nodes := NewNodes(&NodesConfig{RemoveDir: dir})
	nodes.Update("weg", meldung(t, "weg", "weg", 51.0, 7.0))
	nodes.List["weg"].Online = false
	auftrag(t, dir, "weg")
	nodes.Lock()
	nodes.processRemovals()
	nodes.Unlock()
	nodes.Update("weg", meldung(t, "weg", "weg", 51.0, 7.0))
	assert.NotNil(nodes.List["weg"])
	assert.True(nodes.List["weg"].Online)
}
