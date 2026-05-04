package service

import (
	"billing-service/model"
	"billing-service/repository"
	"math"
	"time"
)

type BillService struct {
	billRepo     *repository.BillRepository
	customerRepo *repository.CustomerRepository
	packageRepo  *repository.PackageRepository
	usageRepo    *repository.UsageRepository
	configRepo   *repository.ConfigRepository
}

func NewBillService(
	billRepo *repository.BillRepository,
	customerRepo *repository.CustomerRepository,
	packageRepo *repository.PackageRepository,
	usageRepo *repository.UsageRepository,
	configRepo *repository.ConfigRepository,
) *BillService {
	return &BillService{
		billRepo:     billRepo,
		customerRepo: customerRepo,
		packageRepo:  packageRepo,
		usageRepo:    usageRepo,
		configRepo:   configRepo,
	}
}

func (s *BillService) GenerateMonthlyBill(customerID uint, year int, month int) (*model.Bill, error) {
	existing, err := s.billRepo.GetByCustomerAndMonth(customerID, year, month)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	customer, err := s.customerRepo.GetByID(customerID)
	if err != nil {
		return nil, err
	}

	packageFee, err := s.calculatePackageFee(customerID, year, month)
	if err != nil {
		return nil, err
	}

	smsFee, err := s.calculateSmsFee(customerID, year, month)
	if err != nil {
		return nil, err
	}

	storageFee, err := s.calculateStorageFee(customerID, year, month)
	if err != nil {
		return nil, err
	}

	totalAmount := packageFee + smsFee + storageFee

	dueDays, err := s.configRepo.GetPaymentDueDays()
	if err != nil {
		return nil, err
	}

	nextMonth := time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.Local)
	dueDate := nextMonth.AddDate(0, 0, dueDays-1)

	status := model.BillStatusPending

	unpaidBills, err := s.billRepo.GetUnpaidBillsBefore(customerID, year, month)
	if err != nil {
		return nil, err
	}
	if len(unpaidBills) > 0 {
		status = model.BillStatusOverdue
	}

	bill := &model.Bill{
		CustomerID:  customerID,
		BillYear:    year,
		BillMonth:   month,
		PackageFee:  packageFee,
		SmsFee:      smsFee,
		StorageFee:  storageFee,
		TotalAmount: totalAmount,
		Status:      status,
		DueDate:     dueDate,
	}

	if err := s.billRepo.Create(bill); err != nil {
		return nil, err
	}

	return bill, nil
}

func (s *BillService) calculatePackageFee(customerID uint, year int, month int) (float64, error) {
	customer, err := s.customerRepo.GetByID(customerID)
	if err != nil {
		return 0, err
	}

	pkgChanges, err := s.customerRepo.GetPackageChanges(customerID, year, month)
	if err != nil {
		return 0, err
	}

	daysInMonth := getDaysInMonth(year, month)
	monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

	type period struct {
		start   time.Time
		end     time.Time
		pkgType model.PackageType
	}

	var periods []period

	currentPkg := customer.CurrentPackage
	currentStart := monthStart

	if len(pkgChanges) == 0 {
		periods = append(periods, period{
			start:   monthStart,
			end:     monthEnd,
			pkgType: customer.CurrentPackage,
		})
	} else {
		for _, change := range pkgChanges {
			if change.ChangeDate.After(currentStart) && !change.ChangeDate.After(monthEnd) {
				periods = append(periods, period{
					start:   currentStart,
					end:     change.ChangeDate.Add(-time.Second),
					pkgType: currentPkg,
				})
				currentStart = change.ChangeDate
				currentPkg = change.NewPackage
			}
		}

		if !currentStart.After(monthEnd) {
			periods = append(periods, period{
				start:   currentStart,
				end:     monthEnd,
				pkgType: currentPkg,
			})
		}
	}

	var totalFee float64

	for _, p := range periods {
		pkg, err := s.packageRepo.GetByType(p.pkgType)
		if err != nil {
			return 0, err
		}

		days := p.end.Day() - p.start.Day() + 1
		if p.start.Month() != p.end.Month() {
			days = getDaysInMonth(p.start.Year(), int(p.start.Month())) - p.start.Day() + 1
		}

		dailyPrice := pkg.PriceMonthly / float64(daysInMonth)
		periodFee := dailyPrice * float64(days)

		totalFee += periodFee
	}

	return math.Round(totalFee*100) / 100, nil
}

func (s *BillService) calculateSmsFee(customerID uint, year int, month int) (float64, error) {
	customer, err := s.customerRepo.GetByID(customerID)
	if err != nil {
		return 0, err
	}

	pkg, err := s.packageRepo.GetByType(customer.CurrentPackage)
	if err != nil {
		return 0, err
	}

	usage, err := s.usageRepo.GetByCustomerAndMonth(customerID, year, month)
	if err != nil {
		return 0, err
	}

	if usage == nil || usage.SmsUsed <= pkg.SmsQuota {
		return 0, nil
	}

	extraSms := usage.SmsUsed - pkg.SmsQuota

	smsPrice, err := s.configRepo.GetSmsUnitPrice()
	if err != nil {
		return 0, err
	}

	fee := float64(extraSms) * smsPrice
	return math.Round(fee*100) / 100, nil
}

func (s *BillService) calculateStorageFee(customerID uint, year int, month int) (float64, error) {
	customer, err := s.customerRepo.GetByID(customerID)
	if err != nil {
		return 0, err
	}

	pkg, err := s.packageRepo.GetByType(customer.CurrentPackage)
	if err != nil {
		return 0, err
	}

	usage, err := s.usageRepo.GetByCustomerAndMonth(customerID, year, month)
	if err != nil {
		return 0, err
	}

	if usage == nil || usage.StorageUsed <= float64(pkg.StorageQuota) {
		return 0, nil
	}

	extraStorage := usage.StorageUsed - float64(pkg.StorageQuota)

	storagePrice, err := s.configRepo.GetStorageUnitPrice()
	if err != nil {
		return 0, err
	}

	fee := extraStorage * storagePrice
	return math.Round(fee*100) / 100, nil
}

func (s *BillService) GetBillDetail(billID uint) (*model.BillDetail, error) {
	bill, err := s.billRepo.GetByID(billID)
	if err != nil {
		return nil, err
	}

	customer, err := s.customerRepo.GetByID(bill.CustomerID)
	if err != nil {
		return nil, err
	}

	daysInMonth := getDaysInMonth(bill.BillYear, bill.BillMonth)

	pkgChanges, err := s.customerRepo.GetPackageChanges(customer.ID, bill.BillYear, bill.BillMonth)
	if err != nil {
		return nil, err
	}

	monthStart := time.Date(bill.BillYear, time.Month(bill.BillMonth), 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

	var packageDetails []model.PackagePeriodDetail
	currentPkg := customer.CurrentPackage
	currentStart := monthStart

	if len(pkgChanges) == 0 {
		pkg, err := s.packageRepo.GetByType(customer.CurrentPackage)
		if err != nil {
			return nil, err
		}

		dailyPrice := pkg.PriceMonthly / float64(daysInMonth)
		periodFee := dailyPrice * float64(daysInMonth)

		packageDetails = append(packageDetails, model.PackagePeriodDetail{
			PackageType: customer.CurrentPackage,
			StartDate:   monthStart,
			EndDate:     monthEnd,
			Days:        daysInMonth,
			DailyPrice:  math.Round(dailyPrice*100) / 100,
			PeriodFee:   math.Round(periodFee*100) / 100,
		})
	} else {
		for _, change := range pkgChanges {
			if change.ChangeDate.After(currentStart) && !change.ChangeDate.After(monthEnd) {
				pkg, err := s.packageRepo.GetByType(currentPkg)
				if err != nil {
					return nil, err
				}

				endDate := change.ChangeDate.Add(-time.Second)
				days := endDate.Day() - currentStart.Day() + 1

				dailyPrice := pkg.PriceMonthly / float64(daysInMonth)
				periodFee := dailyPrice * float64(days)

				packageDetails = append(packageDetails, model.PackagePeriodDetail{
					PackageType: currentPkg,
					StartDate:   currentStart,
					EndDate:     endDate,
					Days:        days,
					DailyPrice:  math.Round(dailyPrice*100) / 100,
					PeriodFee:   math.Round(periodFee*100) / 100,
				})

				currentStart = change.ChangeDate
				currentPkg = change.NewPackage
			}
		}

		if !currentStart.After(monthEnd) {
			pkg, err := s.packageRepo.GetByType(currentPkg)
			if err != nil {
				return nil, err
			}

			days := monthEnd.Day() - currentStart.Day() + 1
			dailyPrice := pkg.PriceMonthly / float64(daysInMonth)
			periodFee := dailyPrice * float64(days)

			packageDetails = append(packageDetails, model.PackagePeriodDetail{
				PackageType: currentPkg,
				StartDate:   currentStart,
				EndDate:     monthEnd,
				Days:        days,
				DailyPrice:  math.Round(dailyPrice*100) / 100,
				PeriodFee:   math.Round(periodFee*100) / 100,
			})
		}
	}

	detail := &model.BillDetail{
		Bill:           *bill,
		CustomerName:   customer.Name,
		DaysInMonth:    daysInMonth,
		PackageDetails: packageDetails,
	}

	return detail, nil
}

func (s *BillService) UpdateBillStatuses() error {
	severeDays, err := s.configRepo.GetSevereOverdueDays()
	if err != nil {
		return err
	}

	bills, err := s.billRepo.GetAll()
	if err != nil {
		return err
	}

	now := time.Now()

	for _, bill := range bills {
		if bill.Status == model.BillStatusPaid {
			continue
		}

		daysOverdue := int(now.Sub(bill.DueDate).Hours() / 24)

		if daysOverdue > severeDays {
			if bill.Status != model.BillStatusSevereOverdue {
				if err := s.billRepo.UpdateStatus(bill.ID, model.BillStatusSevereOverdue); err != nil {
					return err
				}
			}
		} else if daysOverdue > 0 {
			if bill.Status == model.BillStatusPending {
				if err := s.billRepo.UpdateStatus(bill.ID, model.BillStatusOverdue); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *BillService) MarkBillAsPaid(billID uint, operator string, remark string) error {
	return s.billRepo.MarkAsPaid(billID, operator, remark)
}

func (s *BillService) GetAllBills() ([]*model.Bill, error) {
	return s.billRepo.GetAll()
}

func (s *BillService) GetCustomerBills(customerID uint) ([]*model.Bill, error) {
	return s.billRepo.GetByCustomer(customerID)
}

func (s *BillService) GetBillByID(billID uint) (*model.Bill, error) {
	return s.billRepo.GetByID(billID)
}

func (s *BillService) GetPaymentHistory(billID uint) ([]*model.BillPayment, error) {
	return s.billRepo.GetPaymentHistory(billID)
}

func getDaysInMonth(year int, month int) int {
	return time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.Local).Day()
}
