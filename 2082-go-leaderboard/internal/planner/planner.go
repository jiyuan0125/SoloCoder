package planner

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"leaderboard/internal/database"
)

type Plan struct {
	ID          int64   `json:"id"`
	TotalAmount float64 `json:"total_amount"`
	Periods     []Period `json:"periods"`
}

type Period struct {
	PeriodIndex  int     `json:"period_index"`
	PlannedAmount float64 `json:"planned_amount"`
	SettledAmount float64 `json:"settled_amount"`
	IsSettled    bool    `json:"is_settled"`
}

type Service struct {
	mu sync.RWMutex
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) CreatePlan(ctx context.Context, totalAmount float64, numPeriods int) (*Plan, error) {
	if totalAmount < 0 {
		return nil, fmt.Errorf("total amount cannot be negative")
	}
	if numPeriods <= 0 {
		return nil, fmt.Errorf("number of periods must be positive")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `INSERT INTO plans (total_amount) VALUES (?)`, totalAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to insert plan: %w", err)
	}

	planID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get plan ID: %w", err)
	}

	periodAmount := totalAmount / float64(numPeriods)
	periods := make([]Period, numPeriods)

	for i := 0; i < numPeriods; i++ {
		if _, err = tx.ExecContext(ctx, `INSERT INTO plan_periods (plan_id, period_index, planned_amount, settled_amount, is_settled) VALUES (?, ?, ?, 0, 0)`,
			planID, i, periodAmount); err != nil {
			return nil, fmt.Errorf("failed to insert plan period: %w", err)
		}
		periods[i] = Period{
			PeriodIndex:   i,
			PlannedAmount: periodAmount,
			SettledAmount: 0,
			IsSettled:     false,
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &Plan{
		ID:          planID,
		TotalAmount: totalAmount,
		Periods:     periods,
	}, nil
}

func (s *Service) GetPlan(ctx context.Context, planID int64) (*Plan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var plan Plan
	err := database.DB.QueryRowContext(ctx, `SELECT id, total_amount FROM plans WHERE id = ?`, planID).Scan(&plan.ID, &plan.TotalAmount)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plan not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to query plan: %w", err)
	}

	rows, err := database.DB.QueryContext(ctx, `SELECT period_index, planned_amount, settled_amount, is_settled FROM plan_periods WHERE plan_id = ? ORDER BY period_index`, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to query plan periods: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var period Period
		var isSettledInt int
		if err := rows.Scan(&period.PeriodIndex, &period.PlannedAmount, &period.SettledAmount, &isSettledInt); err != nil {
			return nil, fmt.Errorf("failed to scan plan period: %w", err)
		}
		period.IsSettled = isSettledInt != 0
		plan.Periods = append(plan.Periods, period)
	}

	return &plan, nil
}

func (s *Service) UpdateTotalAmount(ctx context.Context, planID int64, newTotalAmount float64) (*Plan, error) {
	if newTotalAmount < 0 {
		return nil, fmt.Errorf("total amount cannot be negative")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var oldTotalAmount float64
	err = tx.QueryRowContext(ctx, `SELECT total_amount FROM plans WHERE id = ?`, planID).Scan(&oldTotalAmount)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plan not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to query plan: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `UPDATE plans SET total_amount = ? WHERE id = ?`, newTotalAmount, planID); err != nil {
		return nil, fmt.Errorf("failed to update plan: %w", err)
	}

	rows, err := tx.QueryContext(ctx, `SELECT period_index, planned_amount, settled_amount, is_settled FROM plan_periods WHERE plan_id = ? ORDER BY period_index`, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to query plan periods: %w", err)
	}
	defer rows.Close()

	var periods []Period
	var totalSettled float64
	var numUnsettled int
	var totalPlannedUnsettled float64

	for rows.Next() {
		var period Period
		var isSettledInt int
		if err := rows.Scan(&period.PeriodIndex, &period.PlannedAmount, &period.SettledAmount, &isSettledInt); err != nil {
			return nil, fmt.Errorf("failed to scan plan period: %w", err)
		}
		period.IsSettled = isSettledInt != 0
		periods = append(periods, period)
		
		if period.IsSettled {
			totalSettled += period.SettledAmount
		} else {
			numUnsettled++
			totalPlannedUnsettled += period.PlannedAmount
		}
	}

	remainingAmount := newTotalAmount - totalSettled
	if remainingAmount < 0 {
		return nil, fmt.Errorf("new total amount cannot be less than total settled amount")
	}

	for i := range periods {
		if !periods[i].IsSettled {
			var newPlannedAmount float64
			if totalPlannedUnsettled > 0 {
				newPlannedAmount = periods[i].PlannedAmount / totalPlannedUnsettled * remainingAmount
			} else {
				newPlannedAmount = remainingAmount / float64(numUnsettled)
			}
			
			if _, err = tx.ExecContext(ctx, `UPDATE plan_periods SET planned_amount = ? WHERE plan_id = ? AND period_index = ?`,
				newPlannedAmount, planID, periods[i].PeriodIndex); err != nil {
				return nil, fmt.Errorf("failed to update plan period: %w", err)
			}
			periods[i].PlannedAmount = newPlannedAmount
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	if err = s.validateConsistency(ctx, planID); err != nil {
		return nil, fmt.Errorf("consistency validation failed: %w", err)
	}

	return &Plan{
		ID:          planID,
		TotalAmount: newTotalAmount,
		Periods:     periods,
	}, nil
}

func (s *Service) validateConsistency(ctx context.Context, planID int64) error {
	var totalAmount float64
	err := database.DB.QueryRowContext(ctx, `SELECT total_amount FROM plans WHERE id = ?`, planID).Scan(&totalAmount)
	if err != nil {
		return fmt.Errorf("failed to query plan: %w", err)
	}

	rows, err := database.DB.QueryContext(ctx, `SELECT planned_amount, settled_amount, is_settled FROM plan_periods WHERE plan_id = ?`, planID)
	if err != nil {
		return fmt.Errorf("failed to query plan periods: %w", err)
	}
	defer rows.Close()

	var calculatedTotal float64
	for rows.Next() {
		var planned, settled float64
		var isSettled int
		if err := rows.Scan(&planned, &settled, &isSettled); err != nil {
			return fmt.Errorf("failed to scan plan period: %w", err)
		}
		if isSettled != 0 {
			calculatedTotal += settled
		} else {
			calculatedTotal += planned
		}
	}

	epsilon := 0.0001
	if calculatedTotal-totalAmount > epsilon || totalAmount-calculatedTotal > epsilon {
		return fmt.Errorf("inconsistent total: expected %f, calculated %f", totalAmount, calculatedTotal)
	}

	return nil
}
