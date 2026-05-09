package uninorm

type composer struct{}

func newComposer() *composer {
	return &composer{}
}

func (c *composer) composeRunes(runes []rune) []rune {
	if len(runes) == 0 {
		return runes
	}
	
	result := make([]rune, 0, len(runes))
	
	i := 0
	for i < len(runes) {
		base := runes[i]
		i++
		
		combining := make([]rune, 0)
		for i < len(runes) && !isStarter(runes[i]) {
			combining = append(combining, runes[i])
			i++
		}
		
		if len(combining) == 0 {
			result = append(result, base)
			continue
		}
		
		composed := base
		remainingCombining := make([]rune, 0, len(combining))
		
		for _, cm := range combining {
			if composedValue, ok := composeMap[[2]rune{composed, cm}]; ok {
				composed = composedValue
			} else {
				remainingCombining = append(remainingCombining, cm)
			}
		}
		
		result = append(result, composed)
		result = append(result, remainingCombining...)
	}
	
	return result
}

var composeMap = map[[2]rune]rune{
	{0x0045, 0x0301}: 0x00C9,
	{0x0065, 0x0301}: 0x00E9,
	{0x0045, 0x0300}: 0x00C8,
	{0x0065, 0x0300}: 0x00E8,
	{0x0045, 0x0302}: 0x00CA,
	{0x0065, 0x0302}: 0x00EA,
	{0x0045, 0x0308}: 0x00CB,
	{0x0065, 0x0308}: 0x00EB,
	
	{0x0041, 0x0301}: 0x00C1,
	{0x0061, 0x0301}: 0x00E1,
	{0x0041, 0x0300}: 0x00C0,
	{0x0061, 0x0300}: 0x00E0,
	{0x0041, 0x0302}: 0x00C2,
	{0x0061, 0x0302}: 0x00E2,
	{0x0041, 0x0308}: 0x00C4,
	{0x0061, 0x0308}: 0x00E4,
	
	{0x0049, 0x0301}: 0x00CD,
	{0x0069, 0x0301}: 0x00ED,
	{0x0049, 0x0300}: 0x00CC,
	{0x0069, 0x0300}: 0x00EC,
	{0x0049, 0x0302}: 0x00CE,
	{0x0069, 0x0302}: 0x00EE,
	{0x0049, 0x0308}: 0x00CF,
	{0x0069, 0x0308}: 0x00EF,
	
	{0x004F, 0x0301}: 0x00D3,
	{0x006F, 0x0301}: 0x00F3,
	{0x004F, 0x0300}: 0x00D2,
	{0x006F, 0x0300}: 0x00F2,
	{0x004F, 0x0302}: 0x00D4,
	{0x006F, 0x0302}: 0x00F4,
	{0x004F, 0x0308}: 0x00D6,
	{0x006F, 0x0308}: 0x00F6,
	
	{0x0055, 0x0301}: 0x00DA,
	{0x0075, 0x0301}: 0x00FA,
	{0x0055, 0x0300}: 0x00D9,
	{0x0075, 0x0300}: 0x00F9,
	{0x0055, 0x0302}: 0x00DB,
	{0x0075, 0x0302}: 0x00FB,
	{0x0055, 0x0308}: 0x00DC,
	{0x0075, 0x0308}: 0x00FC,
	
	{0x0043, 0x0327}: 0x00C7,
	{0x0063, 0x0327}: 0x00E7,
	{0x004E, 0x0303}: 0x00D1,
	{0x006E, 0x0303}: 0x00F1,
}

func customNormalize(input string, form Form) string {
	if input == "" {
		return ""
	}
	
	var compat bool
	var doCompose bool
	
	switch form {
	case NFD:
		compat = false
		doCompose = false
	case NFC:
		compat = false
		doCompose = true
	case NFKD:
		compat = true
		doCompose = false
	case NFKC:
		compat = true
		doCompose = true
	default:
		return input
	}
	
	decomposer := newDecomposer(compat)
	decomposed := decomposer.decomposeString(input)
	
	reordered := reorderCombiningMarks(decomposed)
	
	if doCompose {
		composer := newComposer()
		composed := composer.composeRunes(reordered)
		return string(composed)
	}
	
	return string(reordered)
}
