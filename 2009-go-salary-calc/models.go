package main

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"
)

type SalaryStatus string

const (
	StatusNew        SalaryStatus = "新建"
	StatusProcessing SalaryStatus = "处理中"
	StatusPending    SalaryStatus = "待确认"
	StatusCompleted  SalaryStatus = "已完成"
)

var statusOrder = []SalaryStatus{StatusNew, StatusProcessing, StatusPending, StatusCompleted}

func (s SalaryStatus) IsValid() bool {
	for _, status := range statusOrder {
		if s == status {
			return true
		}
	}
	return false
}

func (s SalaryStatus) Next() (SalaryStatus, error) {
	for i, status := range statusOrder {
		if s == status {
			if i >= len(statusOrder)-1 {
				return "", errors.New("已是最终状态，无法继续推进")
			}
			return statusOrder[i+1], nil
		}
	}
	return "", errors.New("无效状态")
}

func (s *SalaryStatus) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("无法将 %T 转换为 SalaryStatus", value)
	}
	*s = SalaryStatus(str)
	return nil
}

func (s SalaryStatus) Value() (driver.Value, error) {
	return string(s), nil
}

type Employee struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	BaseSalary   int64     `json:"base_salary"`
	Allowance    int64     `json:"allowance"`
	Performance  float64   `json:"performance"`
	JoinDate     time.Time `json:"join_date"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ResourceType string

type Resource struct {
	ID          int64        `json:"id"`
	Name        string       `json:"name"`
	Type        string       `json:"type"`
	Description string       `json:"description,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type ResourceRelation struct {
	ID          int64     `json:"id"`
	SourceID    int64     `json:"source_id"`
	SourceType  string    `json:"source_type"`
	TargetID    int64     `json:"target_id"`
	TargetType  string    `json:"target_type"`
	Relation    string    `json:"relation"`
	CreatedAt   time.Time `json:"created_at"`
}

type OvertimeRequest struct {
	WorkdayHours   float64 `json:"workday_hours"`
	WeekendHours   float64 `json:"weekend_hours"`
	HolidayHours   float64 `json:"holiday_hours"`
}

type SalaryCalculationRequest struct {
	EmployeeID   int64             `json:"employee_id"`
	Year         int               `json:"year"`
	Month        int               `json:"month"`
	Overtime     OvertimeRequest   `json:"overtime"`
	ResourceID   int64             `json:"resource_id,omitempty"`
	ResourceType string            `json:"resource_type,omitempty"`
}

type SalaryRecord struct {
	ID                 int64        `json:"id"`
	EmployeeID         int64        `json:"employee_id"`
	Year               int          `json:"year"`
	Month              int          `json:"month"`
	BaseSalary         int64        `json:"base_salary"`
	Allowance          int64        `json:"allowance"`
	Performance        float64      `json:"performance"`
	MonthlySalary      int64        `json:"monthly_salary"`
	HourlyRate         int64        `json:"hourly_rate"`
	OvertimeWorkday    int64        `json:"overtime_workday"`
	OvertimeWeekend    int64        `json:"overtime_weekend"`
	OvertimeHoliday    int64        `json:"overtime_holiday"`
	TotalOvertime      int64        `json:"total_overtime"`
	SocialInsurance    int64        `json:"social_insurance"`
	TaxableIncome      int64        `json:"taxable_income"`
	IncomeTax          int64        `json:"income_tax"`
	NetSalary          int64        `json:"net_salary"`
	ResourceID         *int64       `json:"resource_id,omitempty"`
	ResourceType       *string      `json:"resource_type,omitempty"`
	Status             SalaryStatus `json:"status"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at"`
}

type ResourceSummary struct {
	ResourceID   int64  `json:"resource_id"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	TotalSalary  int64  `json:"total_salary"`
	RecordCount  int64  `json:"record_count"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}


