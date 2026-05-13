package change

import (
	"errors"
	"release-mgmt/internal/database"
	"release-mgmt/internal/model"
)

type AddChangeRequest struct {
	ReleaseID int64  `json:"release_id"`
	Content   string `json:"content"`
	Type      string `json:"type"`
	Operator  string `json:"operator"`
}

func AddChange(req *AddChangeRequest) (*model.ChangeItem, error) {
	if req.Content == "" {
		return nil, errors.New("content is required")
	}

	release, err := database.GetReleaseByID(req.ReleaseID)
	if err != nil {
		return nil, err
	}
	if release == nil {
		return nil, errors.New("release not found")
	}

	if release.Status == model.StatusCompleted {
		return nil, errors.New("cannot add changes to completed release")
	}

	item := &model.ChangeItem{
		ReleaseID: req.ReleaseID,
		Content:   req.Content,
		Type:      req.Type,
	}

	id, err := database.CreateChangeItem(item)
	if err != nil {
		return nil, err
	}
	item.ID = id

	database.AddOperationHistory(&model.OperationHistory{
		ReleaseID: req.ReleaseID,
		Operation: "add_change",
		Operator:  req.Operator,
		Details:   "Added change item: " + truncate(req.Content, 100),
	})

	return item, nil
}

func GetChanges(releaseID int64) ([]model.ChangeItem, error) {
	return database.GetChangeItemsByRelease(releaseID)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
