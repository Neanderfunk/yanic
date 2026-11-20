package siteordomain

import (
	"testing"

	"github.com/FreifunkBremen/yanic/data"
	"github.com/FreifunkBremen/yanic/runtime"
	"github.com/stretchr/testify/assert"
)

func knoten(site, domain string) *runtime.Node {
	return &runtime.Node{Nodeinfo: &data.Nodeinfo{System: data.System{SiteCode: site, DomainCode: domain}}}
}

func TestFilterSiteOrDomain(t *testing.T) {
	assert := assert.New(t)

	_, err := build("21_dias")
	assert.Error(err)
	_, err = build([]interface{}{3, "21_dias"})
	assert.Error(err)

	f, err := build([]interface{}{"nef-21_dias", "nef-21_dias_EOL", "21_dias"})
	assert.NoError(err)

	// alter Knoten, Single-Domain
	assert.NotNil(f.Apply(knoten("nef-21_dias", "")))
	assert.NotNil(f.Apply(knoten("nef-21_dias_EOL", "")))
	// neuer Knoten, Multi-Domain
	assert.NotNil(f.Apply(knoten("ffnef", "21_dias")))
	// Multi-Domain, andere Domain
	assert.Nil(f.Apply(knoten("ffnef", "10_wlf")))
	// Supernode: nur domain_code, auch wenn der in der Liste steht
	assert.Nil(f.Apply(knoten("", "nef-21_dias")))
	assert.Nil(f.Apply(knoten("", "21_dias")))
	// fremde Site
	assert.Nil(f.Apply(knoten("ffxx", "")))
	// ohne Daten
	assert.Nil(f.Apply(&runtime.Node{}))

	// leerer Eintrag in der Liste laesst keine Knoten ohne site_code durch
	f, err = build([]interface{}{""})
	assert.NoError(err)
	assert.Nil(f.Apply(knoten("", "")))
}
