package model

import (
	"time"
)

type PackageType string

const (
	PackageBasic    PackageType = "basic"
	PackagePro      PackageType = "pro"
	PackageEnterprise PackageType = "enterprise"
)

type PackageInfo struct {
	ID           uint
	Type         PackageType
	Name         string
	PriceMonthly float64
	SmsQuota     int
	StorageQuota int // GB
}

type Customer struct {
	ID              uint
	Name            string
	CurrentPackage  PackageType
	PackageStartDate time.Time
	PackageEndDate   *time.Time
	CreatedAt       time.Time
}

type PackageChange struct {
	ID           uint
	CustomerID   uint
	OldPackage   PackageType
	NewPackage   PackageType
	ChangeDate   time.Time
	CreatedAt    time.Time
}

type Usage struct {
	ID          uint
	CustomerID  uint
	Year        int
	Month       int
	SmsUsed     int
	StorageUsed float64 // GB
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type BillStatus string

const (
	BillStatusPending    BillStatus = "pending"
	BillStatusOverdue    BillStatus = "overdue"
	BillStatusSevereOverdue BillStatus = "severe_overdue"
	BillStatusPaid       BillStatus = "paid"
)

type Bill struct {
	ID             uint
	CustomerID     uint
	BillYear       int
	BillMonth      int
	PackageFee     float64
	SmsFee         float64
	StorageFee     float64
	TotalAmount    float64
	Status         BillStatus
	DueDate        time.Time
	PaidAt         *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type BillPayment struct {
	ID        uint
	BillID    uint
	Operator  string
	Remark    string
	PaidAt    time.Time
	CreatedAt time.Time
}

type Config struct {
	ID          uint
	Key         string
	Value       string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type BillDetail struct {
	Bill
	CustomerName   string
	DaysInMonth    int
	PackageDetails []PackagePeriodDetail
}

type PackagePeriodDetail struct {
	PackageType   PackageType
	StartDate     time.Time
	EndDate       time.Time
	Days          int
	DailyPrice    float64
	PeriodFee     float64
}
