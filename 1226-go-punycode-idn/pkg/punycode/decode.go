package punycode

import (
	"strings"
)

var ErrInvalidPunycode = &PunycodeError{"invalid punycode"}

type PunycodeError struct {
	Msg string
}

func (e *PunycodeError) Error() string {
	return e.Msg
}

func Decode(input string) (string, error) {
	if !strings.HasPrefix(input, prefix) {
		return input, nil
	}

	encoded := input[len(prefix):]
	lastDash := strings.LastIndexByte(encoded, delimiter)

	output := make([]rune, 0, len(encoded))

	codeStart := 0
	if lastDash != -1 {
		for i := 0; i < lastDash; i++ {
			output = append(output, rune(encoded[i]))
		}
		codeStart = lastDash + 1
	}

	n := rune(initialN)
	i := 0
	bias := initialBias
	inLen := len(encoded)

	for codeStart < inLen {
		oldi := i
		w := 1

		for k := base; ; k += base {
			if codeStart >= inLen {
				return "", ErrInvalidPunycode
			}

			digit, err := digitToValue(encoded[codeStart])
			if err != nil {
				return "", err
			}
			codeStart++

			i = i + digit*w

			var t int
			if k <= bias {
				t = tmin
			} else if k >= bias+tmax {
				t = tmax
			} else {
				t = k - bias
			}

			if digit < t {
				break
			}

			w = w * (base - t)
		}

		outLen := len(output) + 1
		bias = adapt(i-oldi, outLen, oldi == 0)
		n = n + rune(i/outLen)
		i = i % outLen

		output = append(output, 0)
		copy(output[i+1:], output[i:])
		output[i] = rune(n)
		i++
	}

	return string(output), nil
}
