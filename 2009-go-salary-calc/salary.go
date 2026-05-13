package main

import (
	"math"
	"time"
)

const (
	workingDaysPerMonth = 21.75
	hoursPerDay         = 8
	socialInsuranceRate = 0.105

	taxThreshold    = 5000 * 100
	bracket1Max     = 8000 * 100
	bracket2Max     = 17000 * 100
	bracket3Max     = 30000 * 100
	bracket1Rate    = 0.03
	bracket2Rate    = 0.10
	bracket3Rate    = 0.20
	bracket4Rate    = 0.25

	minPerformance = 0.5
	maxPerformance = 2.0
)

func RoundFloatToInt64(f float64) int64 {
	return int64(math.Round(f))
}

func RoundToNearest(n int64, precision int64) int64 {
	if precision <= 1 {
		return n
	}
	mod := n % precision
	if mod >= precision/2 {
		return n + precision - mod
	}
	return n - mod
}

func ValidatePerformance(performance float64) bool {
	return performance >= minPerformance && performance <= maxPerformance
}

func CalculateProRatedBaseSalary(baseSalary int64, joinDate time.Time, year, month int) int64 {
	monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

	if joinDate.Before(monthStart) || joinDate.Equal(monthStart) {
		return baseSalary
	}

	if joinDate.After(monthEnd) {
		return 0
	}

	daysInMonth := monthEnd.Day()
	workDaysStart := joinDate.Day()
	actualWorkDays := float64(daysInMonth - workDaysStart + 1)

	dailyRate := float64(baseSalary) / workingDaysPerMonth
	proRated := dailyRate * actualWorkDays

	return RoundFloatToInt64(proRated)
}

func CalculateMonthlySalary(baseSalary int64, allowance int64, performance float64) int64 {
	totalBase := baseSalary + allowance
	monthlySalary := float64(totalBase) * performance
	return RoundFloatToInt64(monthlySalary)
}

func CalculateHourlyRate(monthlySalary int64) int64 {
	hourlyRate := float64(monthlySalary) / workingDaysPerMonth / hoursPerDay
	return RoundFloatToInt64(hourlyRate)
}

func CalculateOvertime(hourlyRate int64, overtime OvertimeRequest) (workday, weekend, holiday, total int64) {
	workday = RoundFloatToInt64(float64(hourlyRate) * 1.5 * overtime.WorkdayHours)
	weekend = RoundFloatToInt64(float64(hourlyRate) * 2.0 * overtime.WeekendHours)
	holiday = RoundFloatToInt64(float64(hourlyRate) * 3.0 * overtime.HolidayHours)
	total = workday + weekend + holiday
	return
}

func CalculateSocialInsurance(baseSalary int64) int64 {
	return RoundFloatToInt64(float64(baseSalary) * socialInsuranceRate)
}

func CalculateIncomeTax(taxableIncome int64) int64 {
	if taxableIncome <= taxThreshold {
		return 0
	}

	actualTaxable := taxableIncome - taxThreshold
	var tax int64

	bracket1 := int64(bracket1Max - taxThreshold)
	if actualTaxable <= bracket1 {
		tax = RoundFloatToInt64(float64(actualTaxable) * bracket1Rate)
		return tax
	}
	tax += RoundFloatToInt64(float64(bracket1) * bracket1Rate)

	bracket2 := int64(bracket2Max - bracket1Max)
	if actualTaxable <= bracket1+bracket2 {
		remainder := actualTaxable - bracket1
		tax += RoundFloatToInt64(float64(remainder) * bracket2Rate)
		return tax
	}
	tax += RoundFloatToInt64(float64(bracket2) * bracket2Rate)

	bracket3 := int64(bracket3Max - bracket2Max)
	if actualTaxable <= bracket1+bracket2+bracket3 {
		remainder := actualTaxable - bracket1 - bracket2
		tax += RoundFloatToInt64(float64(remainder) * bracket3Rate)
		return tax
	}
	tax += RoundFloatToInt64(float64(bracket3) * bracket3Rate)

	remainder := actualTaxable - bracket1 - bracket2 - bracket3
	tax += RoundFloatToInt64(float64(remainder) * bracket4Rate)

	return tax
}

func CalculateFullSalary(
	employee *Employee,
	year, month int,
	overtime OvertimeRequest,
) (*SalaryRecord, error) {
	if !ValidatePerformance(employee.Performance) {
		return nil, &ValidationError{Message: "绩效系数必须在 0.5 到 2.0 之间"}
	}

	proRatedBase := CalculateProRatedBaseSalary(employee.BaseSalary, employee.JoinDate, year, month)

	monthlySalary := CalculateMonthlySalary(proRatedBase, employee.Allowance, employee.Performance)

	hourlyRate := CalculateHourlyRate(monthlySalary)

	overtimeWorkday, overtimeWeekend, overtimeHoliday, totalOvertime := CalculateOvertime(hourlyRate, overtime)

	socialInsurance := CalculateSocialInsurance(proRatedBase)

	totalIncome := monthlySalary + totalOvertime

	taxableIncome := totalIncome - socialInsurance
	if taxableIncome < 0 {
		taxableIncome = 0
	}

	incomeTax := CalculateIncomeTax(taxableIncome)

	netSalary := totalIncome - socialInsurance - incomeTax
	if netSalary < 0 {
		netSalary = 0
	}

	record := &SalaryRecord{
		EmployeeID:      employee.ID,
		Year:            year,
		Month:           month,
		BaseSalary:      proRatedBase,
		Allowance:       employee.Allowance,
		Performance:     employee.Performance,
		MonthlySalary:   monthlySalary,
		HourlyRate:      hourlyRate,
		OvertimeWorkday: overtimeWorkday,
		OvertimeWeekend: overtimeWeekend,
		OvertimeHoliday: overtimeHoliday,
		TotalOvertime:   totalOvertime,
		SocialInsurance: socialInsurance,
		TaxableIncome:   taxableIncome,
		IncomeTax:       incomeTax,
		NetSalary:       netSalary,
	}

	return record, nil
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func YuanToFen(yuan float64) int64 {
	return int64(math.Round(yuan * 100))
}

func FenToYuan(fen int64) float64 {
	return float64(fen) / 100.0
}
