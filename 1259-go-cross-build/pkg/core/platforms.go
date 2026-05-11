package core

var validGOOS = map[string]bool{
	"android":   true,
	"darwin":    true,
	"dragonfly": true,
	"freebsd":   true,
	"illumos":   true,
	"ios":       true,
	"js":        true,
	"linux":     true,
	"netbsd":    true,
	"openbsd":   true,
	"plan9":     true,
	"solaris":   true,
	"windows":   true,
	"aix":       true,
}

var validGOARCH = map[string]bool{
	"386":        true,
	"amd64":      true,
	"amd64p32":   true,
	"arm":        true,
	"armbe":      true,
	"arm64":      true,
	"arm64be":    true,
	"ppc64":      true,
	"ppc64le":    true,
	"mips":       true,
	"mipsle":     true,
	"mips64":     true,
	"mips64le":   true,
	"mips64p32":  true,
	"mips64p32le": true,
	"ppc":        true,
	"riscv":      true,
	"riscv64":    true,
	"s390":       true,
	"s390x":      true,
	"sparc":      true,
	"sparc64":    true,
	"wasm":       true,
}

var invalidCombinations = map[string]bool{
	"windows/arm":      true,
	"windows/armbe":    true,
	"windows/arm64be":  true,
	"darwin/386":       true,
	"darwin/arm":       true,
	"darwin/armbe":     true,
	"linux/arm64be":    true,
	"freebsd/arm64be":  true,
	"netbsd/arm64be":   true,
	"openbsd/arm64be":  true,
}

func IsValidGOOS(goos string) bool {
	return validGOOS[goos]
}

func IsValidGOARCH(goarch string) bool {
	return validGOARCH[goarch]
}

func IsValidCombination(goos, goarch string) bool {
	if !IsValidGOOS(goos) || !IsValidGOARCH(goarch) {
		return false
	}
	key := goos + "/" + goarch
	return !invalidCombinations[key]
}

func GetValidPlatforms() []string {
	var platforms []string
	for goos := range validGOOS {
		for goarch := range validGOARCH {
			if IsValidCombination(goos, goarch) {
				platforms = append(platforms, goos+"/"+goarch)
			}
		}
	}
	return platforms
}

func GetDefaultTargets() []struct{ GOOS, GOARCH string } {
	return []struct{ GOOS, GOARCH string }{
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"windows", "amd64"},
		{"windows", "arm64"},
	}
}
