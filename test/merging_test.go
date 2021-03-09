package test

import (
	"encoding/json"
	"testing"

	"github.com/KodepandaID/panggilhttp/pkg/merging"
	"github.com/stretchr/testify/assert"
)

func TestMergeData(t *testing.T) {
	m := merging.New()

	body1 := []byte(`{"id_hotel": 25, "name": "Hotel California", "available": true, "destination_id": 1034}`)
	body2 := []byte(`{"destination_id": 1034, "destinations": ["LAX", "SFO", "OAK"]}`)

	m.Merge(nil, body1)
	m.Merge(nil, body2)

	finalBody := m.Get()

	type Hotels struct {
		IDHotel       int      `json:"id_hotel"`
		DestinationID int      `json:"destination_id"`
		Name          string   `json:"name"`
		Destinations  []string `json:"destinations"`
	}

	var h Hotels
	json.Unmarshal(finalBody, &h)

	assert.Equal(t, 25, h.IDHotel)
	assert.Equal(t, 1034, h.DestinationID)
	assert.Equal(t, "Hotel California", h.Name)
	assert.Equal(t, 3, len(h.Destinations))
}

func TestMergeDataWithWishlistAndBlacklist(t *testing.T) {
	m := merging.New()

	body1 := []byte(`{"id_hotel": 25, "name": "Hotel California", "total_area": 468.19, "available": true, "destination_id": 1034, "unquote": "\"abc\"", "nil": null, "string": ["A", "B"], "int": [1, 2], "bool": [true, false], "coordinate": [40.741895, -73.989308], "empty": [], "object": {"bool": true, "empty": [], "phi": 3.14}}`)
	body2 := []byte(`{"destination_id": 1034, "destinations": ["LAX", "SFO", "OAK"]}`)

	m.MergeFromWhitelist([]string{"id_hotel", "destination_id", "available", "total_area", "unquote", "nil", "string", "int", "bool", "coordinate", "empty", "object"}, body1)
	m.Merge([]string{"destination_id"}, body2)

	finalBody := m.Get()

	type Hotels struct {
		IDHotel       int       `json:"id_hotel"`
		DestinationID int       `json:"destination_id"`
		Name          string    `json:"name"`
		Available     bool      `json:"available"`
		Coordinate    []float64 `json:"coordinate"`
		Destinations  []string  `json:"destinations"`
	}

	var h Hotels
	json.Unmarshal(finalBody, &h)

	assert.Equal(t, 25, h.IDHotel)
	assert.Equal(t, 1034, h.DestinationID)
	assert.Equal(t, 3, len(h.Destinations))
}
