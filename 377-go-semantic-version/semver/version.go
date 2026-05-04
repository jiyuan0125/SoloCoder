package semver

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease []string
}

func Parse(versionStr string) (*Version, error) {
	versionStr = strings.TrimSpace(versionStr)
	if versionStr == "" {
		return nil, fmt.Errorf("version string is empty")
	}

	versionStr = strings.TrimPrefix(versionStr, "v")
	versionStr = strings.TrimPrefix(versionStr, "V")

	parts := strings.SplitN(versionStr, "-", 2)
	versionPart := parts[0]
	preReleasePart := ""
	if len(parts) > 1 {
		preReleasePart = parts[1]
	}

	versionComponents := strings.Split(versionPart, ".")
	if len(versionComponents) < 2 || len(versionComponents) > 3 {
		return nil, fmt.Errorf("invalid version format: %s", versionStr)
	}

	major, err := strconv.Atoi(versionComponents[0])
	if err != nil || major < 0 {
		return nil, fmt.Errorf("invalid major version: %s", versionComponents[0])
	}

	minor, err := strconv.Atoi(versionComponents[1])
	if err != nil || minor < 0 {
		return nil, fmt.Errorf("invalid minor version: %s", versionComponents[1])
	}

	patch := 0
	if len(versionComponents) == 3 {
		patch, err = strconv.Atoi(versionComponents[2])
		if err != nil || patch < 0 {
			return nil, fmt.Errorf("invalid patch version: %s", versionComponents[2])
		}
	}

	preRelease := []string{}
	if preReleasePart != "" {
		preRelease = strings.Split(preReleasePart, ".")
		for _, part := range preRelease {
			if part == "" {
				return nil, fmt.Errorf("invalid pre-release identifier: empty part")
			}
			for _, c := range part {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '-') {
					return nil, fmt.Errorf("invalid pre-release identifier: %s", part)
				}
			}
		}
	}

	return &Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		PreRelease: preRelease,
	}, nil
}

func (v *Version) String() string {
	base := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if len(v.PreRelease) > 0 {
		return base + "-" + strings.Join(v.PreRelease, ".")
	}
	return base
}

func Compare(v1, v2 *Version) int {
	if v1.Major != v2.Major {
		if v1.Major > v2.Major {
			return 1
		}
		return -1
	}

	if v1.Minor != v2.Minor {
		if v1.Minor > v2.Minor {
			return 1
		}
		return -1
	}

	if v1.Patch != v2.Patch {
		if v1.Patch > v2.Patch {
			return 1
		}
		return -1
	}

	v1HasPreRelease := len(v1.PreRelease) > 0
	v2HasPreRelease := len(v2.PreRelease) > 0

	if !v1HasPreRelease && !v2HasPreRelease {
		return 0
	}

	if !v1HasPreRelease && v2HasPreRelease {
		return 1
	}

	if v1HasPreRelease && !v2HasPreRelease {
		return -1
	}

	for i := 0; i < len(v1.PreRelease) || i < len(v2.PreRelease); i++ {
		if i >= len(v1.PreRelease) {
			return -1
		}
		if i >= len(v2.PreRelease) {
			return 1
		}

		result := comparePreReleaseParts(v1.PreRelease[i], v2.PreRelease[i])
		if result != 0 {
			return result
		}
	}

	return 0
}

func comparePreReleaseParts(p1, p2 string) int {
	p1IsNumeric := isNumeric(p1)
	p2IsNumeric := isNumeric(p2)

	if p1IsNumeric && p2IsNumeric {
		num1, _ := strconv.Atoi(p1)
		num2, _ := strconv.Atoi(p2)
		if num1 > num2 {
			return 1
		} else if num1 < num2 {
			return -1
		}
		return 0
	}

	if p1IsNumeric && !p2IsNumeric {
		return -1
	}

	if !p1IsNumeric && p2IsNumeric {
		return 1
	}

	if p1 > p2 {
		return 1
	} else if p1 < p2 {
		return -1
	}
	return 0
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func (v *Version) IncrementMajor() {
	v.Major++
	v.Minor = 0
	v.Patch = 0
	v.PreRelease = []string{}
}

func (v *Version) IncrementMinor() {
	v.Minor++
	v.Patch = 0
	v.PreRelease = []string{}
}

func (v *Version) IncrementPatch() {
	v.Patch++
	v.PreRelease = []string{}
}

type VersionRange struct {
	MinVersion *Version
	MaxVersion *Version
	MinInclusive bool
	MaxInclusive bool
}

func ParseRange(rangeStr string) (*VersionRange, error) {
	rangeStr = strings.TrimSpace(rangeStr)
	if rangeStr == "" {
		return nil, fmt.Errorf("range string is empty")
	}

	var minInclusive, maxInclusive bool
	var minStr, maxStr string

	switch {
	case strings.HasPrefix(rangeStr, "[") && strings.HasSuffix(rangeStr, "]"):
		minInclusive = true
		maxInclusive = true
		rangeStr = rangeStr[1 : len(rangeStr)-1]
	case strings.HasPrefix(rangeStr, "[") && strings.HasSuffix(rangeStr, ")"):
		minInclusive = true
		maxInclusive = false
		rangeStr = rangeStr[1 : len(rangeStr)-1]
	case strings.HasPrefix(rangeStr, "(") && strings.HasSuffix(rangeStr, "]"):
		minInclusive = false
		maxInclusive = true
		rangeStr = rangeStr[1 : len(rangeStr)-1]
	case strings.HasPrefix(rangeStr, "(") && strings.HasSuffix(rangeStr, ")"):
		minInclusive = false
		maxInclusive = false
		rangeStr = rangeStr[1 : len(rangeStr)-1]
	default:
		return nil, fmt.Errorf("invalid range format: %s", rangeStr)
	}

	parts := strings.SplitN(rangeStr, ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("range must contain exactly two versions: %s", rangeStr)
	}

	minStr = strings.TrimSpace(parts[0])
	maxStr = strings.TrimSpace(parts[1])

	minVersion, err := Parse(minStr)
	if err != nil {
		return nil, fmt.Errorf("invalid min version: %s", err)
	}

	maxVersion, err := Parse(maxStr)
	if err != nil {
		return nil, fmt.Errorf("invalid max version: %s", err)
	}

	if Compare(minVersion, maxVersion) > 0 {
		return nil, fmt.Errorf("min version cannot be greater than max version")
	}

	return &VersionRange{
		MinVersion:   minVersion,
		MaxVersion:   maxVersion,
		MinInclusive: minInclusive,
		MaxInclusive: maxInclusive,
	}, nil
}

func (v *Version) IsInRange(r *VersionRange) bool {
	minComp := Compare(v, r.MinVersion)
	maxComp := Compare(v, r.MaxVersion)

	meetsMin := false
	if r.MinInclusive {
		meetsMin = minComp >= 0
	} else {
		meetsMin = minComp > 0
	}

	meetsMax := false
	if r.MaxInclusive {
		meetsMax = maxComp <= 0
	} else {
		meetsMax = maxComp < 0
	}

	return meetsMin && meetsMax
}

func (vr *VersionRange) String() string {
	var minDelim, maxDelim string
	if vr.MinInclusive {
		minDelim = "["
	} else {
		minDelim = "("
	}

	if vr.MaxInclusive {
		maxDelim = "]"
	} else {
		maxDelim = ")"
	}

	return fmt.Sprintf("%s%s,%s%s", minDelim, vr.MinVersion.String(), vr.MaxVersion.String(), maxDelim)
}
