package uninorm

import (
	"sort"
	"unicode"
)

type decomposer struct {
	compat bool
}

func newDecomposer(compat bool) *decomposer {
	return &decomposer{compat: compat}
}

func (d *decomposer) decomposeString(s string) []rune {
	runes := []rune(s)
	result := make([]rune, 0, len(runes)*2)
	
	for _, r := range runes {
		result = append(result, d.decomposeRune(r)...)
	}
	
	return result
}

func (d *decomposer) decomposeRune(r rune) []rune {
	decomposed := d.doDecompose(r)
	if len(decomposed) == 0 {
		return []rune{r}
	}
	
	result := make([]rune, 0, len(decomposed))
	for _, dr := range decomposed {
		result = append(result, d.decomposeRune(dr)...)
	}
	
	return result
}

func (d *decomposer) doDecompose(r rune) []rune {
	if mapping, ok := decomposeMap[r]; ok {
		if !d.compat && mapping.compatOnly {
			return nil
		}
		return mapping.runes
	}
	return nil
}

type decomposition struct {
	runes      []rune
	compatOnly bool
}

var decomposeMap = map[rune]decomposition{
	0x00C9: {runes: []rune{0x0045, 0x0301}, compatOnly: false},
	0x00E9: {runes: []rune{0x0065, 0x0301}, compatOnly: false},
	0x00C8: {runes: []rune{0x0045, 0x0300}, compatOnly: false},
	0x00E8: {runes: []rune{0x0065, 0x0300}, compatOnly: false},
	0x00CA: {runes: []rune{0x0045, 0x0302}, compatOnly: false},
	0x00EA: {runes: []rune{0x0065, 0x0302}, compatOnly: false},
	0x00CB: {runes: []rune{0x0045, 0x0308}, compatOnly: false},
	0x00EB: {runes: []rune{0x0065, 0x0308}, compatOnly: false},
	
	0x00C1: {runes: []rune{0x0041, 0x0301}, compatOnly: false},
	0x00E1: {runes: []rune{0x0061, 0x0301}, compatOnly: false},
	0x00C0: {runes: []rune{0x0041, 0x0300}, compatOnly: false},
	0x00E0: {runes: []rune{0x0061, 0x0300}, compatOnly: false},
	0x00C2: {runes: []rune{0x0041, 0x0302}, compatOnly: false},
	0x00E2: {runes: []rune{0x0061, 0x0302}, compatOnly: false},
	0x00C4: {runes: []rune{0x0041, 0x0308}, compatOnly: false},
	0x00E4: {runes: []rune{0x0061, 0x0308}, compatOnly: false},
	
	0x00CD: {runes: []rune{0x0049, 0x0301}, compatOnly: false},
	0x00ED: {runes: []rune{0x0069, 0x0301}, compatOnly: false},
	0x00CC: {runes: []rune{0x0049, 0x0300}, compatOnly: false},
	0x00EC: {runes: []rune{0x0069, 0x0300}, compatOnly: false},
	0x00CE: {runes: []rune{0x0049, 0x0302}, compatOnly: false},
	0x00EE: {runes: []rune{0x0069, 0x0302}, compatOnly: false},
	0x00CF: {runes: []rune{0x0049, 0x0308}, compatOnly: false},
	0x00EF: {runes: []rune{0x0069, 0x0308}, compatOnly: false},
	
	0x00D3: {runes: []rune{0x004F, 0x0301}, compatOnly: false},
	0x00F3: {runes: []rune{0x006F, 0x0301}, compatOnly: false},
	0x00D2: {runes: []rune{0x004F, 0x0300}, compatOnly: false},
	0x00F2: {runes: []rune{0x006F, 0x0300}, compatOnly: false},
	0x00D4: {runes: []rune{0x004F, 0x0302}, compatOnly: false},
	0x00F4: {runes: []rune{0x006F, 0x0302}, compatOnly: false},
	0x00D6: {runes: []rune{0x004F, 0x0308}, compatOnly: false},
	0x00F6: {runes: []rune{0x006F, 0x0308}, compatOnly: false},
	
	0x00DA: {runes: []rune{0x0055, 0x0301}, compatOnly: false},
	0x00FA: {runes: []rune{0x0075, 0x0301}, compatOnly: false},
	0x00D9: {runes: []rune{0x0055, 0x0300}, compatOnly: false},
	0x00F9: {runes: []rune{0x0075, 0x0300}, compatOnly: false},
	0x00DB: {runes: []rune{0x0055, 0x0302}, compatOnly: false},
	0x00FB: {runes: []rune{0x0075, 0x0302}, compatOnly: false},
	0x00DC: {runes: []rune{0x0055, 0x0308}, compatOnly: false},
	0x00FC: {runes: []rune{0x0075, 0x0308}, compatOnly: false},
	
	0x00C7: {runes: []rune{0x0043, 0x0327}, compatOnly: false},
	0x00E7: {runes: []rune{0x0063, 0x0327}, compatOnly: false},
	0x00D1: {runes: []rune{0x004E, 0x0303}, compatOnly: false},
	0x00F1: {runes: []rune{0x006E, 0x0303}, compatOnly: false},
	
	0xFF21: {runes: []rune{0x0041}, compatOnly: true},
	0xFF22: {runes: []rune{0x0042}, compatOnly: true},
	0xFF23: {runes: []rune{0x0043}, compatOnly: true},
	0xFF24: {runes: []rune{0x0044}, compatOnly: true},
	0xFF25: {runes: []rune{0x0045}, compatOnly: true},
	0xFF26: {runes: []rune{0x0046}, compatOnly: true},
	0xFF27: {runes: []rune{0x0047}, compatOnly: true},
	0xFF28: {runes: []rune{0x0048}, compatOnly: true},
	0xFF29: {runes: []rune{0x0049}, compatOnly: true},
	0xFF2A: {runes: []rune{0x004A}, compatOnly: true},
	0xFF2B: {runes: []rune{0x004B}, compatOnly: true},
	0xFF2C: {runes: []rune{0x004C}, compatOnly: true},
	0xFF2D: {runes: []rune{0x004D}, compatOnly: true},
	0xFF2E: {runes: []rune{0x004E}, compatOnly: true},
	0xFF2F: {runes: []rune{0x004F}, compatOnly: true},
	0xFF30: {runes: []rune{0x0050}, compatOnly: true},
	0xFF31: {runes: []rune{0x0051}, compatOnly: true},
	0xFF32: {runes: []rune{0x0052}, compatOnly: true},
	0xFF33: {runes: []rune{0x0053}, compatOnly: true},
	0xFF34: {runes: []rune{0x0054}, compatOnly: true},
	0xFF35: {runes: []rune{0x0055}, compatOnly: true},
	0xFF36: {runes: []rune{0x0056}, compatOnly: true},
	0xFF37: {runes: []rune{0x0057}, compatOnly: true},
	0xFF38: {runes: []rune{0x0058}, compatOnly: true},
	0xFF39: {runes: []rune{0x0059}, compatOnly: true},
	0xFF3A: {runes: []rune{0x005A}, compatOnly: true},
	
	0xFF41: {runes: []rune{0x0061}, compatOnly: true},
	0xFF42: {runes: []rune{0x0062}, compatOnly: true},
	0xFF43: {runes: []rune{0x0063}, compatOnly: true},
	0xFF44: {runes: []rune{0x0064}, compatOnly: true},
	0xFF45: {runes: []rune{0x0065}, compatOnly: true},
	0xFF46: {runes: []rune{0x0066}, compatOnly: true},
	0xFF47: {runes: []rune{0x0067}, compatOnly: true},
	0xFF48: {runes: []rune{0x0068}, compatOnly: true},
	0xFF49: {runes: []rune{0x0069}, compatOnly: true},
	0xFF4A: {runes: []rune{0x006A}, compatOnly: true},
	0xFF4B: {runes: []rune{0x006B}, compatOnly: true},
	0xFF4C: {runes: []rune{0x006C}, compatOnly: true},
	0xFF4D: {runes: []rune{0x006D}, compatOnly: true},
	0xFF4E: {runes: []rune{0x006E}, compatOnly: true},
	0xFF4F: {runes: []rune{0x006F}, compatOnly: true},
	0xFF50: {runes: []rune{0x0070}, compatOnly: true},
	0xFF51: {runes: []rune{0x0071}, compatOnly: true},
	0xFF52: {runes: []rune{0x0072}, compatOnly: true},
	0xFF53: {runes: []rune{0x0073}, compatOnly: true},
	0xFF54: {runes: []rune{0x0074}, compatOnly: true},
	0xFF55: {runes: []rune{0x0075}, compatOnly: true},
	0xFF56: {runes: []rune{0x0076}, compatOnly: true},
	0xFF57: {runes: []rune{0x0077}, compatOnly: true},
	0xFF58: {runes: []rune{0x0078}, compatOnly: true},
	0xFF59: {runes: []rune{0x0079}, compatOnly: true},
	0xFF5A: {runes: []rune{0x007A}, compatOnly: true},
	
	0xFF10: {runes: []rune{0x0030}, compatOnly: true},
	0xFF11: {runes: []rune{0x0031}, compatOnly: true},
	0xFF12: {runes: []rune{0x0032}, compatOnly: true},
	0xFF13: {runes: []rune{0x0033}, compatOnly: true},
	0xFF14: {runes: []rune{0x0034}, compatOnly: true},
	0xFF15: {runes: []rune{0x0035}, compatOnly: true},
	0xFF16: {runes: []rune{0x0036}, compatOnly: true},
	0xFF17: {runes: []rune{0x0037}, compatOnly: true},
	0xFF18: {runes: []rune{0x0038}, compatOnly: true},
	0xFF19: {runes: []rune{0x0039}, compatOnly: true},
	
	0x2160: {runes: []rune{0x0049}, compatOnly: true},
	0x2161: {runes: []rune{0x0049, 0x0049}, compatOnly: true},
	0x2162: {runes: []rune{0x0049, 0x0049, 0x0049}, compatOnly: true},
	0x2163: {runes: []rune{0x0049, 0x0056}, compatOnly: true},
	0x2164: {runes: []rune{0x0056}, compatOnly: true},
	0x2165: {runes: []rune{0x0056, 0x0049}, compatOnly: true},
	0x2166: {runes: []rune{0x0056, 0x0049, 0x0049}, compatOnly: true},
	0x2167: {runes: []rune{0x0056, 0x0049, 0x0049, 0x0049}, compatOnly: true},
	0x2168: {runes: []rune{0x0049, 0x0058}, compatOnly: true},
	0x2169: {runes: []rune{0x0058}, compatOnly: true},
	0x216A: {runes: []rune{0x0058, 0x0049}, compatOnly: true},
	0x216B: {runes: []rune{0x0058, 0x0049, 0x0049}, compatOnly: true},
	0x216C: {runes: []rune{0x0058, 0x0049, 0x0049, 0x0049}, compatOnly: true},
	0x216D: {runes: []rune{0x0058, 0x004C}, compatOnly: true},
	0x216E: {runes: []rune{0x004C}, compatOnly: true},
	0x216F: {runes: []rune{0x0043}, compatOnly: true},
	
	0x2170: {runes: []rune{0x0069}, compatOnly: true},
	0x2171: {runes: []rune{0x0069, 0x0069}, compatOnly: true},
	0x2172: {runes: []rune{0x0069, 0x0069, 0x0069}, compatOnly: true},
	0x2173: {runes: []rune{0x0069, 0x0076}, compatOnly: true},
	0x2174: {runes: []rune{0x0076}, compatOnly: true},
	0x2175: {runes: []rune{0x0076, 0x0069}, compatOnly: true},
	0x2176: {runes: []rune{0x0076, 0x0069, 0x0069}, compatOnly: true},
	0x2177: {runes: []rune{0x0076, 0x0069, 0x0069, 0x0069}, compatOnly: true},
	0x2178: {runes: []rune{0x0069, 0x0078}, compatOnly: true},
	0x2179: {runes: []rune{0x0078}, compatOnly: true},
	0x217A: {runes: []rune{0x0078, 0x0069}, compatOnly: true},
	0x217B: {runes: []rune{0x0078, 0x0069, 0x0069}, compatOnly: true},
	0x217C: {runes: []rune{0x0078, 0x0069, 0x0069, 0x0069}, compatOnly: true},
	0x217D: {runes: []rune{0x0078, 0x006C}, compatOnly: true},
	0x217E: {runes: []rune{0x006C}, compatOnly: true},
	0x217F: {runes: []rune{0x0063}, compatOnly: true},
	
	0xFB01: {runes: []rune{0x0066, 0x0069}, compatOnly: true},
	0xFB02: {runes: []rune{0x0066, 0x006C}, compatOnly: true},
	0xFB03: {runes: []rune{0x0066, 0x0066, 0x0069}, compatOnly: true},
	0xFB04: {runes: []rune{0x0066, 0x0066, 0x006C}, compatOnly: true},
	0xFB00: {runes: []rune{0x0066, 0x0066}, compatOnly: true},
	
	0x00BD: {runes: []rune{0x0031, 0x2044, 0x0032}, compatOnly: true},
	0x00BC: {runes: []rune{0x0031, 0x2044, 0x0034}, compatOnly: true},
	0x00BE: {runes: []rune{0x0033, 0x2044, 0x0034}, compatOnly: true},
	0x2153: {runes: []rune{0x0031, 0x2044, 0x0033}, compatOnly: true},
	0x2154: {runes: []rune{0x0032, 0x2044, 0x0033}, compatOnly: true},
	0x2155: {runes: []rune{0x0031, 0x2044, 0x0035}, compatOnly: true},
	0x2156: {runes: []rune{0x0032, 0x2044, 0x0035}, compatOnly: true},
	0x2157: {runes: []rune{0x0033, 0x2044, 0x0035}, compatOnly: true},
	0x2158: {runes: []rune{0x0034, 0x2044, 0x0035}, compatOnly: true},
	0x2159: {runes: []rune{0x0031, 0x2044, 0x0036}, compatOnly: true},
	0x215A: {runes: []rune{0x0035, 0x2044, 0x0036}, compatOnly: true},
	0x215B: {runes: []rune{0x0031, 0x2044, 0x0038}, compatOnly: true},
	0x215C: {runes: []rune{0x0033, 0x2044, 0x0038}, compatOnly: true},
	0x215D: {runes: []rune{0x0035, 0x2044, 0x0038}, compatOnly: true},
	0x215E: {runes: []rune{0x0037, 0x2044, 0x0038}, compatOnly: true},
}

func getCCC(r rune) int {
	if ccc, ok := cccMap[r]; ok {
		return ccc
	}
	return 0
}

var cccMap = map[rune]int{
	0x0300: 230, 0x0301: 230, 0x0302: 230, 0x0303: 230,
	0x0304: 230, 0x0305: 230, 0x0306: 230, 0x0307: 230,
	0x0308: 230, 0x0309: 230, 0x030A: 230, 0x030B: 230,
	0x030C: 230, 0x030D: 230, 0x030E: 230, 0x030F: 230,
	
	0x0327: 202, 0x0328: 202, 0x0329: 202, 0x032A: 202,
	0x032B: 202, 0x032C: 202, 0x032D: 202, 0x032E: 202,
	0x032F: 202, 0x0330: 202, 0x0331: 202, 0x0332: 202,
	0x0333: 202, 0x0334: 202, 0x0335: 202, 0x0336: 202,
	0x0337: 202, 0x0338: 202, 0x0339: 202, 0x033A: 202,
	0x033B: 202, 0x033C: 202, 0x033D: 202, 0x033E: 202,
	0x033F: 202,
	
	0x0315: 220, 0x0316: 220, 0x0317: 220, 0x0318: 220,
	0x0319: 220, 0x031A: 220, 0x031B: 220, 0x031C: 220,
	0x031D: 220, 0x031E: 220, 0x031F: 220, 0x0320: 220,
	0x0321: 220, 0x0322: 220, 0x0323: 220, 0x0324: 220,
	0x0325: 220, 0x0326: 220,
}

func isStarter(r rune) bool {
	return getCCC(r) == 0
}

func reorderCombiningMarks(runes []rune) []rune {
	result := make([]rune, 0, len(runes))
	
	i := 0
	for i < len(runes) {
		result = append(result, runes[i])
		i++
		
		combiningStart := i
		for i < len(runes) && !isStarter(runes[i]) {
			i++
		}
		
		if combiningStart < i {
			combining := runes[combiningStart:i]
			sort.SliceStable(combining, func(a, b int) bool {
				return getCCC(combining[a]) < getCCC(combining[b])
			})
			result = append(result, combining...)
		}
	}
	
	return result
}

func isCombining(r rune) bool {
	return unicode.In(r, unicode.Mn, unicode.Mc, unicode.Me)
}
