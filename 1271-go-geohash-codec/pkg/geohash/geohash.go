package geohash

import (
	"fmt"
	"math"
	"strings"
)

const (
	minPrecision = 1
	maxPrecision = 12
)

var base32 = "0123456789bcdefghjkmnpqrstuvwxyz"

var base32Map = make(map[rune]int)

func init() {
	for i, c := range base32 {
		base32Map[c] = i
	}
}

type BoundingBox struct {
	MinLat, MaxLat, MinLng, MaxLng float64
}

func (bb BoundingBox) Center() (lat, lng float64) {
	return (bb.MinLat + bb.MaxLat) / 2, (bb.MinLng + bb.MaxLng) / 2
}

func (bb BoundingBox) Contains(lat, lng float64) bool {
	return lat >= bb.MinLat && lat <= bb.MaxLat &&
		lng >= bb.MinLng && lng <= bb.MaxLng
}

func Encode(lat, lng float64, precision int) (string, error) {
	if precision < minPrecision || precision > maxPrecision {
		return "", fmt.Errorf("精度必须在 %d 到 %d 之间", minPrecision, maxPrecision)
	}
	if lat < -90 || lat > 90 {
		return "", fmt.Errorf("纬度必须在 -90 到 90 之间")
	}
	if lng < -180 || lng > 180 {
		return "", fmt.Errorf("经度必须在 -180 到 180 之间")
	}

	var geohash strings.Builder
	latRange := [2]float64{-90, 90}
	lngRange := [2]float64{-180, 180}
	isEven := true
	var bits uint8
	var bitIndex uint8 = 0

	for geohash.Len() < precision {
		var mid float64
		if isEven {
			mid = (lngRange[0] + lngRange[1]) / 2
			if lng >= mid {
				bits |= 1 << (4 - bitIndex)
				lngRange[0] = mid
			} else {
				lngRange[1] = mid
			}
		} else {
			mid = (latRange[0] + latRange[1]) / 2
			if lat >= mid {
				bits |= 1 << (4 - bitIndex)
				latRange[0] = mid
			} else {
				latRange[1] = mid
			}
		}

		isEven = !isEven
		bitIndex++

		if bitIndex == 5 {
			geohash.WriteByte(base32[bits])
			bits = 0
			bitIndex = 0
		}
	}

	return geohash.String(), nil
}

func Decode(geohash string) (lat, lng float64, err error) {
	bbox, err := BoundingBoxOf(geohash)
	if err != nil {
		return 0, 0, err
	}
	lat, lng = bbox.Center()
	return lat, lng, nil
}

func BoundingBoxOf(geohash string) (BoundingBox, error) {
	if len(geohash) == 0 {
		return BoundingBox{}, fmt.Errorf("Geohash 不能为空")
	}
	if len(geohash) > maxPrecision {
		return BoundingBox{}, fmt.Errorf("Geohash 长度不能超过 %d", maxPrecision)
	}

	latRange := [2]float64{-90, 90}
	lngRange := [2]float64{-180, 180}
	isEven := true

	for _, c := range geohash {
		idx, ok := base32Map[c]
		if !ok {
			return BoundingBox{}, fmt.Errorf("无效的 Geohash 字符: %c", c)
		}

		for bitPos := 4; bitPos >= 0; bitPos-- {
			bit := (idx >> bitPos) & 1
			var mid float64

			if isEven {
				mid = (lngRange[0] + lngRange[1]) / 2
				if bit == 1 {
					lngRange[0] = mid
				} else {
					lngRange[1] = mid
				}
			} else {
				mid = (latRange[0] + latRange[1]) / 2
				if bit == 1 {
					latRange[0] = mid
				} else {
					latRange[1] = mid
				}
			}

			isEven = !isEven
		}
	}

	return BoundingBox{
		MinLat: latRange[0],
		MaxLat: latRange[1],
		MinLng: lngRange[0],
		MaxLng: lngRange[1],
	}, nil
}

func validateGeohash(geohash string) error {
	if len(geohash) == 0 {
		return fmt.Errorf("Geohash 不能为空")
	}
	if len(geohash) > maxPrecision {
		return fmt.Errorf("Geohash 长度不能超过 %d", maxPrecision)
	}
	for _, c := range geohash {
		if _, ok := base32Map[c]; !ok {
			return fmt.Errorf("无效的 Geohash 字符: %c", c)
		}
	}
	return nil
}

func Neighbor(geohash string, direction Direction) (string, error) {
	if err := validateGeohash(geohash); err != nil {
		return "", err
	}

	bbox, err := BoundingBoxOf(geohash)
	if err != nil {
		return "", err
	}

	centerLat, centerLng := bbox.Center()
	latStep := bbox.MaxLat - bbox.MinLat
	lngStep := bbox.MaxLng - bbox.MinLng

	var newLat, newLng float64

	switch direction {
	case North:
		newLat = centerLat + latStep
		newLng = centerLng
	case Northeast:
		newLat = centerLat + latStep
		newLng = centerLng + lngStep
	case East:
		newLat = centerLat
		newLng = centerLng + lngStep
	case Southeast:
		newLat = centerLat - latStep
		newLng = centerLng + lngStep
	case South:
		newLat = centerLat - latStep
		newLng = centerLng
	case Southwest:
		newLat = centerLat - latStep
		newLng = centerLng - lngStep
	case West:
		newLat = centerLat
		newLng = centerLng - lngStep
	case Northwest:
		newLat = centerLat + latStep
		newLng = centerLng - lngStep
	default:
		return "", fmt.Errorf("无效的方向")
	}

	if newLat > 90 {
		newLat = 90 - (newLat - 90)
		newLng = normalizeLng(newLng + 180)
	}
	if newLat < -90 {
		newLat = -90 + (-90 - newLat)
		newLng = normalizeLng(newLng + 180)
	}

	newLng = normalizeLng(newLng)

	return Encode(newLat, newLng, len(geohash))
}

func Neighbors(geohash string) (map[Direction]string, error) {
	if err := validateGeohash(geohash); err != nil {
		return nil, err
	}

	neighbors := make(map[Direction]string)

	for _, dir := range allDirections {
		neighbor, err := Neighbor(geohash, dir)
		if err != nil {
			return nil, err
		}
		neighbors[dir] = neighbor
	}

	return neighbors, nil
}

func normalizeLng(lng float64) float64 {
	for lng > 180 {
		lng -= 360
	}
	for lng < -180 {
		lng += 360
	}
	return lng
}

type Direction int

const (
	North Direction = iota
	Northeast
	East
	Southeast
	South
	Southwest
	West
	Northwest
)

var allDirections = []Direction{North, Northeast, East, Southeast, South, Southwest, West, Northwest}

func (d Direction) String() string {
	switch d {
	case North:
		return "north"
	case Northeast:
		return "northeast"
	case East:
		return "east"
	case Southeast:
		return "southeast"
	case South:
		return "south"
	case Southwest:
		return "southwest"
	case West:
		return "west"
	case Northwest:
		return "northwest"
	default:
		return "unknown"
	}
}

func HaversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKm = 6371.0

	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLng := (lng2 - lng1) * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

func circleBBoxIntersects(circleLat, circleLng, radiusKm float64, bbox BoundingBox) bool {
	if bbox.Contains(circleLat, circleLng) {
		return true
	}

	centerLat, centerLng := bbox.Center()
	distance := HaversineDistance(circleLat, circleLng, centerLat, centerLng)

	diagonalHalf := HaversineDistance(centerLat, centerLng, bbox.MaxLat, bbox.MaxLng)

	return distance <= radiusKm+diagonalHalf
}

func getPrecisionForRadius(radiusKm float64) int {
	switch {
	case radiusKm >= 2500:
		return 1
	case radiusKm >= 625:
		return 2
	case radiusKm >= 156:
		return 3
	case radiusKm >= 39:
		return 4
	case radiusKm >= 4.9:
		return 5
	case radiusKm >= 1.2:
		return 6
	case radiusKm >= 0.152:
		return 7
	case radiusKm >= 0.038:
		return 8
	case radiusKm >= 0.00477:
		return 9
	case radiusKm >= 0.00119:
		return 10
	case radiusKm >= 0.000149:
		return 11
	default:
		return 12
	}
}

func ProximitySearch(centerLat, centerLng, radiusKm float64) ([]string, error) {
	if centerLat < -90 || centerLat > 90 {
		return nil, fmt.Errorf("纬度必须在 -90 到 90 之间")
	}
	if centerLng < -180 || centerLng > 180 {
		return nil, fmt.Errorf("经度必须在 -180 到 180 之间")
	}
	if radiusKm < 0 {
		return nil, fmt.Errorf("半径不能为负数")
	}

	precision := getPrecisionForRadius(radiusKm)

	initialHash, err := Encode(centerLat, centerLng, precision)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	queue := []string{initialHash}
	var result []string

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if seen[current] {
			continue
		}
		seen[current] = true

		bbox, err := BoundingBoxOf(current)
		if err != nil {
			continue
		}

		if circleBBoxIntersects(centerLat, centerLng, radiusKm, bbox) {
			result = append(result, current)

			neighbors, err := Neighbors(current)
			if err != nil {
				continue
			}

			for _, neighbor := range neighbors {
				if !seen[neighbor] {
					queue = append(queue, neighbor)
				}
			}
		}
	}

	return result, nil
}

func VerifyDistance(centerLat, centerLng, lat, lng, radiusKm float64) bool {
	distance := HaversineDistance(centerLat, centerLng, lat, lng)
	return distance <= radiusKm
}
