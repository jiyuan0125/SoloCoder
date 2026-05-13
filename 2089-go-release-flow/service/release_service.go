package service

import (
	"errors"
	"fmt"

	"release-flow/dao"
	"release-flow/model"
)

type ReleaseService struct {
	dao *dao.ReleaseDAO
}

func NewReleaseService(d *dao.ReleaseDAO) *ReleaseService {
	return &ReleaseService{dao: d}
}

func (s *ReleaseService) CreateRelease(name, version string) (*model.Release, error) {
	release := &model.Release{
		Name:    name,
		Version: version,
	}
	err := s.dao.Create(release)
	return release, err
}

func (s *ReleaseService) GetReleaseByID(id int64) (*model.Release, error) {
	return s.dao.GetByID(id)
}

func (s *ReleaseService) GetAllReleases() ([]*model.Release, error) {
	return s.dao.GetAll()
}

func (s *ReleaseService) SubmitForTesting(id int64, operator, reason string) (*model.Release, error) {
	release, err := s.dao.GetByID(id)
	if err != nil {
		return nil, err
	}

	if err := release.ValidateStatusTransition(model.StatusTesting); err != nil {
		return nil, errors.New(err.Error())
	}

	release.CodeFrozen = true
	if err := s.dao.FreezeCode(id); err != nil {
		return nil, err
	}

	log := &model.StatusLog{
		ReleaseID:  release.ID,
		FromStatus: release.Status,
		ToStatus:   model.StatusTesting,
		Operator:   operator,
		Reason:     reason,
	}

	release.Status = model.StatusTesting
	if err := s.dao.UpdateStatus(release, log); err != nil {
		return nil, err
	}

	return release, nil
}

func (s *ReleaseService) RejectTesting(id int64, operator, reason string) (*model.Release, error) {
	if reason == "" {
		return nil, errors.New("打回时必须填写 bug 描述")
	}

	release, err := s.dao.GetByID(id)
	if err != nil {
		return nil, err
	}

	if release.Status != model.StatusTesting {
		return nil, errors.New("只有测试中的发布单才能打回")
	}

	if err := release.ValidateStatusTransition(model.StatusDeveloping); err != nil {
		return nil, errors.New(err.Error())
	}

	log := &model.StatusLog{
		ReleaseID:  release.ID,
		FromStatus: release.Status,
		ToStatus:   model.StatusDeveloping,
		Operator:   operator,
		Reason:     reason,
	}

	release.Status = model.StatusDeveloping
	release.CodeFrozen = false
	if err := s.dao.UpdateStatus(release, log); err != nil {
		return nil, err
	}

	return release, nil
}

func (s *ReleaseService) PassTesting(id int64, operator, reason string) (*model.Release, error) {
	release, err := s.dao.GetByID(id)
	if err != nil {
		return nil, err
	}

	if err := release.ValidateStatusTransition(model.StatusPending); err != nil {
		return nil, errors.New(err.Error())
	}

	log := &model.StatusLog{
		ReleaseID:  release.ID,
		FromStatus: release.Status,
		ToStatus:   model.StatusPending,
		Operator:   operator,
		Reason:     reason,
	}

	release.Status = model.StatusPending
	if err := s.dao.UpdateStatus(release, log); err != nil {
		return nil, err
	}

	return release, nil
}

func (s *ReleaseService) StartCanary(id int64, ratio int, operator, reason string) (*model.Release, error) {
	if !model.IsValidCanaryRatio(ratio) {
		return nil, fmt.Errorf("无效的灰度比例，有效值: 1, 5, 10, 50")
	}

	release, err := s.dao.GetByID(id)
	if err != nil {
		return nil, err
	}

	if err := release.ValidateStatusTransition(model.StatusCanary); err != nil {
		return nil, errors.New(err.Error())
	}

	log := &model.StatusLog{
		ReleaseID:  release.ID,
		FromStatus: release.Status,
		ToStatus:   model.StatusCanary,
		Operator:   operator,
		Reason:     reason,
		CanaryRatio: &ratio,
	}

	release.Status = model.StatusCanary
	release.CanaryRatio = &ratio
	if err := s.dao.UpdateStatus(release, log); err != nil {
		return nil, err
	}

	return release, nil
}

func (s *ReleaseService) FullRelease(id int64, operator, reason string) (*model.Release, error) {
	release, err := s.dao.GetByID(id)
	if err != nil {
		return nil, err
	}

	if err := release.ValidateStatusTransition(model.StatusFullRelease); err != nil {
		return nil, errors.New(err.Error())
	}

	log := &model.StatusLog{
		ReleaseID:  release.ID,
		FromStatus: release.Status,
		ToStatus:   model.StatusFullRelease,
		Operator:   operator,
		Reason:     reason,
	}

	release.Status = model.StatusFullRelease
	if err := s.dao.UpdateStatus(release, log); err != nil {
		return nil, err
	}

	return release, nil
}

func (s *ReleaseService) CompleteRelease(id int64, operator, reason string) (*model.Release, error) {
	release, err := s.dao.GetByID(id)
	if err != nil {
		return nil, err
	}

	if err := release.ValidateStatusTransition(model.StatusCompleted); err != nil {
		return nil, errors.New(err.Error())
	}

	log := &model.StatusLog{
		ReleaseID:  release.ID,
		FromStatus: release.Status,
		ToStatus:   model.StatusCompleted,
		Operator:   operator,
		Reason:     reason,
	}

	release.Status = model.StatusCompleted
	if err := s.dao.UpdateStatus(release, log); err != nil {
		return nil, err
	}

	return release, nil
}

func (s *ReleaseService) CanaryRollback(id int64, operator, reason string) (*model.RollbackRecord, error) {
	release, err := s.dao.GetByID(id)
	if err != nil {
		return nil, err
	}

	if release.Status != model.StatusCanary && release.Status != model.StatusFullRelease {
		return nil, errors.New("只有灰度中或全量发布中才能回滚")
	}

	lastCompleted, err := s.dao.GetLastCompletedRelease()
	if err != nil {
		return nil, err
	}
	if lastCompleted == nil {
		return nil, errors.New("没有可回滚的已完成版本")
	}

	rollbackRecord := &model.RollbackRecord{
		ReleaseID:   release.ID,
		FromVersion: release.Version,
		ToVersion:   lastCompleted.Version,
		Operator:    operator,
		Reason:      reason,
	}

	if err := s.dao.CreateRollbackRecord(rollbackRecord); err != nil {
		return nil, err
	}

	log := &model.StatusLog{
		ReleaseID:  release.ID,
		FromStatus: release.Status,
		ToStatus:   model.StatusDeveloping,
		Operator:   operator,
		Reason:     fmt.Sprintf("回滚到版本: %s, 原因: %s", lastCompleted.Version, reason),
	}

	release.Status = model.StatusDeveloping
	release.CodeFrozen = false
	release.CanaryRatio = nil
	if err := s.dao.UpdateStatus(release, log); err != nil {
		return nil, err
	}

	return rollbackRecord, nil
}

func (s *ReleaseService) GetStatusLogs(id int64) ([]*model.StatusLog, error) {
	_, err := s.dao.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.dao.GetStatusLogs(id)
}

func (s *ReleaseService) GetRollbackRecords(id int64) ([]*model.RollbackRecord, error) {
	_, err := s.dao.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.dao.GetRollbackRecords(id)
}

func (s *ReleaseService) TryCommit(id int64) error {
	release, err := s.dao.GetByID(id)
	if err != nil {
		return err
	}

	if !release.CanAddCommit() {
		return errors.New("代码已冻结或不在开发状态，不允许新的 commit")
	}

	return nil
}
