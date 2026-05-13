package service

import (
	"asset-depreciation/database"
	"asset-depreciation/models"
	"errors"
)

var (
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrAssetNotFound     = errors.New("asset not found")
)

type Transition struct {
	From   models.AssetStatus
	To     models.AssetStatus
	Action string
}

var validTransitions = map[Transition]bool{
	{models.StatusDraft, models.StatusPending, "submit"}:   true,
	{models.StatusPending, models.StatusApproved, "approve"}: true,
	{models.StatusPending, models.StatusDraft, "reject"}:    true,
	{models.StatusApproved, models.StatusRunning, "execute"}: true,
	{models.StatusRunning, models.StatusCompleted, "complete"}: true,
}

func ValidateTransition(from, to models.AssetStatus, action string) bool {
	return validTransitions[Transition{from, to, action}]
}

func GetNextStatus(current models.AssetStatus, action string) (models.AssetStatus, error) {
	switch {
	case current == models.StatusDraft && action == "submit":
		return models.StatusPending, nil
	case current == models.StatusPending && action == "approve":
		return models.StatusApproved, nil
	case current == models.StatusPending && action == "reject":
		return models.StatusDraft, nil
	case current == models.StatusApproved && action == "execute":
		return models.StatusRunning, nil
	case current == models.StatusRunning && action == "complete":
		return models.StatusCompleted, nil
	default:
		return "", ErrInvalidTransition
	}
}

func TransitionAsset(assetID int64, action, remark string) (*models.Asset, error) {
	asset, err := database.GetAssetByID(assetID)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, ErrAssetNotFound
	}

	nextStatus, err := GetNextStatus(asset.Status, action)
	if err != nil {
		return nil, err
	}

	if err := database.UpdateAssetStatus(assetID, asset.Status, nextStatus, action, remark); err != nil {
		return nil, err
	}

	return database.GetAssetByID(assetID)
}

func ValidateCreateAssetRequest(req *models.CreateAssetRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	if req.OriginalValue <= 0 {
		return errors.New("original_value must be greater than 0")
	}
	if req.UsefulLifeYears < 1 {
		return errors.New("useful_life_years must be at least 1")
	}
	if req.PurchaseDate == "" {
		return errors.New("purchase_date is required")
	}
	if req.Method != models.MethodStraightLine && req.Method != models.MethodDoubleDeclining {
		return errors.New("invalid depreciation method")
	}
	if req.SalvageRate < 0 || req.SalvageRate >= 1 {
		return errors.New("salvage_rate must be between 0 (inclusive) and 1 (exclusive)")
	}
	return nil
}
