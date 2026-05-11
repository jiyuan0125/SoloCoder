package gomod

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	semVerRegex = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)
	pseudoRegex = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)-(\d{14})-([a-f0-9]+)(?:-([0-9A-Za-z-]+))?$`)
)

func ParseVersion(version string) (Version, error) {
	if version == "" {
		return Version{}, fmt.Errorf("empty version")
	}
	
	if strings.HasPrefix(version, "v") {
		if matches := pseudoRegex.FindStringSubmatch(version); matches != nil {
			major, _ := strconv.Atoi(matches[1])
			minor, _ := strconv.Atoi(matches[2])
			patch, _ := strconv.Atoi(matches[3])
			
			return Version{
				Raw:        version,
				Major:      major,
				Minor:      minor,
				Patch:      patch,
				IsPseudo:   true,
				PseudoTime: matches[4],
				PseudoHash: matches[5],
				PreRelease: matches[6],
			}, nil
		}
		
		if matches := semVerRegex.FindStringSubmatch(version); matches != nil {
			major, _ := strconv.Atoi(matches[1])
			minor, _ := strconv.Atoi(matches[2])
			patch, _ := strconv.Atoi(matches[3])
			
			return Version{
				Raw:        version,
				Major:      major,
				Minor:      minor,
				Patch:      patch,
				IsPseudo:   false,
				PreRelease: matches[4],
				Build:      matches[5],
			}, nil
		}
	}
	
	return Version{}, fmt.Errorf("invalid version format: %s", version)
}

func CompareVersions(a, b string) (int, error) {
	va, err := ParseVersion(a)
	if err != nil {
		return 0, err
	}
	
	vb, err := ParseVersion(b)
	if err != nil {
		return 0, err
	}
	
	if va.Less(vb) {
		return -1, nil
	}
	if va.Greater(vb) {
		return 1, nil
	}
	return 0, nil
}

func MaxVersion(versions []string) (string, error) {
	if len(versions) == 0 {
		return "", fmt.Errorf("no versions to compare")
	}
	
	max := versions[0]
	for _, v := range versions[1:] {
		cmp, err := CompareVersions(v, max)
		if err != nil {
			return "", err
		}
		if cmp > 0 {
			max = v
		}
	}
	
	return max, nil
}

func MinVersion(versions []string) (string, error) {
	if len(versions) == 0 {
		return "", fmt.Errorf("no versions to compare")
	}
	
	min := versions[0]
	for _, v := range versions[1:] {
		cmp, err := CompareVersions(v, min)
		if err != nil {
			return "", err
		}
		if cmp < 0 {
			min = v
		}
	}
	
	return min, nil
}

func GetMajorVersionPath(path string, major int) string {
	if major < 2 {
		return path
	}
	
	suffix := fmt.Sprintf("/v%d", major)
	if strings.HasSuffix(path, suffix) {
		return path
	}
	
	return path + suffix
}

func GetVersionMajor(version string) (int, error) {
	v, err := ParseVersion(version)
	if err != nil {
		return 0, err
	}
	return v.Major, nil
}
