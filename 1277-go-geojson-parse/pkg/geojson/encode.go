package geojson

import (
	"encoding/json"

	"github.com/spatial-index/pkg/geometry"
)

type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	Geometry   GeoJSONGeometry        `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
	ID         interface{}            `json:"id,omitempty"`
}

type GeoJSONGeometry struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
}

func pointToCoords(p geometry.Point) []float64 {
	return []float64{p.Lon, p.Lat}
}

func pointsToCoords(points []geometry.Point) [][]float64 {
	coords := make([][]float64, len(points))
	for i, p := range points {
		coords[i] = pointToCoords(p)
	}
	return coords
}

func ringToCoords(ring []geometry.Point) [][]float64 {
	return pointsToCoords(ring)
}

func polygonToCoords(poly geometry.Polygon) [][][]float64 {
	coords := make([][][]float64, len(poly.Rings))
	for i, ring := range poly.Rings {
		coords[i] = ringToCoords(ring)
	}
	return coords
}

func GeometryToGeoJSON(g geometry.Geometry) GeoJSONGeometry {
	geom := GeoJSONGeometry{Type: string(g.Type)}

	switch g.Type {
	case geometry.GeometryTypePoint:
		if g.Point != nil {
			geom.Coordinates = pointToCoords(*g.Point)
		} else {
			geom.Coordinates = []float64{}
		}
	case geometry.GeometryTypeLineString:
		if g.LineString != nil {
			geom.Coordinates = pointsToCoords(g.LineString.Points)
		} else {
			geom.Coordinates = [][]float64{}
		}
	case geometry.GeometryTypePolygon:
		if g.Polygon != nil {
			geom.Coordinates = polygonToCoords(*g.Polygon)
		} else {
			geom.Coordinates = [][][]float64{}
		}
	case geometry.GeometryTypeMultiPoint:
		if g.MultiPoint != nil {
			geom.Coordinates = pointsToCoords(g.MultiPoint.Points)
		} else {
			geom.Coordinates = [][]float64{}
		}
	case geometry.GeometryTypeMultiLineString:
		if g.MultiLineString != nil {
			coords := make([][][]float64, len(g.MultiLineString.LineStrings))
			for i, ls := range g.MultiLineString.LineStrings {
				coords[i] = pointsToCoords(ls.Points)
			}
			geom.Coordinates = coords
		} else {
			geom.Coordinates = [][][]float64{}
		}
	case geometry.GeometryTypeMultiPolygon:
		if g.MultiPolygon != nil {
			coords := make([][][][]float64, len(g.MultiPolygon.Polygons))
			for i, poly := range g.MultiPolygon.Polygons {
				coords[i] = polygonToCoords(poly)
			}
			geom.Coordinates = coords
		} else {
			geom.Coordinates = [][][][]float64{}
		}
	}

	return geom
}

func FeatureToGeoJSON(f geometry.Feature) GeoJSONFeature {
	return GeoJSONFeature{
		Type:       "Feature",
		Geometry:   GeometryToGeoJSON(f.Geometry),
		Properties: f.Properties,
		ID:         f.ID,
	}
}

func FeatureCollectionToGeoJSON(fc geometry.FeatureCollection) map[string]interface{} {
	features := make([]GeoJSONFeature, len(fc.Features))
	for i, f := range fc.Features {
		features[i] = FeatureToGeoJSON(f)
	}

	return map[string]interface{}{
		"type":     "FeatureCollection",
		"features": features,
	}
}

func MarshalFeature(f geometry.Feature) ([]byte, error) {
	return json.Marshal(FeatureToGeoJSON(f))
}

func MarshalFeatureCollection(fc geometry.FeatureCollection) ([]byte, error) {
	return json.Marshal(FeatureCollectionToGeoJSON(fc))
}
