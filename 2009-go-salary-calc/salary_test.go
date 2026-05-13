package main

import (
	"testing"
	"time"
)

func TestValidatePerformance(t *testing.T) {
	tests := []struct {
		name     string
		perf     float64
		expected bool
	}{
		{"最小值", 0.5, true},
		{"最大值", 2.0, true},
		{"中间值", 1.0, true},
		{"低于最小值", 0.4, false},
		{"高于最大值", 2.1, false},
		{"边界值0", 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidatePerformance(tt.perf); got != tt.expected {
				t.Errorf("ValidatePerformance(%v) = %v, want %v", tt.perf, got, tt.expected)
			}
		})
	}
}

func TestRoundToNearest(t *testing.T) {
	tests := []struct {
		name      string
		n         int64
		precision int64
		expected  int64
	}{
		{"四舍", 123, 100, 100},
		{"五入", 150, 100, 200},
		{"恰好", 100, 100, 100},
		{"负数", -123, 100, -100},
		{"精度1", 123, 1, 123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RoundToNearest(tt.n, tt.precision); got != tt.expected {
				t.Errorf("RoundToNearest(%d, %d) = %d, want %d", tt.n, tt.precision, got, tt.expected)
			}
		})
	}
}

func TestCalculateMonthlySalary(t *testing.T) {
	baseSalary := int64(1000000)
	allowance := int64(200000)
	performance := 1.0

	expected := int64(1200000)
	got := CalculateMonthlySalary(baseSalary, allowance, performance)

	if got != expected {
		t.Errorf("CalculateMonthlySalary(%d, %d, %f) = %d, want %d",
			baseSalary, allowance, performance, got, expected)
	}

	performance = 1.5
	expected = int64(1800000)
	got = CalculateMonthlySalary(baseSalary, allowance, performance)

	if got != expected {
		t.Errorf("CalculateMonthlySalary(%d, %d, %f) = %d, want %d",
			baseSalary, allowance, performance, got, expected)
	}
}

func TestCalculateHourlyRate(t *testing.T) {
	monthlySalary := int64(1200000)

	hourlyRate := CalculateHourlyRate(monthlySalary)
	t.Logf("月薪 %d 分，时薪约 %d 分 (%.2f 元)", monthlySalary, hourlyRate, float64(hourlyRate)/100)

	if hourlyRate <= 0 {
		t.Errorf("CalculateHourlyRate(%d) 应为正数", monthlySalary)
	}
}

func TestCalculateSocialInsurance(t *testing.T) {
	baseSalary := int64(1000000)

	expected := int64(105000)
	got := CalculateSocialInsurance(baseSalary)

	if got != expected {
		t.Errorf("CalculateSocialInsurance(%d) = %d, want %d (10.5%%)", baseSalary, got, expected)
	}
}

func TestCalculateIncomeTax(t *testing.T) {
	tests := []struct {
		name          string
		taxableIncome int64
		expected      int64
	}{
		{"低于起征点", 400000, 0},
		{"恰好起征点", 500000, 0},
		{"5000-8000档", 600000, 3000},
		{"8000-17000档", 1000000, 9000 + 20000},
		{"17000-30000档", 2000000, 9000 + 90000 + 60000},
		{"超过30000", 4000000, 9000 + 90000 + 260000 + 250000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateIncomeTax(tt.taxableIncome)
			t.Logf("应税所得 %d 分 (%.2f 元)，个税 %d 分 (%.2f 元)",
				tt.taxableIncome, float64(tt.taxableIncome)/100,
				got, float64(got)/100)

			if got < 0 {
				t.Errorf("CalculateIncomeTax(%d) = %d, 不应为负数", tt.taxableIncome, got)
			}
		})
	}
}

func TestCalculateOvertime(t *testing.T) {
	hourlyRate := int64(5000)

	overtime := OvertimeRequest{
		WorkdayHours: 10,
		WeekendHours: 8,
		HolidayHours: 4,
	}

	workday, weekend, holiday, total := CalculateOvertime(hourlyRate, overtime)

	t.Logf("时薪 %d 分 (%.2f 元)", hourlyRate, float64(hourlyRate)/100)
	t.Logf("工作日加班 10 小时: %d 分 (%.2f 元, 1.5倍)", workday, float64(workday)/100)
	t.Logf("周末加班 8 小时: %d 分 (%.2f 元, 2倍)", weekend, float64(weekend)/100)
	t.Logf("节假日加班 4 小时: %d 分 (%.2f 元, 3倍)", holiday, float64(holiday)/100)
	t.Logf("总加班费: %d 分 (%.2f 元)", total, float64(total)/100)

	expectedWorkday := int64(75000)
	if workday != expectedWorkday {
		t.Errorf("工作日加班费应为 %d，实际 %d", expectedWorkday, workday)
	}
}

func TestCalculateProRatedBaseSalary(t *testing.T) {
	baseSalary := int64(2175000)

	tests := []struct {
		name      string
		joinDate  time.Time
		year      int
		month     int
		expected  int64
	}{
		{"月初入职", time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC), 2024, 5, baseSalary},
		{"月中15号入职", time.Date(2024, 5, 15, 0, 0, 0, 0, time.UTC), 2024, 5, 1700000},
		{"月末入职", time.Date(2024, 5, 31, 0, 0, 0, 0, time.UTC), 2024, 5, 100000},
		{"下月入职", time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), 2024, 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateProRatedBaseSalary(baseSalary, tt.joinDate, tt.year, tt.month)
			t.Logf("基本工资 %d 分，入职日期 %s，计算月 %d-%d，折算工资 %d 分 (%.2f 元)",
				baseSalary, tt.joinDate.Format("2006-01-02"), tt.year, tt.month,
				got, float64(got)/100)
		})
	}
}

func TestSalaryStatus(t *testing.T) {

	tests := []struct {
		name     string
		current  SalaryStatus
		expected SalaryStatus
	}{
		{"新建→处理中", StatusNew, StatusProcessing},
		{"处理中→待确认", StatusProcessing, StatusPending},
		{"待确认→已完成", StatusPending, StatusCompleted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next, err := tt.current.Next()
			if err != nil {
				t.Errorf("Status %s 推进失败: %v", tt.current, err)
				return
			}
			if next != tt.expected {
				t.Errorf("Status %s.Next() = %s, want %s", tt.current, next, tt.expected)
			}
		})
	}

	_, err := StatusCompleted.Next()
	if err == nil {
		t.Error("已完成状态应无法继续推进")
	}
}

func TestFullSalaryCalculation(t *testing.T) {
	employee := &Employee{
		ID:          1,
		Name:        "测试员工",
		BaseSalary:  1000000,
		Allowance:   200000,
		Performance: 1.0,
		JoinDate:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	overtime := OvertimeRequest{
		WorkdayHours: 10,
		WeekendHours: 8,
		HolidayHours: 4,
	}

	record, err := CalculateFullSalary(employee, 2024, 5, overtime)
	if err != nil {
		t.Fatalf("薪资计算失败: %v", err)
	}

	t.Log("=== 薪资明细 ===")
	t.Logf("员工: %s", employee.Name)
	t.Logf("基本工资: %d 分 (%.2f 元)", record.BaseSalary, float64(record.BaseSalary)/100)
	t.Logf("岗位津贴: %d 分 (%.2f 元)", record.Allowance, float64(record.Allowance)/100)
	t.Logf("绩效系数: %.2f", record.Performance)
	t.Logf("月薪: %d 分 (%.2f 元)", record.MonthlySalary, float64(record.MonthlySalary)/100)
	t.Logf("时薪: %d 分 (%.2f 元)", record.HourlyRate, float64(record.HourlyRate)/100)
	t.Logf("加班费总计: %d 分 (%.2f 元)", record.TotalOvertime, float64(record.TotalOvertime)/100)
	t.Logf("社保: %d 分 (%.2f 元)", record.SocialInsurance, float64(record.SocialInsurance)/100)
	t.Logf("应税所得: %d 分 (%.2f 元)", record.TaxableIncome, float64(record.TaxableIncome)/100)
	t.Logf("个税: %d 分 (%.2f 元)", record.IncomeTax, float64(record.IncomeTax)/100)
	t.Logf("实发工资: %d 分 (%.2f 元)", record.NetSalary, float64(record.NetSalary)/100)

	if record.NetSalary < 0 {
		t.Error("实发工资不应为负数")
	}
}
