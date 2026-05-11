package main

import (
	"fmt"
)

const (
	base        = 36
	tmin        = 1
	tmax        = 26
	skew        = 38
	damp        = 700
	initialBias = 72
	initialN    = 128
	delimiter   = '-'
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
	return 0, fmt.Errorf("invalid digit")
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

func encodeDebug(input string) (string, error) {
	inputRunes := []rune(input)
	n := rune(initialN)
	delta := 0
	bias := initialBias
	h := 0
	b := 0

	fmt.Printf("输入: %q (码点: ", input)
	for i, r := range inputRunes {
		if i > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("U+%04X", r)
	}
	fmt.Println(")")

	for _, r := range inputRunes {
		if r < 0x80 {
			h++
		}
	}
	b = h

	fmt.Printf("初始: h=%d, b=%d, n=U+%04X, bias=%d\n", h, b, n, bias)

	output := make([]byte, 0, len(inputRunes)*2)

	for _, r := range inputRunes {
		if r < 0x80 {
			output = append(output, byte(r))
		}
	}

	if b > 0 {
		output = append(output, delimiter)
	}

	fmt.Printf("ASCII前缀输出: %q\n", string(output))

	step := 0
	for h < len(inputRunes) {
		step++
		m := rune(0x10FFFF)
		for _, r := range inputRunes {
			if r >= n && r < m {
				m = r
			}
		}

		fmt.Printf("\n[步骤%d] 找到最小非处理码点: m=U+%04X\n", step, m)
		fmt.Printf("  delta += (%d - %d) * (%d + 1) = %d + %d = ", m, n, h, delta, int(m-n)*(h+1))
		delta = delta + int(m-n)*(h+1)
		fmt.Printf("%d\n", delta)
		n = m

		for charIdx, r := range inputRunes {
			if r < n {
				delta++
				fmt.Printf("  字符#%d (%c) < n, delta++ -> %d\n", charIdx, r, delta)
			}
			if r == n {
				fmt.Printf("\n  处理字符#%d (%c=U+%04X): q=delta=%d\n", charIdx, r, r, delta)
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
					fmt.Printf("    k=%d, t=%d, q=%d", k, t, q)
					if q < t {
						fmt.Printf(" < t -> 退出内循环\n")
						break
					}
					digitVal := t + (q-t)%(base-t)
					digit := valueToDigit(digitVal)
					fmt.Printf(" >= t -> digit=%d(%c), q=(%d-%d)/%d=", digitVal, digit, q, t, base-t)
					output = append(output, digit)
					q = (q - t) / (base - t)
					fmt.Printf("%d\n", q)
				}
				lastDigit := valueToDigit(q)
				fmt.Printf("    最终digit: %d(%c)\n", q, lastDigit)
				output = append(output, lastDigit)

				newBias := adapt(delta, h+1, h == b)
				fmt.Printf("  adapt(delta=%d, h+1=%d, firstTime=%v): bias=%d -> %d\n", delta, h+1, h == b, bias, newBias)
				bias = newBias
				delta = 0
				h++
				fmt.Printf("  h=%d, n=U+%04X\n", h, n)
			}
		}
		delta++
		fmt.Printf("\n[步骤%d结束] delta++, n++ -> delta=%d, n=U+%04X\n", step, delta, n+1)
		n++
	}

	result := "xn--" + string(output)
	if b == len(inputRunes) {
		result = input
	}

	fmt.Printf("\n=== 最终编码结果: %q ===\n", result)
	return result, nil
}

func main() {
	tests := []string{"例", "例子", "日本", "日本語"}
	for _, t := range tests {
		fmt.Println("========================================")
		fmt.Printf("开始编码: %q\n", t)
		fmt.Println("========================================")
		encodeDebug(t)
		fmt.Println()
	}
}
