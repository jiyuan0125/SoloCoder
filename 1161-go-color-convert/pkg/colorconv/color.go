package colorconv

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type RGB struct {
	R, G, B uint8
}

type HSL struct {
	H, S, L float64
}

type HSV struct {
	H, S, V float64
}

type CMYK struct {
	C, M, Y, K float64
}

func NewRGB(r, g, b int) (*RGB, error) {
	if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
		return nil, fmt.Errorf("RGB values must be in range [0, 255], got (%d, %d, %d)", r, g, b)
	}
	return &RGB{R: uint8(r), G: uint8(g), B: uint8(b)}, nil
}

func NewHSL(h, s, l float64) (*HSL, error) {
	if h < 0 || h > 360 {
		return nil, fmt.Errorf("HSL hue must be in range [0, 360], got %f", h)
	}
	if s < 0 || s > 1 {
		return nil, fmt.Errorf("HSL saturation must be in range [0, 1], got %f", s)
	}
	if l < 0 || l > 1 {
		return nil, fmt.Errorf("HSL lightness must be in range [0, 1], got %f", l)
	}
	return &HSL{H: h, S: s, L: l}, nil
}

func NewHSV(h, s, v float64) (*HSV, error) {
	if h < 0 || h > 360 {
		return nil, fmt.Errorf("HSV hue must be in range [0, 360], got %f", h)
	}
	if s < 0 || s > 1 {
		return nil, fmt.Errorf("HSV saturation must be in range [0, 1], got %f", s)
	}
	if v < 0 || v > 1 {
		return nil, fmt.Errorf("HSV value must be in range [0, 1], got %f", v)
	}
	return &HSV{H: h, S: s, V: v}, nil
}

func NewCMYK(c, m, y, k float64) (*CMYK, error) {
	if c < 0 || c > 1 {
		return nil, fmt.Errorf("CMYK cyan must be in range [0, 1], got %f", c)
	}
	if m < 0 || m > 1 {
		return nil, fmt.Errorf("CMYK magenta must be in range [0, 1], got %f", m)
	}
	if y < 0 || y > 1 {
		return nil, fmt.Errorf("CMYK yellow must be in range [0, 1], got %f", y)
	}
	if k < 0 || k > 1 {
		return nil, fmt.Errorf("CMYK black must be in range [0, 1], got %f", k)
	}
	return &CMYK{C: c, M: m, Y: y, K: k}, nil
}

var hexRegex = regexp.MustCompile(`^#?([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

func ParseHex(hex string) (*RGB, error) {
	match := hexRegex.FindStringSubmatch(hex)
	if match == nil {
		return nil, fmt.Errorf("invalid hex color format: %s", hex)
	}
	hexStr := match[1]
	if len(hexStr) == 3 {
		hexStr = strings.Repeat(string(hexStr[0]), 2) +
			strings.Repeat(string(hexStr[1]), 2) +
			strings.Repeat(string(hexStr[2]), 2)
	}
	r, _ := strconv.ParseUint(hexStr[0:2], 16, 8)
	g, _ := strconv.ParseUint(hexStr[2:4], 16, 8)
	b, _ := strconv.ParseUint(hexStr[4:6], 16, 8)
	return &RGB{R: uint8(r), G: uint8(g), B: uint8(b)}, nil
}

func (rgb *RGB) ToHex() string {
	return fmt.Sprintf("#%02X%02X%02X", rgb.R, rgb.G, rgb.B)
}

func (rgb *RGB) ToHSL() *HSL {
	r := float64(rgb.R) / 255
	g := float64(rgb.G) / 255
	b := float64(rgb.B) / 255

	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	h, s := 0.0, 0.0
	l := (max + min) / 2

	if max == min {
		h = 0
		s = 0
	} else {
		d := max - min
		if l > 0.5 {
			s = d / (2 - max - min)
		} else {
			s = d / (max + min)
		}

		switch max {
		case r:
			h = (g - b) / d
			if g < b {
				h += 6
			}
		case g:
			h = (b-r)/d + 2
		case b:
			h = (r-g)/d + 4
		}
		h /= 6
	}

	return &HSL{
		H: round(h * 360),
		S: round(s),
		L: round(l),
	}
}

func (rgb *RGB) ToHSV() *HSV {
	r := float64(rgb.R) / 255
	g := float64(rgb.G) / 255
	b := float64(rgb.B) / 255

	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	h, s, v := 0.0, 0.0, max
	d := max - min

	if max != 0 {
		s = d / max
	}

	if max == min {
		h = 0
	} else {
		switch max {
		case r:
			h = (g - b) / d
			if g < b {
				h += 6
			}
		case g:
			h = (b-r)/d + 2
		case b:
			h = (r-g)/d + 4
		}
		h /= 6
	}

	return &HSV{
		H: round(h * 360),
		S: round(s),
		V: round(v),
	}
}

func (rgb *RGB) ToCMYK() *CMYK {
	r := float64(rgb.R) / 255
	g := float64(rgb.G) / 255
	b := float64(rgb.B) / 255

	k := 1 - math.Max(r, math.Max(g, b))
	if k == 1 {
		return &CMYK{C: 0, M: 0, Y: 0, K: 1}
	}

	c := (1 - r - k) / (1 - k)
	m := (1 - g - k) / (1 - k)
	y := (1 - b - k) / (1 - k)

	return &CMYK{
		C: round(c),
		M: round(m),
		Y: round(y),
		K: round(k),
	}
}

func (hsl *HSL) ToRGB() *RGB {
	h := hsl.H / 360
	s := hsl.S
	l := hsl.L

	var r, g, b float64

	if s == 0 {
		r = l
		g = l
		b = l
	} else {
		q := 0.0
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q
		r = hueToRGB(p, q, h+1.0/3)
		g = hueToRGB(p, q, h)
		b = hueToRGB(p, q, h-1.0/3)
	}

	return &RGB{
		R: uint8(math.Round(r * 255)),
		G: uint8(math.Round(g * 255)),
		B: uint8(math.Round(b * 255)),
	}
}

func (hsl *HSL) ToHSV() *HSV {
	l := hsl.L
	s := hsl.S

	v := l + s*math.Min(l, 1-l)
	var newS float64
	if v == 0 {
		newS = 0
	} else {
		newS = 2 * (1 - l/v)
	}

	return &HSV{
		H: hsl.H,
		S: round(newS),
		V: round(v),
	}
}

func (hsl *HSL) ToCMYK() *CMYK {
	return hsl.ToRGB().ToCMYK()
}

func (hsv *HSV) ToRGB() *RGB {
	h := hsv.H / 360
	s := hsv.S
	v := hsv.V

	i := math.Floor(h * 6)
	f := h*6 - i
	p := v * (1 - s)
	q := v * (1 - f*s)
	t := v * (1 - (1-f)*s)

	var r, g, b float64
	switch int(i) % 6 {
	case 0:
		r, g, b = v, t, p
	case 1:
		r, g, b = q, v, p
	case 2:
		r, g, b = p, v, t
	case 3:
		r, g, b = p, q, v
	case 4:
		r, g, b = t, p, v
	case 5:
		r, g, b = v, p, q
	}

	return &RGB{
		R: uint8(math.Round(r * 255)),
		G: uint8(math.Round(g * 255)),
		B: uint8(math.Round(b * 255)),
	}
}

func (hsv *HSV) ToHSL() *HSL {
	v := hsv.V
	s := hsv.S

	l := v * (1 - s/2)
	var newS float64
	if l == 0 || l == 1 {
		newS = 0
	} else {
		newS = (v - l) / math.Min(l, 1-l)
	}

	return &HSL{
		H: hsv.H,
		S: round(newS),
		L: round(l),
	}
}

func (hsv *HSV) ToCMYK() *CMYK {
	return hsv.ToRGB().ToCMYK()
}

func (cmyk *CMYK) ToRGB() *RGB {
	r := (1 - cmyk.C) * (1 - cmyk.K) * 255
	g := (1 - cmyk.M) * (1 - cmyk.K) * 255
	b := (1 - cmyk.Y) * (1 - cmyk.K) * 255

	r = clamp0To255(r)
	g = clamp0To255(g)
	b = clamp0To255(b)

	return &RGB{
		R: uint8(math.Round(r)),
		G: uint8(math.Round(g)),
		B: uint8(math.Round(b)),
	}
}

func (cmyk *CMYK) ToHSL() *HSL {
	return cmyk.ToRGB().ToHSL()
}

func (cmyk *CMYK) ToHSV() *HSV {
	return cmyk.ToRGB().ToHSV()
}

func (rgb *RGB) RelativeLuminance() float64 {
	r := linearize(float64(rgb.R) / 255)
	g := linearize(float64(rgb.G) / 255)
	b := linearize(float64(rgb.B) / 255)
	return round(0.2126*r + 0.7152*g + 0.0722*b)
}

func RelativeLuminanceFromHex(hex string) (float64, error) {
	rgb, err := ParseHex(hex)
	if err != nil {
		return 0, err
	}
	return rgb.RelativeLuminance(), nil
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t++
	}
	if t > 1 {
		t--
	}
	if t < 1.0/6 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2 {
		return q
	}
	if t < 2.0/3 {
		return p + (q-p)*(2.0/3-t)*6
	}
	return p
}

func linearize(x float64) float64 {
	if x <= 0.04045 {
		return x / 12.92
	}
	return math.Pow((x+0.055)/1.055, 2.4)
}

func clamp0To255(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 255 {
		return 255
	}
	return x
}

func round(x float64) float64 {
	return math.Round(x*10000) / 10000
}
