package parser

import (
	"errors"
	"regexp"
	"strings"
)

func Parse(address string) (*Address, error) {
	cleaned := cleanAddress(address)
	if cleaned == "" {
		return nil, errors.New("address is empty or contains only whitespace")
	}
	
	result := &Address{}
	remaining := cleaned
	
	remaining = matchProvince(result, remaining)
	remaining = strings.TrimSpace(remaining)
	remaining = matchCity(result, remaining)
	remaining = strings.TrimSpace(remaining)
	remaining = matchDistrict(result, remaining)
	remaining = strings.TrimSpace(remaining)
	
	inferMissingParts(result)
	handleMunicipalities(result)
	
	result.Detail = strings.TrimSpace(remaining)
	
	return result, nil
}

func cleanAddress(address string) string {
	address = strings.ReplaceAll(address, ",", " ")
	address = strings.ReplaceAll(address, "，", " ")
	
	reg := regexp.MustCompile(`\s+`)
	address = reg.ReplaceAllString(address, " ")
	
	return strings.TrimSpace(address)
}

func matchProvince(result *Address, remaining string) string {
	bestMatch := ""
	bestLength := 0
	
	for _, div := range divisions {
		fullName := div.Province
		shortName := strings.TrimSuffix(fullName, "省")
		shortName = strings.TrimSuffix(shortName, "市")
		shortName = strings.TrimSuffix(shortName, "自治区")
		shortName = strings.TrimSuffix(shortName, "壮族")
		shortName = strings.TrimSuffix(shortName, "回族")
		shortName = strings.TrimSuffix(shortName, "维吾尔")
		
		if strings.HasPrefix(remaining, fullName) && len(fullName) > bestLength {
			bestMatch = fullName
			bestLength = len(fullName)
		} else if strings.HasPrefix(remaining, shortName) && len(shortName) > bestLength {
			bestMatch = shortName
			bestLength = len(shortName)
		}
	}
	
	if bestMatch != "" {
		for _, div := range divisions {
			if strings.Contains(div.Province, bestMatch) || strings.Contains(bestMatch, strings.TrimSuffix(div.Province, "省")) ||
			   strings.Contains(bestMatch, strings.TrimSuffix(strings.TrimSuffix(div.Province, "自治区"), "壮族")) ||
			   strings.Contains(bestMatch, strings.TrimSuffix(strings.TrimSuffix(div.Province, "自治区"), "回族")) ||
			   strings.Contains(bestMatch, strings.TrimSuffix(strings.TrimSuffix(div.Province, "自治区"), "维吾尔")) {
				result.Province = div.Province
				break
			}
		}
		remaining = remaining[bestLength:]
	}
	
	return remaining
}

func matchCity(result *Address, remaining string) string {
	bestMatch := ""
	bestLength := 0
	
	for _, div := range divisions {
		if result.Province != "" && !strings.Contains(div.Province, result.Province) && !strings.Contains(result.Province, div.Province) {
			continue
		}
		
		fullName := div.City
		shortName := strings.TrimSuffix(fullName, "市")
		shortName = strings.TrimSuffix(shortName, "区")
		
		if strings.HasPrefix(remaining, fullName) && len(fullName) > bestLength {
			bestMatch = fullName
			bestLength = len(fullName)
		} else if strings.HasPrefix(remaining, shortName) && len(shortName) > bestLength {
			bestMatch = shortName
			bestLength = len(shortName)
		}
	}
	
	if bestMatch != "" {
		for _, div := range divisions {
			if result.Province != "" && !strings.Contains(div.Province, result.Province) && !strings.Contains(result.Province, div.Province) {
				continue
			}
			
			if strings.Contains(div.City, bestMatch) || strings.Contains(bestMatch, strings.TrimSuffix(div.City, "市")) {
				result.City = div.City
				if result.Province == "" {
					result.Province = div.Province
				}
				break
			}
		}
		remaining = remaining[bestLength:]
	}
	
	return remaining
}

func matchDistrict(result *Address, remaining string) string {
	bestMatch := ""
	bestLength := 0
	bestDiv := Division{}
	
	for _, div := range divisions {
		if result.Province != "" && !strings.Contains(div.Province, result.Province) && !strings.Contains(result.Province, div.Province) {
			continue
		}
		if result.City != "" && !strings.Contains(div.City, result.City) && !strings.Contains(result.City, div.City) {
			continue
		}
		
		fullName := div.District
		shortName := strings.TrimSuffix(fullName, "区")
		shortName = strings.TrimSuffix(shortName, "县")
		
		if strings.HasPrefix(remaining, fullName) && len(fullName) > bestLength {
			bestMatch = fullName
			bestLength = len(fullName)
			bestDiv = div
		} else if strings.HasPrefix(remaining, shortName) && len(shortName) > bestLength {
			bestMatch = shortName
			bestLength = len(shortName)
			bestDiv = div
		}
	}
	
	if bestMatch != "" {
		result.District = bestDiv.District
		if result.City == "" {
			result.City = bestDiv.City
		}
		if result.Province == "" {
			result.Province = bestDiv.Province
		}
		remaining = remaining[bestLength:]
	}
	
	return remaining
}

func inferMissingParts(result *Address) {
	if result.District != "" {
		for _, div := range divisions {
			if div.District == result.District {
				if result.City == "" {
					result.City = div.City
				}
				if result.Province == "" {
					result.Province = div.Province
				}
				break
			}
		}
	}
	
	if result.City != "" {
		for _, div := range divisions {
			if div.City == result.City {
				if result.Province == "" {
					result.Province = div.Province
				}
				break
			}
		}
	}
}

func handleMunicipalities(result *Address) {
	municipalities := []string{"北京市", "上海市", "天津市", "重庆市"}
	
	for _, m := range municipalities {
		if result.City == m || result.Province == m {
			result.Province = ""
			if result.City == "" {
				result.City = m
			}
			break
		}
	}
}
