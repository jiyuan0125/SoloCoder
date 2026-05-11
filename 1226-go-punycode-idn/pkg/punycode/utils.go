package punycode

const (
	base        = 36
	tmin        = 1
	tmax        = 26
	skew        = 38
	damp        = 700
	initialBias = 72
	initialN    = 128
	delimiter   = '-'
	prefix      = "xn--"
)

func digitToValue(digit byte) (int, error) {
	if digit >= '0' && digit <= '9' {
		return int(digit - '0' + 26), nil
	}
	if digit >= 'a' && digit <= 'z' {
		return int(digit - 'a'), nil
	}
	if digit >= 'A' && digit <= 'Z' {
		return int(digit - 'A'), nil
	}
	return 0, ErrInvalidPunycode
}

func valueToDigit(value int) byte {
	if value < 26 {
		return byte('a' + value)
	}
	return byte('0' + value - 26)
}

func adapt(delta int, numPoints int, firstTime bool) int {
	if firstTime {
		delta = delta / damp
	} else {
		delta = delta / 2
	}
	delta = delta + delta/numPoints
	k := 0
	for delta > (base-tmin)*tmax/2 {
		delta = delta / (base - tmin)
		k = k + base
	}
	return k + (base-tmin+1)*delta/(delta+skew)
}
