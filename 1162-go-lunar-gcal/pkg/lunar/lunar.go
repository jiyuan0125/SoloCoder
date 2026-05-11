package lunar

import (
	"fmt"
	"time"
)

type LunarDate struct {
	Year        int
	Month       int
	Day         int
	IsLeapMonth bool
}

func (l *LunarDate) String() string {
	ganzhi := GetTiangan(l.Year) + GetDizhi(l.Year)
	month := GetMonthName(l.Month, l.IsLeapMonth)
	day := GetDayName(l.Day)
	sx := GetShengxiao(l.Year)
	return fmt.Sprintf("%s年（%s） %s %s", ganzhi, sx, month, day)
}

func (l *LunarDate) YearGanzhi() string {
	return GetTiangan(l.Year) + GetDizhi(l.Year)
}

func (l *LunarDate) YearShengxiao() string {
	return GetShengxiao(l.Year)
}

func (l *LunarDate) MonthName() string {
	return GetMonthName(l.Month, l.IsLeapMonth)
}

func (l *LunarDate) DayName() string {
	return GetDayName(l.Day)
}

func yearIndex(year int) int {
	return year - MinYear
}

func getLeapMonth(year int) int {
	if year < MinYear || year > MaxYear {
		return 0
	}
	idx := yearIndex(year)
	return int(lunarInfo[idx] & 0xf)
}

func getLeapDays(year int) int {
	if getLeapMonth(year) == 0 {
		return 0
	}
	idx := yearIndex(year)
	if (lunarInfo[idx] & 0x10000) != 0 {
		return 30
	}
	return 29
}

func getMonthDays(year, month int) int {
	if year < MinYear || year > MaxYear || month < 1 || month > 12 {
		return 0
	}
	idx := yearIndex(year)
	bit := uint(16 - month)
	if (lunarInfo[idx] & (1 << bit)) != 0 {
		return 30
	}
	return 29
}

func getYearDays(year int) int {
	sum := 348
	idx := yearIndex(year)
	for i := uint32(0x8000); i > 0x8; i >>= 1 {
		if (lunarInfo[idx] & i) != 0 {
			sum += 1
		}
	}
	return sum + getLeapDays(year)
}

func SolarToLunar(t time.Time) (*LunarDate, error) {
	t = t.In(time.FixedZone("CST", 8*3600))
	year := t.Year()
	if year < MinYear || year > MaxYear {
		return nil, fmt.Errorf("年份超出范围（%d-%d）", MinYear, MaxYear)
	}

	baseDate := time.Date(1900, 1, 31, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
	offset := int(t.Sub(baseDate).Hours() / 24)

	if offset < 0 {
		return nil, fmt.Errorf("日期超出范围")
	}

	temp := 0
	lyear := MinYear
	for ; lyear <= MaxYear && offset > 0; lyear++ {
		temp = getYearDays(lyear)
		if offset < temp {
			break
		}
		offset -= temp
	}

	if lyear > MaxYear {
		return nil, fmt.Errorf("年份超出范围（%d-%d）", MinYear, MaxYear)
	}

	leap := getLeapMonth(lyear)
	isLeap := false
	lmonth := 1

	for lmonth <= 12 {
		var monthDays int
		if isLeap {
			monthDays = getLeapDays(lyear)
		} else {
			monthDays = getMonthDays(lyear, lmonth)
		}

		if offset < monthDays {
			break
		}
		offset -= monthDays

		if leap > 0 && lmonth == leap && !isLeap {
			isLeap = true
		} else {
			if isLeap {
				isLeap = false
			}
			lmonth++
		}
	}

	lday := offset + 1
	return &LunarDate{
		Year:        lyear,
		Month:       lmonth,
		Day:         lday,
		IsLeapMonth: isLeap,
	}, nil
}

func LunarToSolar(lunar LunarDate) (time.Time, error) {
	if lunar.Year < MinYear || lunar.Year > MaxYear {
		return time.Time{}, fmt.Errorf("年份超出范围（%d-%d）", MinYear, MaxYear)
	}
	if lunar.Month < 1 || lunar.Month > 12 {
		return time.Time{}, fmt.Errorf("月份超出范围（1-12）")
	}
	if lunar.Day < 1 || lunar.Day > 30 {
		return time.Time{}, fmt.Errorf("日期超出范围（1-30）")
	}

	leapMonth := getLeapMonth(lunar.Year)
	if lunar.IsLeapMonth {
		if leapMonth == 0 {
			return time.Time{}, fmt.Errorf("农历 %d 年没有闰月", lunar.Year)
		}
		if lunar.Month != leapMonth {
			return time.Time{}, fmt.Errorf("农历 %d 年的闰月是闰%d月，不是闰%d月", lunar.Year, leapMonth, lunar.Month)
		}
		leapDays := getLeapDays(lunar.Year)
		if lunar.Day > leapDays {
			return time.Time{}, fmt.Errorf("农历 %d 年闰%d月只有 %d 天", lunar.Year, lunar.Month, leapDays)
		}
	} else {
		monthDays := getMonthDays(lunar.Year, lunar.Month)
		if lunar.Day > monthDays {
			return time.Time{}, fmt.Errorf("农历 %d 年%d月只有 %d 天", lunar.Year, lunar.Month, monthDays)
		}
	}

	offset := 0
	for y := MinYear; y < lunar.Year; y++ {
		offset += getYearDays(y)
	}

	leap := getLeapMonth(lunar.Year)
	isLeap := false
	for m := 1; m < lunar.Month || (lunar.IsLeapMonth && m == lunar.Month && !isLeap); {
		if isLeap {
			offset += getLeapDays(lunar.Year)
		} else {
			offset += getMonthDays(lunar.Year, m)
		}

		if leap > 0 && m == leap && !isLeap {
			isLeap = true
		} else {
			if isLeap {
				isLeap = false
			}
			m++
		}
	}

	offset += lunar.Day - 1

	baseDate := time.Date(1900, 1, 31, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
	return baseDate.AddDate(0, 0, offset), nil
}
