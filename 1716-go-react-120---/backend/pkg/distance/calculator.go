package distance

import (
	"strings"
)

type AreaType string

const (
	AreaTypeUrban  AreaType = "urban"
	AreaTypeSuburban AreaType = "suburban"
)

var areaDistanceMatrix = map[string]map[string]float64{
	"市中心": {
		"市中心": 2.0,
		"东城区": 5.0,
		"西城区": 4.0,
		"南城区": 6.0,
		"北城区": 5.0,
		"东郊": 15.0,
		"西郊": 12.0,
	},
	"东城区": {
		"市中心": 5.0,
		"东城区": 3.0,
		"西城区": 8.0,
		"南城区": 7.0,
		"北城区": 6.0,
		"东郊": 10.0,
		"西郊": 18.0,
	},
	"西城区": {
		"市中心": 4.0,
		"东城区": 8.0,
		"西城区": 3.0,
		"南城区": 8.0,
		"北城区": 7.0,
		"东郊": 20.0,
		"西郊": 8.0,
	},
	"南城区": {
		"市中心": 6.0,
		"东城区": 7.0,
		"西城区": 8.0,
		"南城区": 3.0,
		"北城区": 10.0,
		"东郊": 12.0,
		"西郊": 15.0,
	},
	"北城区": {
		"市中心": 5.0,
		"东城区": 6.0,
		"西城区": 7.0,
		"南城区": 10.0,
		"北城区": 3.0,
		"东郊": 18.0,
		"西郊": 10.0,
	},
	"东郊": {
		"市中心": 15.0,
		"东城区": 10.0,
		"西城区": 20.0,
		"南城区": 12.0,
		"北城区": 18.0,
		"东郊": 5.0,
		"西郊": 30.0,
	},
	"西郊": {
		"市中心": 12.0,
		"东城区": 18.0,
		"西城区": 8.0,
		"南城区": 15.0,
		"北城区": 10.0,
		"东郊": 30.0,
		"西郊": 5.0,
	},
}

var suburbanAreas = map[string]bool{
	"东郊": true,
	"西郊": true,
}

func DetectArea(address string) string {
	areas := []string{"市中心", "东城区", "西城区", "南城区", "北城区", "东郊", "西郊"}
	for _, area := range areas {
		if strings.Contains(address, area) {
			return area
		}
	}
	return "市中心"
}

func GetAreaType(area string) AreaType {
	if suburbanAreas[area] {
		return AreaTypeSuburban
	}
	return AreaTypeUrban
}

func CalculateDistance(from, to string) float64 {
	fromArea := DetectArea(from)
	toArea := DetectArea(to)
	
	if distances, ok := areaDistanceMatrix[fromArea]; ok {
		if dist, ok := distances[toArea]; ok {
			return dist
		}
	}
	
	return 8.0
}

func CalculateTime(distance float64, areaType AreaType) int {
	if areaType == AreaTypeUrban {
		return int((distance / 3.0) * 10)
	}
	return int((distance / 5.0) * 10)
}

func CalculateRoute(from, to string) (float64, int) {
	distance := CalculateDistance(from, to)
	areaType := GetAreaType(DetectArea(to))
	time := CalculateTime(distance, areaType)
	return distance, time
}
