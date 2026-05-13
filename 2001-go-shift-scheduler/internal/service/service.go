package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"shift-scheduler/internal/models"
	"shift-scheduler/internal/repository"
)

const (
	MaxWeeklyHours    = 40.0
	MinShiftGapHours  = 11.0
)

type ValidationError struct {
	Code    int
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func getWeekRange(date time.Time) (time.Time, time.Time) {
	weekday := date.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	weekStart := date.AddDate(0, 0, -(int(weekday) - 1)).Truncate(24 * time.Hour)
	weekEnd := weekStart.AddDate(0, 0, 6)
	return weekStart, weekEnd
}

func ValidateShiftCreation(empID int64, shiftDate, startTime, endTime time.Time) error {
	existing, err := repository.GetShiftByEmployeeAndDate(empID, shiftDate)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if existing != nil {
		return &ValidationError{Code: 409, Message: fmt.Sprintf("该员工当天已有班次：%s 至 %s",
			existing.StartTime.Format("15:04"), existing.EndTime.Format("15:04"))}
	}

	lastShift, err := repository.GetLastShiftBefore(empID, shiftDate)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if lastShift != nil {
		lastEndTime := lastShift.EndTime
		lastEndDateTime := lastShift.ShiftDate.Add(time.Duration(lastEndTime.Hour())*time.Hour + time.Duration(lastEndTime.Minute())*time.Minute)
		
		newStartDateTime := shiftDate.Add(time.Duration(startTime.Hour())*time.Hour + time.Duration(startTime.Minute())*time.Minute)
		
		hoursDiff := newStartDateTime.Sub(lastEndDateTime).Hours()
		if hoursDiff < MinShiftGapHours {
			earliestPossible := lastEndDateTime.Add(MinShiftGapHours * time.Hour)
			return &ValidationError{
				Code:    400,
				Message: fmt.Sprintf("班次间隔不足11小时。上次下班时间：%s，最早可排时间：%s",
					lastEndDateTime.Format("2006-01-02 15:04"),
					earliestPossible.Format("2006-01-02 15:04")),
			}
		}
	}

	weekStart, weekEnd := getWeekRange(shiftDate)
	weeklyHours, err := repository.GetWeeklyHours(empID, weekStart, weekEnd)
	if err != nil {
		return err
	}

	newHours := endTime.Sub(startTime).Hours()
	if weeklyHours+newHours > MaxWeeklyHours {
		available := MaxWeeklyHours - weeklyHours
		return &ValidationError{
			Code:    400,
			Message: fmt.Sprintf("周工时超过40小时限制。当前已排：%.1f小时，剩余可用：%.1f小时",
				weeklyHours, available),
		}
	}

	return nil
}

func CreateShift(empID int64, shiftDate, startTime, endTime time.Time, position string) (*models.ShiftMaster, error) {
	if err := ValidateShiftCreation(empID, shiftDate, startTime, endTime); err != nil {
		return nil, err
	}

	isHoliday, err := repository.IsHoliday(shiftDate)
	if err != nil {
		return nil, err
	}

	shift, err := repository.CreateShift(empID, shiftDate, startTime, endTime, position, isHoliday)
	if err != nil {
		return nil, err
	}

	return shift, nil
}

func ValidateStatusTransition(current, next models.ShiftStatus) bool {
	transitions := map[models.ShiftStatus][]models.ShiftStatus{
		models.StatusDraft:       {models.StatusReviewing},
		models.StatusReviewing:   {models.StatusApproved, models.StatusRejected},
		models.StatusRejected:    {models.StatusReviewing},
		models.StatusApproved:    {models.StatusInProgress},
		models.StatusInProgress:  {models.StatusCompleted},
		models.StatusCompleted:   {},
	}

	allowed, exists := transitions[current]
	if !exists {
		return false
	}

	for _, s := range allowed {
		if s == next {
			return true
		}
	}
	return false
}

func AdvanceStatus(shiftID int64, operatorID int64, note string) error {
	shift, err := repository.GetShiftByID(shiftID)
	if err != nil {
		return err
	}

	var nextStatus models.ShiftStatus
	switch shift.Status {
	case models.StatusDraft:
		nextStatus = models.StatusReviewing
	case models.StatusReviewing:
		nextStatus = models.StatusApproved
	case models.StatusApproved:
		nextStatus = models.StatusInProgress
	case models.StatusInProgress:
		nextStatus = models.StatusCompleted
	default:
		return &ValidationError{Code: 400, Message: "当前状态无法继续推进"}
	}

	if !ValidateStatusTransition(shift.Status, nextStatus) {
		return &ValidationError{Code: 400, Message: fmt.Sprintf("非法状态跳转：%s -> %s", shift.Status, nextStatus)}
	}

	return repository.UpdateShiftStatus(shiftID, nextStatus, operatorID, note)
}

func RejectShift(shiftID int64, operatorID int64, note string) error {
	shift, err := repository.GetShiftByID(shiftID)
	if err != nil {
		return err
	}

	if shift.Status != models.StatusReviewing {
		return &ValidationError{Code: 400, Message: "只有审核中状态的班次可以被退回"}
	}

	return repository.UpdateShiftStatus(shiftID, models.StatusRejected, operatorID, note)
}

func RequestSwap(requesterShiftID, responderShiftID int64) (*models.SwapRequest, error) {
	reqShift, err := repository.GetShiftByID(requesterShiftID)
	if err != nil {
		return nil, err
	}

	respShift, err := repository.GetShiftByID(responderShiftID)
	if err != nil {
		return nil, err
	}

	if reqShift.Status != models.StatusApproved || respShift.Status != models.StatusApproved {
		return nil, &ValidationError{Code: 400, Message: "只有已通过状态的班次可以申请调班"}
	}

	if err := ValidateShiftCreation(respShift.EmployeeID, reqShift.ShiftDate, reqShift.StartTime, reqShift.EndTime); err != nil {
		return nil, &ValidationError{Code: 400, Message: fmt.Sprintf("调班后接收方(%d)不满足约束：%s", respShift.EmployeeID, err.Error())}
	}

	if err := ValidateShiftCreation(reqShift.EmployeeID, respShift.ShiftDate, respShift.StartTime, respShift.EndTime); err != nil {
		return nil, &ValidationError{Code: 400, Message: fmt.Sprintf("调班后发起方(%d)不满足约束：%s", reqShift.EmployeeID, err.Error())}
	}

	return repository.CreateSwapRequest(requesterShiftID, responderShiftID)
}

func ConfirmSwap(swapID int64) error {
	swap, err := repository.GetSwapRequestByID(swapID)
	if err != nil {
		return err
	}

	if swap.Status != "pending" {
		return &ValidationError{Code: 400, Message: "该调班请求已处理"}
	}

	if err := repository.ExecuteSwap(swapID); err != nil {
		return err
	}

	return repository.UpdateSwapRequestStatus(swapID, "confirmed")
}

func CancelSwap(swapID int64) error {
	swap, err := repository.GetSwapRequestByID(swapID)
	if err != nil {
		return err
	}

	if swap.Status != "pending" {
		return &ValidationError{Code: 400, Message: "该调班请求已处理"}
	}

	return repository.UpdateSwapRequestStatus(swapID, "cancelled")
}

func RejectSwap(swapID int64) error {
	swap, err := repository.GetSwapRequestByID(swapID)
	if err != nil {
		return err
	}

	if swap.Status != "pending" {
		return &ValidationError{Code: 400, Message: "该调班请求已处理"}
	}

	return repository.UpdateSwapRequestStatus(swapID, "rejected")
}

func GetEffectiveHours(shift *models.ShiftMaster) float64 {
	if shift.IsHoliday {
		return shift.ScheduledHours * 2
	}
	return shift.ScheduledHours
}

func GetMonthlyStats(year, month int) ([]models.MonthlyStats, error) {
	stats, err := repository.GetMonthlyStats(year, month)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func GetDepartmentStats(deptID, year, month int) ([]models.MonthlyStats, error) {
	stats, err := repository.GetDepartmentMonthlyStats(deptID, year, month)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

var ErrNotFound = errors.New("not found")
