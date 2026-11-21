package influxdb

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FreifunkBremen/yanic/data"
	"github.com/FreifunkBremen/yanic/runtime"
)

// stat.cpu.steal nur, wenn der Knoten es meldet (Supernodes, nicht Gluon)
func TestStealNurWennGemeldet(t *testing.T) {
	assert := assert.New(t)
	steal := int64(42694)
	knoten := func(id string, s *int64) *runtime.Node {
		return &runtime.Node{
			Online: true,
			Statistics: &data.Statistics{NodeID: id, ProcStats: &data.ProcStats{
				CPU: data.ProcStatsCPU{User: 1, Steal: s}}},
			Nodeinfo: &data.Nodeinfo{NodeID: id},
		}
	}
	points := testPoints(knoten("f2beef002107", &steal), knoten("deadbeef0001", nil))
	assert.Len(points, 2)
	gesehen := 0
	for _, p := range points {
		fields, _ := p.Fields()
		switch p.Tags()["nodeid"] {
		case "f2beef002107":
			assert.EqualValues(42694, fields["stat.cpu.steal"])
			gesehen++
		case "deadbeef0001":
			_, da := fields["stat.cpu.steal"]
			assert.False(da)
			assert.EqualValues(1, fields["stat.cpu.user"])
			gesehen++
		}
	}
	assert.Equal(2, gesehen)
}
