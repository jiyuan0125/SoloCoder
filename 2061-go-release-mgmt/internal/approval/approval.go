package approval

import (
	"errors"
	"release-mgmt/internal/database"
	"release-mgmt/internal/model"
)

const MinApprovalsRequired = 2

type ApproveRequest struct {
	ReleaseID int64  `json:"release_id"`
	Approver  string `json:"approver"`
	Approved  bool   `json:"approved"`
	Reason    string `json:"reason,omitempty"`
}

func Approve(req *ApproveRequest) error {
	if req.Approver == "" {
		return errors.New("approver is required")
	}

	release, err := database.GetReleaseByID(req.ReleaseID)
	if err != nil {
		return err
	}
	if release == nil {
		return errors.New("release not found")
	}

	if release.Status != model.StatusPending {
		return errors.New("release is not in pending review status")
	}

	if req.Approver == release.SubmittedBy {
		return errors.New("cannot approve your own release")
	}

	existing, err := database.GetApprovalByUser(req.ReleaseID, req.Approver)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("already submitted approval for this release")
	}

	if !req.Approved && req.Reason == "" {
		return errors.New("rejection reason is required")
	}

	approval := &model.Approval{
		ReleaseID: req.ReleaseID,
		Approver:  req.Approver,
		Approved:  req.Approved,
		Reason:    req.Reason,
	}

	_, err = database.CreateApproval(approval)
	if err != nil {
		return err
	}

	if !req.Approved {
		err = database.UpdateReleaseStatus(req.ReleaseID, model.StatusRejected)
		if err != nil {
			return err
		}

		database.AddOperationHistory(&model.OperationHistory{
			ReleaseID: req.ReleaseID,
			Operation: "reject_release",
			Operator:  req.Approver,
			Details:   "Release rejected, reason: " + req.Reason,
		})
		return nil
	}

	database.AddOperationHistory(&model.OperationHistory{
		ReleaseID: req.ReleaseID,
		Operation: "approve_release",
		Operator:  req.Approver,
		Details:   "Release approved",
	})

	return checkAndUpdateStatus(req.ReleaseID)
}

func checkAndUpdateStatus(releaseID int64) error {
	approvals, err := database.GetApprovalsByRelease(releaseID)
	if err != nil {
		return err
	}

	approvedCount := 0
	for _, a := range approvals {
		if a.Approved {
			approvedCount++
		}
	}

	if approvedCount >= MinApprovalsRequired {
		return database.UpdateReleaseStatus(releaseID, model.StatusApproved)
	}

	return nil
}

func GetApprovals(releaseID int64) ([]model.Approval, error) {
	return database.GetApprovalsByRelease(releaseID)
}

func IsFullyApproved(releaseID int64) (bool, error) {
	approvals, err := database.GetApprovalsByRelease(releaseID)
	if err != nil {
		return false, err
	}

	approvedCount := 0
	for _, a := range approvals {
		if a.Approved {
			approvedCount++
		}
	}

	return approvedCount >= MinApprovalsRequired, nil
}
