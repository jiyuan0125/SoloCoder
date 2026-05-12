package services

import (
	"smart-exam/database"
	"smart-exam/models"

	"gorm.io/gorm"
)

type BillingService struct {
	db *gorm.DB
}

func NewBillingService() *BillingService {
	return &BillingService{db: database.DB}
}

func (s *BillingService) ChargeExam(billID uint, examID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var bill models.Bill
		if err := tx.First(&bill, billID).Error; err != nil {
			return err
		}

		if bill.Status != models.BillStatusActive {
			return nil
		}

		var examAmount float64 = 1.0

		var items []models.BillItem
		if err := tx.Where("bill_id = ? AND is_completed = ?", billID, false).Find(&items).Error; err != nil {
			return err
		}

		if len(items) > 0 {
			for i := range items {
				if items[i].ExamID != nil && *items[i].ExamID == examID {
					items[i].IsCompleted = true
					examAmount = items[i].Amount
					if err := tx.Save(&items[i]).Error; err != nil {
						return err
					}
					break
				}
			}
		}

		if bill.RemainingAmount < examAmount {
			return nil
		}

		bill.RemainingAmount -= examAmount

		if bill.RemainingAmount <= 0 {
			bill.Status = models.BillStatusExhausted
		}

		return tx.Save(&bill).Error
	})
}

func (s *BillingService) AdjustBillAmount(billID uint, newTotal float64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var bill models.Bill
		if err := tx.First(&bill, billID).Error; err != nil {
			return err
		}

		var completedItems []models.BillItem
		if err := tx.Where("bill_id = ? AND is_completed = ?", billID, true).Find(&completedItems).Error; err != nil {
			return err
		}

		var totalCompleted float64
		for _, item := range completedItems {
			totalCompleted += item.Amount
		}

		var incompleteItems []models.BillItem
		if err := tx.Where("bill_id = ? AND is_completed = ?", billID, false).Find(&incompleteItems).Error; err != nil {
			return err
		}

		var currentIncompleteTotal float64
		for _, item := range incompleteItems {
			currentIncompleteTotal += item.Amount
		}

		newIncompleteTotal := newTotal - totalCompleted
		if newIncompleteTotal < 0 {
			newIncompleteTotal = 0
		}

		if currentIncompleteTotal > 0 && newIncompleteTotal > 0 {
			ratio := newIncompleteTotal / currentIncompleteTotal
			for i := range incompleteItems {
				incompleteItems[i].Amount = incompleteItems[i].Amount * ratio
				if err := tx.Save(&incompleteItems[i]).Error; err != nil {
					return err
				}
			}
		}

		bill.TotalAmount = newTotal
		bill.RemainingAmount = newIncompleteTotal

		if bill.RemainingAmount <= 0 && bill.Status == models.BillStatusActive {
			bill.Status = models.BillStatusExhausted
		} else if bill.RemainingAmount > 0 && bill.Status == models.BillStatusExhausted {
			bill.Status = models.BillStatusActive
		}

		return tx.Save(&bill).Error
	})
}
