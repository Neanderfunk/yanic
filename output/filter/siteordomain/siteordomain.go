// Package siteordomain filtert nach site_code oder domain_code (Neanderfunk).
//
// Ein Knoten kommt durch, wenn sein site_code in der Liste steht, oder wenn
// er einen site_code hat und sein domain_code in der Liste steht. So passen
// alte Single-Domain-Knoten (site_code "nef-10_wlf") und neue
// Multi-Domain-Knoten (site_code "ffnef", domain_code "10_wlf") in dieselbe
// Ortsausgabe. Supernodes melden nur einen domain_code und keinen site_code
// und bleiben deshalb draussen.
//
// Anders als domain_as_site veraendert der Filter den Knoten nicht und
// haengt nicht von der Reihenfolge der Filter ab.
package siteordomain

import (
	"errors"

	"github.com/FreifunkBremen/yanic/output/filter"
	"github.com/FreifunkBremen/yanic/runtime"
)

type codes map[string]struct{}

func init() {
	filter.Register("site_or_domain", build)
}

func build(config interface{}) (filter.Filter, error) {
	values, ok := config.([]interface{})
	if !ok {
		return nil, errors.New("invalid configuration, array (of strings) expected")
	}
	list := make(codes)
	for _, value := range values {
		code, ok := value.(string)
		if !ok {
			return nil, errors.New("invalid configuration, array of strings expected")
		}
		list[code] = struct{}{}
	}
	return list, nil
}

func (list codes) Apply(node *runtime.Node) *runtime.Node {
	nodeinfo := node.Nodeinfo
	if nodeinfo == nil {
		return nil
	}
	s := nodeinfo.System
	if _, ok := list[s.SiteCode]; ok && s.SiteCode != "" {
		return node
	}
	if s.SiteCode != "" && s.DomainCode != "" {
		if _, ok := list[s.DomainCode]; ok {
			return node
		}
	}
	return nil
}
