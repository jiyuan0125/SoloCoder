//go:build ignore

package main

import (
	"fmt"
	"math"
	"time"
)

func julianDay(t time.Time) float64 {
	utc := t.UTC()
	y := utc.Year()
	m := int(utc.Month())
	d := float64(utc.Day()) +
		float64(utc.Hour())/24.0 +
		float64(utc.Minute())/1440.0 +
		float64(utc.Second())/86400.0

	if m <= 2 {
		y--
		m += 12
	}
	A := math.Floor(float64(y) / 100.0)
	B := 2.0 - A + math.Floor(A/4.0)
	return math.Floor(365.25*float64(y+4716)) + math.Floor(30.6001*float64(m+1)) + d + B - 1524.5
}

func main() {
	t := time.Date(2024, 2, 4, 16, 26, 0, 0, time.UTC)
	jd := julianDay(t)
	fmt.Printf("2024-02-04 16:26 UTC -> JD: %f\n", jd)

	t2 := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
	jd2 := julianDay(t2)
	fmt.Printf("2000-01-01 12:00 UTC -> JD: %f (expected: ~2451545.0)\n", jd2)

	T := (jd - 2451545.0) / 36525.0
	fmt.Printf("T (Julian centuries from J2000): %f\n", T)

	L0 := 280.46645 + 36000.76983*T + 0.0003032*T*T
	fmt.Printf("L0 (mean longitude): %f degrees\n", math.Mod(L0, 360.0))

	M := 357.52910 + 35999.05030*T - 0.0001559*T*T
	fmt.Printf("M (mean anomaly): %f degrees\n", math.Mod(M, 360.0))
}
