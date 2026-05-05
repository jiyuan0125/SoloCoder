package common

import (
	"fmt"
)

func FormatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	units := []string{"B", "K", "M", "G", "T", "P"}
	return fmt.Sprintf("%.1f%s", float64(size)/float64(div), units[exp])
}

func ParseSize(sizeStr string) (int64, error) {
	var size float64
	var unit string
	
	n, err := fmt.Sscanf(sizeStr, "%f%s", &size, &unit)
	if err != nil {
		if n == 1 {
			return int64(size), nil
		}
		return 0, err
	}
	
	unitMap := map[string]int64{
		"B": 1,
		"K": 1024,
		"M": 1024 * 1024,
		"G": 1024 * 1024 * 1024,
		"T": 1024 * 1024 * 1024 * 1024,
	}
	
	multiplier, ok := unitMap[unit]
	if !ok {
		return 0, fmt.Errorf("无效的单位: %s", unit)
	}
	
	return int64(size * float64(multiplier)), nil
}
