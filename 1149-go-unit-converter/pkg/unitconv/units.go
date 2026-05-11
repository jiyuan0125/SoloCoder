package unitconv

import (
	"fmt"
	"strings"
)

type UnitType string

const (
	UnitTypeLength   UnitType = "length"
	UnitTypeWeight   UnitType = "weight"
	UnitTypeTemp     UnitType = "temperature"
	UnitTypeArea     UnitType = "area"
	UnitTypeVolume   UnitType = "volume"
	UnitTypeTime     UnitType = "time"
	UnitTypeCompound UnitType = "compound"
)

type UnitInfo struct {
	Name      string
	Aliases   []string
	UnitType  UnitType
	BaseValue float64
	Offset    float64
}

var units map[string]*UnitInfo
var unitTypeMap map[string]UnitType
var allUnitNames []string

func init() {
	units = make(map[string]*UnitInfo)
	unitTypeMap = make(map[string]UnitType)
	allUnitNames = []string{}

	lengthUnits := []*UnitInfo{
		{Name: "mm", Aliases: []string{"millimeter", "millimeters"}, UnitType: UnitTypeLength, BaseValue: 0.001},
		{Name: "cm", Aliases: []string{"centimeter", "centimeters"}, UnitType: UnitTypeLength, BaseValue: 0.01},
		{Name: "m", Aliases: []string{"meter", "meters"}, UnitType: UnitTypeLength, BaseValue: 1},
		{Name: "km", Aliases: []string{"kilometer", "kilometers"}, UnitType: UnitTypeLength, BaseValue: 1000},
		{Name: "in", Aliases: []string{"inch", "inches"}, UnitType: UnitTypeLength, BaseValue: 0.0254},
		{Name: "ft", Aliases: []string{"foot", "feet"}, UnitType: UnitTypeLength, BaseValue: 0.3048},
		{Name: "mi", Aliases: []string{"mile", "miles"}, UnitType: UnitTypeLength, BaseValue: 1609.344},
	}

	weightUnits := []*UnitInfo{
		{Name: "mg", Aliases: []string{"milligram", "milligrams"}, UnitType: UnitTypeWeight, BaseValue: 0.001},
		{Name: "g", Aliases: []string{"gram", "grams"}, UnitType: UnitTypeWeight, BaseValue: 1},
		{Name: "kg", Aliases: []string{"kilogram", "kilograms"}, UnitType: UnitTypeWeight, BaseValue: 1000},
		{Name: "t", Aliases: []string{"ton", "tons", "tonne", "tonnes"}, UnitType: UnitTypeWeight, BaseValue: 1000000},
		{Name: "oz", Aliases: []string{"ounce", "ounces"}, UnitType: UnitTypeWeight, BaseValue: 28.349523125},
		{Name: "lb", Aliases: []string{"pound", "pounds"}, UnitType: UnitTypeWeight, BaseValue: 453.59237},
	}

	tempUnits := []*UnitInfo{
		{Name: "C", Aliases: []string{"celsius", "degrees_celsius", "°c"}, UnitType: UnitTypeTemp, BaseValue: 1, Offset: 0},
		{Name: "F", Aliases: []string{"fahrenheit", "degrees_fahrenheit", "°f"}, UnitType: UnitTypeTemp, BaseValue: 5.0 / 9.0, Offset: 32},
		{Name: "K", Aliases: []string{"kelvin", "kelvins"}, UnitType: UnitTypeTemp, BaseValue: 1, Offset: 273.15},
	}

	areaUnits := []*UnitInfo{
		{Name: "m²", Aliases: []string{"m2", "square_meter", "square_meters"}, UnitType: UnitTypeArea, BaseValue: 1},
		{Name: "km²", Aliases: []string{"km2", "square_kilometer", "square_kilometers"}, UnitType: UnitTypeArea, BaseValue: 1000000},
		{Name: "ha", Aliases: []string{"hectare", "hectares"}, UnitType: UnitTypeArea, BaseValue: 10000},
		{Name: "mu", Aliases: []string{"亩"}, UnitType: UnitTypeArea, BaseValue: 10000.0 / 15.0},
		{Name: "ft²", Aliases: []string{"ft2", "square_foot", "square_feet"}, UnitType: UnitTypeArea, BaseValue: 0.09290304},
		{Name: "ac", Aliases: []string{"acre", "acres"}, UnitType: UnitTypeArea, BaseValue: 4046.8564224},
	}

	volumeUnits := []*UnitInfo{
		{Name: "ml", Aliases: []string{"milliliter", "milliliters"}, UnitType: UnitTypeVolume, BaseValue: 0.001},
		{Name: "l", Aliases: []string{"liter", "liters"}, UnitType: UnitTypeVolume, BaseValue: 1},
		{Name: "m³", Aliases: []string{"m3", "cubic_meter", "cubic_meters"}, UnitType: UnitTypeVolume, BaseValue: 1000},
		{Name: "gal_us", Aliases: []string{"us_gallon", "us_gallons", "gallon_us"}, UnitType: UnitTypeVolume, BaseValue: 3.785411784},
		{Name: "gal_uk", Aliases: []string{"uk_gallon", "uk_gallons", "gallon_uk"}, UnitType: UnitTypeVolume, BaseValue: 4.54609},
	}

	timeUnits := []*UnitInfo{
		{Name: "s", Aliases: []string{"second", "seconds", "sec"}, UnitType: UnitTypeTime, BaseValue: 1},
		{Name: "min", Aliases: []string{"minute", "minutes"}, UnitType: UnitTypeTime, BaseValue: 60},
		{Name: "h", Aliases: []string{"hour", "hours"}, UnitType: UnitTypeTime, BaseValue: 3600},
		{Name: "d", Aliases: []string{"day", "days"}, UnitType: UnitTypeTime, BaseValue: 86400},
	}

	allUnits := [][]*UnitInfo{lengthUnits, weightUnits, tempUnits, areaUnits, volumeUnits, timeUnits}

	for _, group := range allUnits {
		for _, u := range group {
			units[strings.ToLower(u.Name)] = u
			unitTypeMap[strings.ToLower(u.Name)] = u.UnitType
			allUnitNames = append(allUnitNames, u.Name)
			for _, alias := range u.Aliases {
				aliasLower := strings.ToLower(alias)
				if _, exists := units[aliasLower]; !exists {
					units[aliasLower] = u
					unitTypeMap[aliasLower] = u.UnitType
				}
			}
		}
	}
}

func GetUnitInfo(name string) (*UnitInfo, bool) {
	info, ok := units[strings.ToLower(name)]
	return info, ok
}

func GetUnitType(name string) (UnitType, bool) {
	t, ok := unitTypeMap[strings.ToLower(name)]
	return t, ok
}

func GetAllSupportedUnits() string {
	return fmt.Sprintf("Supported units: %s", strings.Join(allUnitNames, ", "))
}

func IsTemperatureUnit(name string) bool {
	t, ok := GetUnitType(name)
	return ok && t == UnitTypeTemp
}
