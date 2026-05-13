package version

import (
	"errors"
	"regexp"
	"release-mgmt/internal/database"
	"release-mgmt/internal/model"
)

var semverRegex = regexp.MustCompile(`^([0-9]+)\.([0-9]+)\.([0-9]+)$`)

func IsValidSemver(version string) bool {
	return semverRegex.MatchString(version)
}

type CreateReleaseRequest struct {
	Version     string `json:"version"`
	Description string `json:"description"`
	SubmittedBy string `json:"submitted_by"`
}

func CreateRelease(req *CreateReleaseRequest) (*model.Release, error) {
	if !IsValidSemver(req.Version) {
		return nil, errors.New("invalid semver format, expected MAJOR.MINOR.PATCH")
	}

	exists, err := database.VersionExists(req.Version)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("duplicate version")
	}

	if req.SubmittedBy == "" {
		return nil, errors.New("submitted_by is required")
	}

	release := &model.Release{
		Version:     req.Version,
		Description: req.Description,
		SubmittedBy: req.SubmittedBy,
	}

	id, err := database.CreateRelease(release)
	if err != nil {
		return nil, err
	}
	release.ID = id

	database.AddOperationHistory(&model.OperationHistory{
		ReleaseID: id,
		Operation: "create_release",
		Operator:  req.SubmittedBy,
		Details:   "Created release version " + req.Version,
	})

	return release, nil
}

func GetRelease(version string) (*model.Release, error) {
	release, err := database.GetReleaseByVersion(version)
	if err != nil {
		return nil, err
	}
	return release, nil
}

func GetReleaseByID(id int64) (*model.Release, error) {
	release, err := database.GetReleaseByID(id)
	if err != nil {
		return nil, err
	}
	return release, nil
}

func GetAllReleases() ([]model.Release, error) {
	return database.GetAllReleases()
}

func SubmitForReview(releaseID int64, operator string) error {
	release, err := database.GetReleaseByID(releaseID)
	if err != nil {
		return err
	}
	if release == nil {
		return errors.New("release not found")
	}

	if release.Status != model.StatusDraft && release.Status != model.StatusRejected {
		return errors.New("can only submit releases in draft or rejected status")
	}

	err = database.UpdateReleaseStatus(releaseID, model.StatusPending)
	if err != nil {
		return err
	}

	database.AddOperationHistory(&model.OperationHistory{
		ReleaseID: releaseID,
		Operation: "submit_review",
		Operator:  operator,
		Details:   "Submitted release for review",
	})

	return nil
}

func CanDeployToStaging(releaseID int64) error {
	release, err := database.GetReleaseByID(releaseID)
	if err != nil {
		return err
	}
	if release == nil {
		return errors.New("release not found")
	}

	if release.Status != model.StatusApproved {
		return errors.New("release must be approved before staging deployment")
	}

	return nil
}

func CanDeployToProduction(releaseID int64) error {
	release, err := database.GetReleaseByID(releaseID)
	if err != nil {
		return err
	}
	if release == nil {
		return errors.New("release not found")
	}

	if release.Status != model.StatusApproved && release.Status != model.StatusDeploying {
		return errors.New("release must be approved and have staging deployment")
	}

	staging, err := database.GetLatestStagingDeployment(releaseID)
	if err != nil {
		return err
	}
	if staging == nil {
		return errors.New("no staging deployment found")
	}

	if !staging.SmokeTestPass {
		return errors.New("staging smoke test not passed, cannot deploy to production")
	}

	return nil
}

func MarkAsCompleted(releaseID int64, operator string) error {
	err := database.UpdateReleaseStatus(releaseID, model.StatusCompleted)
	if err != nil {
		return err
	}

	database.AddOperationHistory(&model.OperationHistory{
		ReleaseID: releaseID,
		Operation: "complete_release",
		Operator:  operator,
		Details:   "Release deployment completed",
	})

	return nil
}

func MarkAsDeploying(releaseID int64, operator string) error {
	err := database.UpdateReleaseStatus(releaseID, model.StatusDeploying)
	if err != nil {
		return err
	}

	database.AddOperationHistory(&model.OperationHistory{
		ReleaseID: releaseID,
		Operation: "start_deployment",
		Operator:  operator,
		Details:   "Deployment process started",
	})

	return nil
}
