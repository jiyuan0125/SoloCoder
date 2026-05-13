package deployment

import (
	"errors"
	"math/rand"
	"release-mgmt/internal/database"
	"release-mgmt/internal/model"
	"release-mgmt/internal/version"
	"time"
)

type DeployRequest struct {
	ReleaseID    int64  `json:"release_id"`
	Environment  string `json:"environment"`
	Operator     string `json:"operator"`
	GrayRelease  bool   `json:"gray_release,omitempty"`
	GrayPercent  int    `json:"gray_percent,omitempty"`
	Force        bool   `json:"force,omitempty"`
}

func DeployToStaging(req *DeployRequest) (*model.Deployment, error) {
	if err := version.CanDeployToStaging(req.ReleaseID); err != nil {
		return nil, err
	}

	deployment := &model.Deployment{
		ReleaseID:   req.ReleaseID,
		Environment: "staging",
		Status:      "deploying",
	}

	id, err := database.CreateDeployment(deployment)
	if err != nil {
		return nil, err
	}
	deployment.ID = id

	version.MarkAsDeploying(req.ReleaseID, req.Operator)

	database.AddOperationHistory(&model.OperationHistory{
		ReleaseID: req.ReleaseID,
		Operation: "deploy_staging",
		Operator:  req.Operator,
		Details:   "Started deployment to staging environment",
	})

	time.Sleep(100 * time.Millisecond)
	database.UpdateDeploymentStatus(id, "completed")
	deployment.Status = "completed"

	smokePass := runSmokeTest()
	deployment.SmokeTestPass = smokePass

	updateSmokeTestResult(id, smokePass)

	if smokePass {
		database.AddOperationHistory(&model.OperationHistory{
			ReleaseID: req.ReleaseID,
			Operation: "smoke_test_pass",
			Operator:  "system",
			Details:   "Staging smoke test passed",
		})
	} else {
		database.AddOperationHistory(&model.OperationHistory{
			ReleaseID: req.ReleaseID,
			Operation: "smoke_test_fail",
			Operator:  "system",
			Details:   "Staging smoke test failed",
		})
	}

	return deployment, nil
}

func DeployToProduction(req *DeployRequest) (*model.Deployment, error) {
	if err := version.CanDeployToProduction(req.ReleaseID); err != nil {
		return nil, err
	}

	if req.GrayRelease {
		if req.GrayPercent <= 0 || req.GrayPercent > 100 {
			return nil, errors.New("gray release percent must be between 1-100")
		}
	} else {
		req.GrayPercent = 100
	}

	deployment := &model.Deployment{
		ReleaseID:   req.ReleaseID,
		Environment: "production",
		Status:      "deploying",
		GrayRelease: req.GrayRelease,
		GrayPercent: req.GrayPercent,
	}

	id, err := database.CreateDeployment(deployment)
	if err != nil {
		return nil, err
	}
	deployment.ID = id

	database.AddOperationHistory(&model.OperationHistory{
		ReleaseID: req.ReleaseID,
		Operation: "deploy_production",
		Operator:  req.Operator,
		Details:   "Started deployment to production environment",
	})

	time.Sleep(100 * time.Millisecond)
	database.UpdateDeploymentStatus(id, "completed")
	deployment.Status = "completed"

	if req.GrayPercent == 100 {
		version.MarkAsCompleted(req.ReleaseID, req.Operator)
	}

	return deployment, nil
}

func RollbackProduction(releaseID int64, operator string) (*model.Deployment, error) {
	current, err := database.GetLatestCompletedDeployment("production")
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, errors.New("no production deployment found to rollback from")
	}

	prevReleaseID := findPreviousProductionRelease(current.ReleaseID)
	if prevReleaseID == 0 {
		return nil, errors.New("no previous version found for rollback")
	}

	rollback := &model.Deployment{
		ReleaseID:    prevReleaseID,
		Environment:  "production",
		Status:       "deploying",
		IsRollback:   true,
		RollbackFrom: current.ReleaseID,
		GrayPercent:  100,
	}

	id, err := database.CreateDeployment(rollback)
	if err != nil {
		return nil, err
	}
	rollback.ID = id

	database.AddOperationHistory(&model.OperationHistory{
		ReleaseID: current.ReleaseID,
		Operation: "rollback_triggered",
		Operator:  operator,
		Details:   "Rollback triggered, rolling back to previous version",
	})

	time.Sleep(100 * time.Millisecond)
	database.UpdateDeploymentStatus(id, "completed")
	rollback.Status = "completed"

	err = database.UpdateReleaseStatus(current.ReleaseID, model.StatusRollbacked)
	if err != nil {
		return nil, err
	}

	database.AddOperationHistory(&model.OperationHistory{
		ReleaseID: prevReleaseID,
		Operation: "rollback_complete",
		Operator:  operator,
		Details:   "Rollback completed successfully",
	})

	return rollback, nil
}

func runSmokeTest() bool {
	return rand.Float32() > 0.1
}

func updateSmokeTestResult(id int64, passed bool) error {
	_, err := database.GetDB().Exec(`
		UPDATE deployments SET smoke_test_pass = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, passed, id)
	return err
}

func findPreviousProductionRelease(currentReleaseID int64) int64 {
	rows, err := database.GetDB().Query(`
		SELECT DISTINCT release_id
		FROM deployments
		WHERE environment = 'production'
		AND status = 'completed'
		AND release_id < ?
		ORDER BY release_id DESC
		LIMIT 1
	`, currentReleaseID)
	if err != nil {
		return 0
	}
	defer rows.Close()

	var prevID int64
	if rows.Next() {
		rows.Scan(&prevID)
		return prevID
	}
	return 0
}
