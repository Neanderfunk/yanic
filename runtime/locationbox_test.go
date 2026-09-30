package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FreifunkBremen/yanic/data"
)

var deutschland = []float64{47.2, 5.8, 55.1, 15.1}

func TestOrtImRahmen(t *testing.T) {
	assert := assert.New(t)
	ort := func(lat, lon float64) *data.Nodeinfo {
		return &data.Nodeinfo{Location: &data.Location{Latitude: lat, Longitude: lon}}
	}

	ni := ort(51.25, 6.97)
	assert.False(ortImRahmen(deutschland, ni), "im Rahmen bleibt")
	assert.Equal(51.25, ni.Location.Latitude)

	ni = ort(6.97, 51.25)
	assert.True(ortImRahmen(deutschland, ni), "vertauscht wird getauscht")
	assert.Equal(51.25, ni.Location.Latitude)
	assert.Equal(6.97, ni.Location.Longitude)

	ni = ort(39.55, 2.63)
	assert.False(ortImRahmen(deutschland, ni), "beide Lesarten ausserhalb bleibt")
	assert.Equal(39.55, ni.Location.Latitude)

	ni = ort(51.0, 52.0)
	assert.False(ortImRahmen([]float64{50, 50, 53, 53}, ni), "beide Lesarten innerhalb bleibt")

	ni = ort(6.97, 51.25)
	assert.False(ortImRahmen(nil, ni), "ohne Rahmen unveraendert")
	assert.False(ortImRahmen(deutschland, &data.Nodeinfo{}), "ohne Ort unveraendert")
	assert.False(ortImRahmen(deutschland, nil))
}

func TestUpdateTauschtVorDemAlias(t *testing.T) {
	assert := assert.New(t)
	nodes := NewNodes(&NodesConfig{LocationBBox: deutschland})

	node := nodes.Update("e8ab2f35b3af", meldung(t, "e8ab2f35b3af", "vertauscht", 6.97, 51.25))
	assert.Equal(51.25, node.Nodeinfo.Location.Latitude)
	assert.Equal(6.97, node.Nodeinfo.Location.Longitude)

	// zweite Runde mit derselben Meldung: bleibt getauscht
	node = nodes.Update("e8ab2f35b3af", meldung(t, "e8ab2f35b3af", "vertauscht", 6.97, 51.25))
	assert.Equal(51.25, node.Nodeinfo.Location.Latitude)
}
