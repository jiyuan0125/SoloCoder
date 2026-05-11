package hilbert

import (
	"errors"
)

const maxOrder = 31

type Range struct {
	Min, Max uint64
}

type Point struct {
	X, Y uint64
}

func validateOrder(n int) error {
	if n < 0 || n > maxOrder {
		return errors.New("order must be between 0 and 31")
	}
	return nil
}

func validatePoint(n int, p Point) error {
	size := uint64(1) << n
	if p.X >= size || p.Y >= size {
		return errors.New("point coordinates out of range")
	}
	return nil
}

func validateIndex(n int, d uint64) error {
	maxD := (uint64(1) << (2 * n)) - 1
	if d > maxD {
		return errors.New("index out of range")
	}
	return nil
}

func PointToIndex(n int, p Point) (uint64, error) {
	if err := validateOrder(n); err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, nil
	}
	if err := validatePoint(n, p); err != nil {
		return 0, err
	}

	x, y := p.X, p.Y
	var d uint64
	var s uint64

	for s = 1 << (n - 1); s > 0; s >>= 1 {
		rx := (x & s) > 0
		ry := (y & s) > 0
		d += s * s * ((3 * boolToInt(rx)) ^ boolToInt(ry))
		x, y = rotate(s, x, y, rx, ry)
	}

	return d, nil
}

func IndexToPoint(n int, d uint64) (Point, error) {
	if err := validateOrder(n); err != nil {
		return Point{}, err
	}
	if n == 0 {
		return Point{0, 0}, nil
	}
	if err := validateIndex(n, d); err != nil {
		return Point{}, err
	}

	var x, y uint64
	var t uint64 = d
	var s uint64

	for s = 1; s < (1 << n); s <<= 1 {
		rx := 1 & (t >> 1)
		ry := 1 & (t ^ rx)
		x, y = rotate(s, x, y, rx == 1, ry == 1)
		x += s * rx
		y += s * ry
		t >>= 2
	}

	return Point{x, y}, nil
}

func rotate(n uint64, x, y uint64, rx, ry bool) (uint64, uint64) {
	if !ry {
		if rx {
			x = n - 1 - x
			y = n - 1 - y
		}
		x, y = y, x
	}
	return x, y
}

func boolToInt(b bool) uint64 {
	if b {
		return 1
	}
	return 0
}
