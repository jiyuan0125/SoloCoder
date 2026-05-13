package utils

import (
	"fmt"
	"image/color"
	"strings"
)

func ValidateFormat(format string) bool {
	switch strings.ToUpper(format) {
	case "PNG", "JPEG", "JPG", "SVG":
		return true
	default:
		return false
	}
}

func ValidateLogoFormat(filename string) bool {
	ext := strings.ToLower(filename)
	if strings.HasSuffix(ext, ".png") || strings.HasSuffix(ext, ".jpg") || strings.HasSuffix(ext, ".jpeg") {
		return true
	}
	return false
}

func ParseColor(hex string) (color.Color, error) {
	hex = strings.TrimPrefix(hex, "#")
	hex = strings.ToUpper(hex)
	if len(hex) == 3 {
		hex = fmt.Sprintf("%c%c%c%c%c%c", hex[0], hex[0], hex[1], hex[1], hex[2], hex[2])
	}
	if len(hex) != 6 {
		return nil, fmt.Errorf("invalid color format")
	}
	var r, g, b uint8
	_, err := fmt.Sscanf(hex, "%02X%02X%02X", &r, &g, &b)
	if err != nil {
		return nil, err
	}
	return color.RGBA{R: r, G: g, B: b, A: 255}, nil
}

func ParseErrorLevel(level string) int {
	switch strings.ToUpper(level) {
	case "L":
		return 1
	case "M":
		return 2
	case "Q":
		return 3
	case "H":
		return 4
	default:
		return 2
	}
}
