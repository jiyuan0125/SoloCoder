package punycode

func toLowerASCII(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}

func Encode(input string) (string, error) {
	inputRunes := []rune(input)
	n := rune(initialN)
	delta := 0
	bias := initialBias
	h := 0
	b := 0

	for _, r := range inputRunes {
		if r < 0x80 {
			h++
		}
	}
	b = h

	output := make([]byte, 0, len(inputRunes)*2)

	for _, r := range inputRunes {
		if r < 0x80 {
			output = append(output, byte(toLowerASCII(r)))
		}
	}

	if b > 0 {
		output = append(output, delimiter)
	}

	for h < len(inputRunes) {
		m := rune(0x10FFFF)
		for _, r := range inputRunes {
			if r >= n && r < m {
				m = r
			}
		}

		delta = delta + int(m-n)*(h+1)
		n = m

		for _, r := range inputRunes {
			if r < n {
				delta++
			}
			if r == n {
				q := delta
				for k := base; ; k += base {
					var t int
					if k <= bias {
						t = tmin
					} else if k >= bias+tmax {
						t = tmax
					} else {
						t = k - bias
					}
					if q < t {
						break
					}
					output = append(output, valueToDigit(t+(q-t)%(base-t)))
					q = (q - t) / (base - t)
				}
				output = append(output, valueToDigit(q))
				bias = adapt(delta, h+1, h == b)
				delta = 0
				h++
			}
		}
		delta++
		n++
	}

	if b == len(inputRunes) {
		result := make([]rune, len(inputRunes))
		for i, r := range inputRunes {
			result[i] = toLowerASCII(r)
		}
		return string(result), nil
	}

	return prefix + string(output), nil
}
