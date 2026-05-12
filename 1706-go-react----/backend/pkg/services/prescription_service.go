package services

import (
	"errors"
	"fmt"
	"hospital-pharmacy/pkg/models"
	"hospital-pharmacy/pkg/repositories"
	"hospital-pharmacy/pkg/utils"
	"time"
)

type PrescriptionItemRequest struct {
	DrugID   uint    `json:"drug_id"`
	Quantity float64 `json:"quantity"`
	Usage    string  `json:"usage"`
}

type CreatePrescriptionRequest struct {
	PatientName string                      `json:"patient_name"`
	PatientID   string                      `json:"patient_id"`
	Diagnosis   string                      `json:"diagnosis"`
	Items       []PrescriptionItemRequest   `json:"items"`
}

func CreatePrescription(req CreatePrescriptionRequest) (*models.Prescription, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("处方必须包含药品")
	}

	today := time.Now().Format("20060102")
	seq, err := repositories.GetLatestPrescriptionSeq(today)
	if err != nil {
		return nil, err
	}

	prescriptionNo := "CF" + today + fmt.Sprintf("%04d", seq+1)

	var prescriptionItems []models.PrescriptionItem
	for _, item := range req.Items {
		drug, err := repositories.GetDrugByID(item.DrugID)
		if err != nil {
			return nil, errors.New("药品不存在")
		}

		if err := utils.CheckPrescriptionItem(drug, item.Quantity); err != nil {
			return nil, err
		}

		prescriptionItems = append(prescriptionItems, models.PrescriptionItem{
			DrugID:   item.DrugID,
			DrugName: drug.GenericName,
			Quantity: item.Quantity,
			Usage:    item.Usage,
		})
	}

	prescription := &models.Prescription{
		PrescriptionNo: prescriptionNo,
		PatientName:    req.PatientName,
		PatientID:      req.PatientID,
		Diagnosis:      req.Diagnosis,
		Status:         models.PrescriptionPending,
		Items:          prescriptionItems,
	}

	if err := repositories.CreatePrescription(prescription); err != nil {
		return nil, err
	}

	return prescription, nil
}

func ApprovePrescription(id uint, operator1, operator2 string) error {
	prescription, err := repositories.GetPrescriptionByID(id)
	if err != nil {
		return errors.New("处方不存在")
	}

	if prescription.Status != models.PrescriptionPending {
		if prescription.Status == models.PrescriptionApproved {
			return errors.New("处方已审核通过，无法重复审核")
		}
		return errors.New("处方已取消，无法审核")
	}

	for _, item := range prescription.Items {
		drug, err := repositories.GetDrugByID(item.DrugID)
		if err != nil {
			return err
		}

		available, err := repositories.GetAvailableStock(item.DrugID)
		if err != nil {
			return err
		}

		if float64(available) < item.Quantity {
			return errors.New(fmt.Sprintf("药品【%s】库存不足，需要: %.0f，可用: %d",
				drug.GenericName, item.Quantity, available))
		}
	}

	for _, item := range prescription.Items {
		qty := int(item.Quantity)
		if qty < 1 {
			qty = 1
		}
		err := StockOut(StockOutRequest{
			DrugID:    item.DrugID,
			Quantity:  qty,
			Operator1: operator1,
			Operator2: operator2,
			Remark:    "处方出库: " + prescription.PrescriptionNo,
		})
		if err != nil {
			return err
		}
	}

	prescription.Status = models.PrescriptionApproved
	return repositories.UpdatePrescription(prescription)
}

func RejectPrescription(id uint, comment string) error {
	prescription, err := repositories.GetPrescriptionByID(id)
	if err != nil {
		return errors.New("处方不存在")
	}

	if prescription.Status != models.PrescriptionPending {
		return errors.New("只能拒绝待审核的处方")
	}

	prescription.Status = models.PrescriptionCancelled
	prescription.ReviewComment = comment
	return repositories.UpdatePrescription(prescription)
}

func CancelPrescription(id uint) error {
	prescription, err := repositories.GetPrescriptionByID(id)
	if err != nil {
		return errors.New("处方不存在")
	}

	if prescription.Status == models.PrescriptionApproved {
		return errors.New("已审核通过的处方不能取消")
	}

	if prescription.Status == models.PrescriptionCancelled {
		return errors.New("处方已取消")
	}

	prescription.Status = models.PrescriptionCancelled
	return repositories.UpdatePrescription(prescription)
}

func GetPrescription(id uint) (*models.Prescription, error) {
	return repositories.GetPrescriptionByID(id)
}

func ListPrescriptions(status string) ([]models.Prescription, error) {
	return repositories.ListPrescriptions(status)
}
