package duty

import (
	"context"
	"errors"
	"time"

	"workorder-flow/internal/database"
	"workorder-flow/internal/models"
)

type Scheduler struct {
	db *database.DB
}

func NewScheduler(db *database.DB) *Scheduler {
	return &Scheduler{db: db}
}

var ErrNoDutyUser = errors.New("no duty user available today")

func (s *Scheduler) GetDutyUserForTeam(ctx context.Context, teamID int64, date time.Time) (*models.User, error) {
	dateStr := date.Format("2006-01-02")
	user, err := s.db.GetDutyUserByTeamAndDate(ctx, teamID, dateStr)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, ErrNoDutyUser
		}
		return nil, err
	}
	return user, nil
}

func (s *Scheduler) GetDutyUserForType(ctx context.Context, typ models.WorkOrderType, date time.Time) (*models.User, *models.Team, error) {
	team, err := s.db.GetTeamByType(ctx, typ)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil, ErrNoDutyUser
		}
		return nil, nil, err
	}

	user, err := s.GetDutyUserForTeam(ctx, team.ID, date)
	if err != nil {
		return nil, nil, err
	}

	return user, team, nil
}

func (s *Scheduler) AssignDuty(ctx context.Context, typ models.WorkOrderType) (*models.User, *models.Team, error) {
	now := time.Now()
	user, team, err := s.GetDutyUserForType(ctx, typ, now)
	if err != nil {
		return nil, nil, err
	}
	return user, team, nil
}
