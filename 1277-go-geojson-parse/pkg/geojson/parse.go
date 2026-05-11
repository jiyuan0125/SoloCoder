package geojson

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spatial-index/pkg/geometry"
)

var (
	ErrInvalidType         = errors.New("invalid GeoJSON type")
	ErrInvalidCoordinates  = errors.New("invalid coordinates")
	ErrMissingType         = errors.New("missing 'type' field")
	ErrMissingCoordinates  = errors.New("missing 'coordinates' field")
	ErrMissingGeometry     = errors.New("missing 'geometry' field")
	ErrPointTooFewCoords   = errors.New("Point must have at least 2 coordinates")
	ErrLineStringTooFew    = errors.New("LineString must have at least 2 points")
	ErrPolygonRingTooFew   = errors.New("Polygon ring must have at least 4 points")
	ErrPolygonNotClosed    = errors.New("Polygon ring is not closed")
)

func Parse(data []byte) (*geometry.FeatureCollection, error) {
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	typeVal, ok := obj["type"]
	if !ok {
		return nil, ErrMissingType
	}

	typeStr, ok := typeVal.(string)
	if !ok {
		return nil, ErrInvalidType
	}

	switch typeStr {
	case "FeatureCollection":
		return parseFeatureCollection(obj)
	case "Feature":
		feature, err := parseFeature(obj)
		if err != nil {
			return nil, err
		}
		return &geometry.FeatureCollection{Features: []geometry.Feature{*feature}}, nil
	case "Point", "LineString", "Polygon", "MultiPoint", "MultiLineString", "MultiPolygon":
		geom, err := parseGeometry(obj)
		if err != nil {
			return nil, err
		}
		feature := geometry.Feature{
			Geometry:   *geom,
			Properties: make(map[string]interface{}),
		}
		return &geometry.FeatureCollection{Features: []geometry.Feature{feature}}, nil
	default:
		return nil, fmt.Errorf("unsupported GeoJSON type: %s", typeStr)
	}
}

func parseFeatureCollection(obj map[string]interface{}) (*geometry.FeatureCollection, error) {
	featuresVal, ok := obj["features"]
	if !ok {
		return &geometry.FeatureCollection{Features: []geometry.Feature{}}, nil
	}

	featuresArr, ok := featuresVal.([]interface{})
	if !ok {
		return nil, errors.New("'features' must be an array")
	}

	fc := &geometry.FeatureCollection{
		Features: make([]geometry.Feature, 0, len(featuresArr)),
	}

	for i, featVal := range featuresArr {
		featObj, ok := featVal.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("feature %d is not an object", i)
		}

		feature, err := parseFeature(featObj)
		if err != nil {
			return nil, fmt.Errorf("feature %d: %w", i, err)
		}

		fc.Features = append(fc.Features, *feature)
	}

	return fc, nil
}

func parseFeature(obj map[string]interface{}) (*geometry.Feature, error) {
	geomVal, ok := obj["geometry"]
	if !ok {
		return nil, ErrMissingGeometry
	}

	var geom *geometry.Geometry
	var err error

	if geomVal == nil {
		geom = &geometry.Geometry{}
	} else {
		geomObj, ok := geomVal.(map[string]interface{})
		if !ok {
			return nil, errors.New("'geometry' must be an object")
		}
		geom, err = parseGeometry(geomObj)
		if err != nil {
			return nil, err
		}
	}

	feature := &geometry.Feature{
		Geometry:   *geom,
		Properties: make(map[string]interface{}),
	}

	if propsVal, ok := obj["properties"]; ok && propsVal != nil {
		if props, ok := propsVal.(map[string]interface{}); ok {
			feature.Properties = props
		}
	}

	if idVal, ok := obj["id"]; ok {
		feature.ID = idVal
	}

	return feature, nil
}

func parseGeometry(obj map[string]interface{}) (*geometry.Geometry, error) {
	typeVal, ok := obj["type"]
	if !ok {
		return nil, ErrMissingType
	}

	typeStr, ok := typeVal.(string)
	if !ok {
		return nil, ErrInvalidType
	}

	geom := &geometry.Geometry{}

	switch typeStr {
	case "Point":
		geom.Type = geometry.GeometryTypePoint
		coords, err := parsePointCoordinates(obj)
		if err != nil {
			return nil, err
		}
		pt := geometry.NewPoint(coords[0], coords[1])
		geom.Point = &pt
	case "LineString":
		geom.Type = geometry.GeometryTypeLineString
		points, err := parseLineStringCoordinates(obj)
		if err != nil {
			return nil, err
		}
		ls := geometry.NewLineString(points)
		geom.LineString = &ls
	case "Polygon":
		geom.Type = geometry.GeometryTypePolygon
		poly, err := parsePolygonCoordinates(obj)
		if err != nil {
			return nil, err
		}
		geometry.FixPolygonRingOrientation(&poly)
		geom.Polygon = &poly
	case "MultiPoint":
		geom.Type = geometry.GeometryTypeMultiPoint
		points, err := parseMultiPointCoordinates(obj)
		if err != nil {
			return nil, err
		}
		geom.MultiPoint = &geometry.MultiPoint{Points: points}
	case "MultiLineString":
		geom.Type = geometry.GeometryTypeMultiLineString
		lsArr, err := parseMultiLineStringCoordinates(obj)
		if err != nil {
			return nil, err
		}
		geom.MultiLineString = &geometry.MultiLineString{LineStrings: lsArr}
	case "MultiPolygon":
		geom.Type = geometry.GeometryTypeMultiPolygon
		polys, err := parseMultiPolygonCoordinates(obj)
		if err != nil {
			return nil, err
		}
		for i := range polys {
			geometry.FixPolygonRingOrientation(&polys[i])
		}
		geom.MultiPolygon = &geometry.MultiPolygon{Polygons: polys}
	default:
		return nil, fmt.Errorf("unsupported geometry type: %s", typeStr)
	}

	return geom, nil
}

func parsePointCoordinates(obj map[string]interface{}) ([]float64, error) {
	coordsVal, ok := obj["coordinates"]
	if !ok {
		return nil, ErrMissingCoordinates
	}

	coords, ok := coordsVal.([]interface{})
	if !ok {
		return nil, ErrInvalidCoordinates
	}

	if len(coords) < 2 {
		return nil, ErrPointTooFewCoords
	}

	result := make([]float64, len(coords))
	for i, c := range coords {
		f, ok := c.(float64)
		if !ok {
			return nil, fmt.Errorf("coordinate %d is not a number", i)
		}
		result[i] = f
	}

	return result, nil
}

func parseLineStringCoordinates(obj map[string]interface{}) ([]geometry.Point, error) {
	coordsVal, ok := obj["coordinates"]
	if !ok {
		return nil, ErrMissingCoordinates
	}

	coordsArr, ok := coordsVal.([]interface{})
	if !ok {
		return nil, ErrInvalidCoordinates
	}

	if len(coordsArr) < 2 {
		return nil, ErrLineStringTooFew
	}

	points := make([]geometry.Point, 0, len(coordsArr))
	for i, coordVal := range coordsArr {
		coord, ok := coordVal.([]interface{})
		if !ok {
			return nil, fmt.Errorf("point %d is not an array", i)
		}
		if len(coord) < 2 {
			return nil, fmt.Errorf("point %d: %w", i, ErrPointTooFewCoords)
		}

		lon, ok := coord[0].(float64)
		if !ok {
			return nil, fmt.Errorf("point %d longitude is not a number", i)
		}
		lat, ok := coord[1].(float64)
		if !ok {
			return nil, fmt.Errorf("point %d latitude is not a number", i)
		}

		points = append(points, geometry.NewPoint(lon, lat))
	}

	return points, nil
}

func parsePolygonCoordinates(obj map[string]interface{}) (geometry.Polygon, error) {
	coordsVal, ok := obj["coordinates"]
	if !ok {
		return geometry.Polygon{}, ErrMissingCoordinates
	}

	ringsArr, ok := coordsVal.([]interface{})
	if !ok {
		return geometry.Polygon{}, ErrInvalidCoordinates
	}

	poly := geometry.Polygon{
		Rings: make([][]geometry.Point, 0, len(ringsArr)),
	}

	for i, ringVal := range ringsArr {
		ringArr, ok := ringVal.([]interface{})
		if !ok {
			return geometry.Polygon{}, fmt.Errorf("ring %d is not an array", i)
		}

		ring, err := parseRing(ringArr, i)
		if err != nil {
			return geometry.Polygon{}, err
		}

		poly.Rings = append(poly.Rings, ring)
	}

	return poly, nil
}

func parseRing(ringArr []interface{}, ringIdx int) ([]geometry.Point, error) {
	if len(ringArr) < 4 {
		return nil, fmt.Errorf("ring %d: %w", ringIdx, ErrPolygonRingTooFew)
	}

	ring := make([]geometry.Point, 0, len(ringArr))
	for i, coordVal := range ringArr {
		coord, ok := coordVal.([]interface{})
		if !ok {
			return nil, fmt.Errorf("ring %d point %d is not an array", ringIdx, i)
		}
		if len(coord) < 2 {
			return nil, fmt.Errorf("ring %d point %d: %w", ringIdx, i, ErrPointTooFewCoords)
		}

		lon, ok := coord[0].(float64)
		if !ok {
			return nil, fmt.Errorf("ring %d point %d longitude is not a number", ringIdx, i)
		}
		lat, ok := coord[1].(float64)
		if !ok {
			return nil, fmt.Errorf("ring %d point %d latitude is not a number", ringIdx, i)
		}

		ring = append(ring, geometry.NewPoint(lon, lat))
	}

	return geometry.EnsureRingClosed(ring), nil
}

func parseMultiPointCoordinates(obj map[string]interface{}) ([]geometry.Point, error) {
	coordsVal, ok := obj["coordinates"]
	if !ok {
		return nil, ErrMissingCoordinates
	}

	coordsArr, ok := coordsVal.([]interface{})
	if !ok {
		return nil, ErrInvalidCoordinates
	}

	points := make([]geometry.Point, 0, len(coordsArr))
	for i, coordVal := range coordsArr {
		coord, ok := coordVal.([]interface{})
		if !ok {
			return nil, fmt.Errorf("point %d is not an array", i)
		}
		if len(coord) < 2 {
			return nil, fmt.Errorf("point %d: %w", i, ErrPointTooFewCoords)
		}

		lon, ok := coord[0].(float64)
		if !ok {
			return nil, fmt.Errorf("point %d longitude is not a number", i)
		}
		lat, ok := coord[1].(float64)
		if !ok {
			return nil, fmt.Errorf("point %d latitude is not a number", i)
		}

		points = append(points, geometry.NewPoint(lon, lat))
	}

	return points, nil
}

func parseMultiLineStringCoordinates(obj map[string]interface{}) ([]geometry.LineString, error) {
	coordsVal, ok := obj["coordinates"]
	if !ok {
		return nil, ErrMissingCoordinates
	}

	lsArr, ok := coordsVal.([]interface{})
	if !ok {
		return nil, ErrInvalidCoordinates
	}

	lineStrings := make([]geometry.LineString, 0, len(lsArr))
	for i, lsVal := range lsArr {
		lsCoords, ok := lsVal.([]interface{})
		if !ok {
			return nil, fmt.Errorf("linestring %d is not an array", i)
		}

		if len(lsCoords) < 2 {
			return nil, fmt.Errorf("linestring %d: %w", i, ErrLineStringTooFew)
		}

		points := make([]geometry.Point, 0, len(lsCoords))
		for j, coordVal := range lsCoords {
			coord, ok := coordVal.([]interface{})
			if !ok {
				return nil, fmt.Errorf("linestring %d point %d is not an array", i, j)
			}
			if len(coord) < 2 {
				return nil, fmt.Errorf("linestring %d point %d: %w", i, j, ErrPointTooFewCoords)
			}

			lon, ok := coord[0].(float64)
			if !ok {
				return nil, fmt.Errorf("linestring %d point %d longitude is not a number", i, j)
			}
			lat, ok := coord[1].(float64)
			if !ok {
				return nil, fmt.Errorf("linestring %d point %d latitude is not a number", i, j)
			}

			points = append(points, geometry.NewPoint(lon, lat))
		}

		lineStrings = append(lineStrings, geometry.NewLineString(points))
	}

	return lineStrings, nil
}

func parseMultiPolygonCoordinates(obj map[string]interface{}) ([]geometry.Polygon, error) {
	coordsVal, ok := obj["coordinates"]
	if !ok {
		return nil, ErrMissingCoordinates
	}

	polysArr, ok := coordsVal.([]interface{})
	if !ok {
		return nil, ErrInvalidCoordinates
	}

	polygons := make([]geometry.Polygon, 0, len(polysArr))
	for i, polyVal := range polysArr {
		ringsArr, ok := polyVal.([]interface{})
		if !ok {
			return nil, fmt.Errorf("polygon %d is not an array", i)
		}

		poly := geometry.Polygon{
			Rings: make([][]geometry.Point, 0, len(ringsArr)),
		}

		for j, ringVal := range ringsArr {
			ringArr, ok := ringVal.([]interface{})
			if !ok {
				return nil, fmt.Errorf("polygon %d ring %d is not an array", i, j)
			}

			ring, err := parseRing(ringArr, j)
			if err != nil {
				return nil, fmt.Errorf("polygon %d: %w", i, err)
			}

			poly.Rings = append(poly.Rings, ring)
		}

		polygons = append(polygons, poly)
	}

	return polygons, nil
}
