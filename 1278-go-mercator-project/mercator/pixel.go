package mercator

import (
	"math"
)

type PixelResult struct {
	X       int
	Y       int
	Warning string
}

type ViewBoundsResult struct {
	NW      LatLng
	NE      LatLng
	SW      LatLng
	SE      LatLng
	Warning string
}

type LatLng struct {
	Lat float64
	Lng float64
}

type MapView struct {
	Center LatLng
	Zoom   int
	Width  int
	Height int
}

type Pixel struct {
	X int
	Y int
}

func LatLngToPixel(centerLat, centerLng float64, zoom, width, height int, targetLat, targetLng float64) PixelResult {
	var warning string
	
	if zoom < MinZoom || zoom > MaxZoom {
		zoom = clampZoom(zoom)
		warning = "zoom level out of range, clamped"
	}
	
	centerLat = clampLat(centerLat)
	centerLng = normalizeLng(centerLng)
	targetLat = clampLat(targetLat)
	targetLng = normalizeLng(targetLng)
	
	mercCenter := LatLngToMercator(centerLat, centerLng)
	mercTarget := LatLngToMercator(targetLat, targetLng)
	
	scale := math.Pow(2.0, float64(zoom))
	pixelsPerMeter := scale * float64(TileSizePixels) / MercatorEarthCircumference
	
	deltaX := (mercTarget.X - mercCenter.X) * pixelsPerMeter
	deltaY := (mercCenter.Y - mercTarget.Y) * pixelsPerMeter
	
	pixelX := width/2 + int(math.Round(deltaX))
	pixelY := height/2 + int(math.Round(deltaY))
	
	return PixelResult{X: pixelX, Y: pixelY, Warning: warning}
}

func PixelToLatLng(centerLat, centerLng float64, zoom, width, height int, pixelX, pixelY int) ConvertResult {
	var warning string
	
	if zoom < MinZoom || zoom > MaxZoom {
		zoom = clampZoom(zoom)
		warning = "zoom level out of range, clamped"
	}
	
	centerLat = clampLat(centerLat)
	centerLng = normalizeLng(centerLng)
	
	mercCenter := LatLngToMercator(centerLat, centerLng)
	
	scale := math.Pow(2.0, float64(zoom))
	metersPerPixel := MercatorEarthCircumference / (scale * float64(TileSizePixels))
	
	deltaX := float64(pixelX - width/2) * metersPerPixel
	deltaY := float64(height/2 - pixelY) * metersPerPixel
	
	targetX := mercCenter.X + deltaX
	targetY := mercCenter.Y + deltaY
	
	result := MercatorToLatLng(targetX, targetY)
	if warning != "" && result.Warning == "" {
		result.Warning = warning
	} else if warning != "" && result.Warning != "" {
		result.Warning = warning + "; " + result.Warning
	}
	
	return result
}

func GetViewBounds(centerLat, centerLng float64, zoom, width, height int) ViewBoundsResult {
	var warning string
	
	if zoom < MinZoom || zoom > MaxZoom {
		zoom = clampZoom(zoom)
		warning = "zoom level out of range, clamped"
	}
	
	nwResult := PixelToLatLng(centerLat, centerLng, zoom, width, height, 0, 0)
	neResult := PixelToLatLng(centerLat, centerLng, zoom, width, height, width, 0)
	swResult := PixelToLatLng(centerLat, centerLng, zoom, width, height, 0, height)
	seResult := PixelToLatLng(centerLat, centerLng, zoom, width, height, width, height)
	
	return ViewBoundsResult{
		NW:      LatLng{Lat: nwResult.Y, Lng: nwResult.X},
		NE:      LatLng{Lat: neResult.Y, Lng: neResult.X},
		SW:      LatLng{Lat: swResult.Y, Lng: swResult.X},
		SE:      LatLng{Lat: seResult.Y, Lng: seResult.X},
		Warning: warning,
	}
}
