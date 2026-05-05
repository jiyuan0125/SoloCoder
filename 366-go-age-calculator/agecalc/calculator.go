package agecalc

import (
	"time"
)

func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func daysInMonth(year int, month time.Month) int {
	switch month {
	case time.April, time.June, time.September, time.November:
		return 30
	case time.February:
		if isLeapYear(year) {
			return 29
		}
		return 28
	default:
		return 31
	}
}

// adjustBirthdayForLeapYear 调整闰年2月29日出生的人在非闰年的生日
// 选择：非闰年时生日算2月28日
// 理由：
// 1. 法律上常见的处理方式（如中国法律规定，2月29日出生的人，在非闰年的生日视为2月28日）
// 2. 对于年龄计算（如满18岁、退休年龄等），2月28日视为已满周岁
// 3. 从"已满多少岁"的角度看，2月28日已经过完了2月的大部分时间，而3月1日则进入了新的月份
func adjustBirthdayForLeapYear(birthYear int, birthMonth time.Month, birthDay int, targetYear int) (time.Month, int) {
	if birthMonth == time.February && birthDay == 29 {
		if !isLeapYear(targetYear) {
			return time.February, 28
		}
	}
	return birthMonth, birthDay
}

func resolveCurrentDate(currentDate time.Time) time.Time {
	if currentDate.IsZero() {
		return time.Now()
	}
	return currentDate
}

func validateDates(birthDate, currentDate time.Time) error {
	if birthDate.After(currentDate) {
		return ErrBirthDateAfterCurrent
	}
	return nil
}

func CalculateAge(birthDate, currentDate time.Time) (Age, error) {
	currentDate = resolveCurrentDate(currentDate)
	if err := validateDates(birthDate, currentDate); err != nil {
		return Age{}, err
	}

	birthYear, birthMonth, birthDay := birthDate.Date()
	currentYear, currentMonth, currentDay := currentDate.Date()

	years := currentYear - birthYear
	months := int(currentMonth - birthMonth)
	days := currentDay - birthDay

	if days < 0 {
		months--
		prevMonth := currentMonth - 1
		prevYear := currentYear
		if prevMonth < time.January {
			prevMonth = time.December
			prevYear--
		}
		days += daysInMonth(prevYear, prevMonth)
	}

	if months < 0 {
		years--
		months += 12
	}

	return Age{Years: years, Months: months, Days: days}, nil
}

func CalculateAgeYears(birthDate, currentDate time.Time) (int, error) {
	currentDate = resolveCurrentDate(currentDate)
	if err := validateDates(birthDate, currentDate); err != nil {
		return 0, err
	}

	birthYear, birthMonth, birthDay := birthDate.Date()
	currentYear, currentMonth, currentDay := currentDate.Date()

	adjMonth, adjDay := adjustBirthdayForLeapYear(birthYear, birthMonth, birthDay, currentYear)

	years := currentYear - birthYear

	if currentMonth < adjMonth || (currentMonth == adjMonth && currentDay < adjDay) {
		years--
	}

	return years, nil
}

func HasReachedAge(birthDate, currentDate time.Time, targetAge int) (bool, error) {
	ageYears, err := CalculateAgeYears(birthDate, currentDate)
	if err != nil {
		return false, err
	}
	return ageYears >= targetAge, nil
}

func IsAdult(birthDate, currentDate time.Time) (bool, error) {
	return HasReachedAge(birthDate, currentDate, 18)
}

func HasReachedRetirementAge(birthDate, currentDate time.Time, gender Gender) (bool, error) {
	retirementAge := 60
	if gender == Female {
		retirementAge = 55
	}
	return HasReachedAge(birthDate, currentDate, retirementAge)
}

func DaysBetween(startDate, endDate time.Time) int {
	start := startDate.Truncate(24 * time.Hour)
	end := endDate.Truncate(24 * time.Hour)

	if start.After(end) {
		start, end = end, start
	}

	diff := end.Sub(start)
	return int(diff.Hours() / 24)
}

func DaysUntilNextBirthday(birthDate, currentDate time.Time) (int, error) {
	currentDate = resolveCurrentDate(currentDate)
	if err := validateDates(birthDate, currentDate); err != nil {
		return 0, err
	}

	birthYear, birthMonth, birthDay := birthDate.Date()
	currentYear, currentMonth, currentDay := currentDate.Date()

	adjMonthThisYear, adjDayThisYear := adjustBirthdayForLeapYear(birthYear, birthMonth, birthDay, currentYear)
	birthdayThisYear := time.Date(currentYear, adjMonthThisYear, adjDayThisYear, 0, 0, 0, 0, currentDate.Location())

	if currentDate.Before(birthdayThisYear) {
		return DaysBetween(currentDate, birthdayThisYear), nil
	}

	if currentDate.Equal(birthdayThisYear) || (currentMonth == adjMonthThisYear && currentDay == adjDayThisYear) {
		return 0, nil
	}

	nextYear := currentYear + 1
	adjMonthNextYear, adjDayNextYear := adjustBirthdayForLeapYear(birthYear, birthMonth, birthDay, nextYear)
	birthdayNextYear := time.Date(nextYear, adjMonthNextYear, adjDayNextYear, 0, 0, 0, 0, currentDate.Location())

	return DaysBetween(currentDate, birthdayNextYear), nil
}
