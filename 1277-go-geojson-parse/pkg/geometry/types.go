package geometry

type Point struct {
	Lon float64
	Lat float64
}

func NewPoint(lon, lat float64) Point {
	return Point{Lon: lon, Lat: lat}
}

type LineString struct {
	Points []Point
}

func NewLineString(points []Point) LineString {
	return LineString{Points: points}
}

type Polygon struct {
	Rings [][]Point
}

func NewPolygon(rings [][]Point) Polygon {
	return Polygon{Rings: rings}
}

type MultiPoint struct {
	Points []Point
}

type MultiLineString struct {
	LineStrings []LineString
}

type MultiPolygon struct {
	Polygons []Polygon
}

type GeometryType string

const (
	GeometryTypePoint              GeometryType = "Point"
	GeometryTypeLineString         GeometryType = "LineString"
	GeometryTypePolygon            GeometryType = "Polygon"
	GeometryTypeMultiPoint         GeometryType = "MultiPoint"
	GeometryTypeMultiLineString    GeometryType = "MultiLineString"
	GeometryTypeMultiPolygon       GeometryType = "MultiPolygon"
)

type Geometry struct {
	Type             GeometryType
	Point            *Point
	LineString       *LineString
	Polygon          *Polygon
	MultiPoint       *MultiPoint
	MultiLineString  *MultiLineString
	MultiPolygon     *MultiPolygon
}

type Feature struct {
	Geometry   Geometry
	Properties map[string]interface{}
	ID         interface{}
}

type FeatureCollection struct {
	Features []Feature
}

type MBR struct {
	MinLon float64
	MinLat float64
	MaxLon float64
	MaxLat float64
}
