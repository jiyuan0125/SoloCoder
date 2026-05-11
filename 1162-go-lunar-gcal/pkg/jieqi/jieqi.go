package jieqi

import (
	"fmt"
	"math"
	"time"
)

const (
	MinYear = 1900
	MaxYear = 2100
	deg2rad = math.Pi / 180.0
)

var jieqiNames = []string{
	"小寒", "大寒", "立春", "雨水", "惊蛰", "春分",
	"清明", "谷雨", "立夏", "小满", "芒种", "夏至",
	"小暑", "大暑", "立秋", "处暑", "白露", "秋分",
	"寒露", "霜降", "立冬", "小雪", "大雪", "冬至",
}

func GetJieqiNames() []string {
	return jieqiNames
}

func normAngle(angle float64) float64 {
	angle = math.Mod(angle, 360.0)
	if angle < 0 {
		angle += 360.0
	}
	return angle
}

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

func timeFromJulianDay(jd float64) time.Time {
	jd += 0.5
	Z := math.Floor(jd)
	F := jd - Z
	var A float64
	if Z < 2299161 {
		A = Z
	} else {
		alpha := math.Floor((Z - 1867216.25) / 36524.25)
		A = Z + 1 + alpha - math.Floor(alpha/4.0)
	}
	B := A + 1524
	C := math.Floor((B - 122.1) / 365.25)
	D := math.Floor(365.25 * C)
	E := math.Floor((B - D) / 30.6001)
	day := B - D - math.Floor(30.6001*E) + F
	var month, year float64
	if E < 14 {
		month = E - 1
	} else {
		month = E - 13
	}
	if month > 2 {
		year = C - 4716
	} else {
		year = C - 4715
	}
	dayFrac := day - math.Floor(day)
	day = math.Floor(day)
	hour := dayFrac * 24
	minute := (hour - math.Floor(hour)) * 60
	second := (minute - math.Floor(minute)) * 60

	return time.Date(int(year), time.Month(int(month)), int(day),
		int(math.Floor(hour)), int(math.Floor(minute)), int(math.Floor(second)), 0, time.UTC)
}

func sunLongitude(jd float64) float64 {
	T := (jd - 2451545.0) / 36525.0
	T2 := T * T

	L0 := 280.46645 + 36000.76983*T + 0.0003032*T2
	M := 357.52910 + 35999.05030*T - 0.0001559*T2

	L0 = normAngle(L0)
	M = normAngle(M)

	C := (1.914600 - 0.004817*T - 0.000014*T2)*math.Sin(M*deg2rad)
	C += (0.019993 - 0.000101*T) * math.Sin(2*M*deg2rad)
	C += 0.000290 * math.Sin(3*M*deg2rad)

	omega := 125.04 - 1934.136*T
	lambda := L0 + C

	deltaPsi := -0.00478 * math.Sin(omega*deg2rad)
	lambda += deltaPsi

	return normAngle(lambda)
}

func findJieqiTime(year, termIndex int) time.Time {
	targetLon := 285.0 + float64(termIndex)*15.0
	targetLon = math.Mod(targetLon, 360.0)
	if targetLon < 0 {
		targetLon += 360.0
	}

	estMonth := termIndex/2 + 1
	estDay := 15
	if termIndex == 0 {
		estDay = 5
	} else if termIndex == 1 {
		estDay = 20
	} else if termIndex == 2 {
		estDay = 4
	} else if termIndex == 3 {
		estDay = 19
	} else if termIndex == 4 {
		estDay = 5
	} else if termIndex == 5 {
		estDay = 20
	} else if termIndex == 6 {
		estDay = 4
	} else if termIndex == 7 {
		estDay = 20
	} else if termIndex == 8 {
		estDay = 5
	} else if termIndex == 9 {
		estDay = 21
	} else if termIndex == 10 {
		estDay = 5
	} else if termIndex == 11 {
		estDay = 21
	} else if termIndex == 12 {
		estDay = 7
	} else if termIndex == 13 {
		estDay = 22
	} else if termIndex == 14 {
		estDay = 7
	} else if termIndex == 15 {
		estDay = 23
	} else if termIndex == 16 {
		estDay = 7
	} else if termIndex == 17 {
		estDay = 22
	} else if termIndex == 18 {
		estDay = 8
	} else if termIndex == 19 {
		estDay = 23
	} else if termIndex == 20 {
		estDay = 7
	} else if termIndex == 21 {
		estDay = 22
	} else if termIndex == 22 {
		estDay = 7
	} else if termIndex == 23 {
		estDay = 21
	}

	t := time.Date(year, time.Month(estMonth), estDay, 12, 0, 0, 0, time.UTC)
	jd := julianDay(t)

	for iter := 0; iter < 100; iter++ {
		lon := sunLongitude(jd)
		deltaLon := normAngle(targetLon - lon)
		if deltaLon > 180 {
			deltaLon -= 360
		}
		if deltaLon < -180 {
			deltaLon += 360
		}

		if math.Abs(deltaLon) < 0.0001 {
			break
		}

		eps := 0.00001
		lon0 := sunLongitude(jd - eps)
		lon1 := sunLongitude(jd + eps)
		rate := (lon1 - lon0) / (2 * eps)
		if math.Abs(rate) < 0.1 {
			rate = 1
		}

		deltaJD := deltaLon / rate
		if math.Abs(deltaJD) > 10 {
			deltaJD = math.Copysign(10, deltaJD)
		}
		jd += deltaJD
	}

	return timeFromJulianDay(jd).In(time.FixedZone("CST", 8*3600))
}

type Jieqi struct {
	Name string
	Time time.Time
}

func GetJieqi(year int) ([]Jieqi, error) {
	if year < MinYear || year > MaxYear {
		return nil, fmt.Errorf("年份超出范围（%d-%d）", MinYear, MaxYear)
	}

	result := make([]Jieqi, 24)

	for i := 0; i < 24; i++ {
		result[i] = Jieqi{
			Name: jieqiNames[i],
			Time: findJieqiTime(year, i),
		}
	}

	return result, nil
}
