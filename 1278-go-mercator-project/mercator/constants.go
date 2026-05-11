package mercator

import "math"

const (
	EarthRadiusMeters = 6378137.0
	
	MercatorMinX = -20037508.342789244
	MercatorMaxX = 20037508.342789244
	MercatorMinY = -20037508.342789244
	MercatorMaxY = 20037508.342789244
	
	LatMax = 85.05112877980659
	LatMin = -85.05112877980659
	
	MaxZoom = 22
	MinZoom = 0
	
	TileSizePixels = 256
)

var (
	LatMaxRad = LatMax * math.Pi / 180.0
	LatMinRad = LatMin * math.Pi / 180.0
	MercatorEarthCircumference = 2 * math.Pi * EarthRadiusMeters
)
