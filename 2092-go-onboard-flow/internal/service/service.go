package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"onboard-flow/internal/db"
	"onboard-flow/internal/model"
)

var (
	ErrEmployeeNotFound     = errors.New("employee not found")
	ErrInvalidStep          = errors.New("invalid step")
	ErrStepNotReady         = errors.New("previous step not completed")
	ErrProcessTerminated    = errors.New("process already terminated")
	ErrMentorCapacityFull   = errors.New("mentor has reached maximum trainees")
	ErrExamMaxAttempts      = errors.New("maximum exam attempts reached")
	ErrBackgroundCheckFail  = errors.New("background check failed")
	ErrTrainingFailed       = errors.New("training failed after max attempts")
)

type FlowService struct{}

func NewFlowService() *FlowService {
	return &FlowService{}
}

func (s *FlowService) CreateEmployee(name, email, department string, hireDate time.Time, totalAmount float64) (*model.Employee, error) {
	return db.CreateEmployee(name, email, department, hireDate, totalAmount)
}

func (s *FlowService) GetEmployee(id int64) (*model.Employee, error) {
	emp, err := db.GetEmployeeByID(id)
	if err == sql.ErrNoRows {
		return nil, ErrEmployeeNotFound
	}
	return emp, err
}

func (s *FlowService) GetEmployeeFlow(id int64) (*model.Employee, []*model.Step, error) {
	emp, err := s.GetEmployee(id)
	if err != nil {
		return nil, nil, err
	}

	steps, err := db.GetAllStepsForEmployee(id)
	if err != nil {
		return nil, nil, err
	}

	return emp, steps, nil
}

func (s *FlowService) StartStep(employeeID int64, stepName model.StepName) error {
	emp, err := s.GetEmployee(employeeID)
	if err != nil {
		return err
	}

	if emp.Status == model.StatusTerminated {
		return ErrProcessTerminated
	}

	if model.StepOrder[stepName] > model.StepOrder[emp.CurrentStep] {
		return fmt.Errorf("%w: current step is %s", ErrStepNotReady, emp.CurrentStep)
	}

	if model.StepOrder[stepName] < model.StepOrder[emp.CurrentStep] {
		return nil
	}

	step, err := db.GetStepByEmployeeAndName(employeeID, stepName)
	if err != nil {
		return err
	}

	if step.Status == model.StatusPending {
		return db.UpdateStepStatus(employeeID, stepName, model.StatusInProgress, nil)
	}

	return nil
}

func (s *FlowService) CompleteStep(employeeID int64, stepName model.StepName, passed *bool) error {
	emp, err := s.GetEmployee(employeeID)
	if err != nil {
		return err
	}

	if emp.Status == model.StatusTerminated {
		return ErrProcessTerminated
	}

	if model.StepOrder[stepName] != model.StepOrder[emp.CurrentStep] {
		return fmt.Errorf("%w: current step is %s", ErrStepNotReady, emp.CurrentStep)
	}

	if passed != nil && !*passed {
		if stepName == model.StepBackgroundCheck {
			db.TerminateEmployee(employeeID, "Background check failed")
			return ErrBackgroundCheckFail
		}
	}

	err = db.UpdateStepStatus(employeeID, stepName, model.StatusCompleted, passed)
	if err != nil {
		return err
	}

	currentIdx := -1
	for i, name := range model.StepsInOrder {
		if name == stepName {
			currentIdx = i
			break
		}
	}

	if currentIdx >= 0 && currentIdx < len(model.StepsInOrder)-1 {
		nextStep := model.StepsInOrder[currentIdx+1]
		err = db.UpdateEmployeeCurrentStep(employeeID, nextStep)
		if err != nil {
			return err
		}

		if nextStep == model.StepProbationPeriod {
			_, err := db.GetOrCreateProbationReview(employeeID)
			if err != nil {
				return err
			}
		}
	}

	s.validateFlowConsistency(employeeID)
	return nil
}

func (s *FlowService) validateFlowConsistency(employeeID int64) {
	steps, err := db.GetAllStepsForEmployee(employeeID)
	if err != nil {
		return
	}

	completedSteps := make(map[model.StepName]bool)
	for _, step := range steps {
		if step.Status == model.StatusCompleted {
			completedSteps[step.StepName] = true
		}
	}

	for i, step := range steps {
		if i > 0 && step.Status == model.StatusCompleted {
			prevStep := steps[i-1]
			if prevStep.Status != model.StatusCompleted {
				return
			}
		}
	}
}

func (s *FlowService) RecordDelay(employeeID int64, stepName model.StepName, reason string) error {
	step, err := db.GetStepByEmployeeAndName(employeeID, stepName)
	if err != nil {
		return err
	}

	err = db.RecordDelay(step.ID, employeeID, stepName, reason)
	if err != nil {
		return err
	}

	emp, err := s.GetEmployee(employeeID)
	if err != nil {
		return err
	}

	_ = emp
	return nil
}

func (s *FlowService) GetDelayRecords(employeeID int64) ([]*model.DelayRecord, error) {
	_, err := s.GetEmployee(employeeID)
	if err != nil {
		return nil, err
	}
	return db.GetDelayRecords(employeeID)
}

func (s *FlowService) UpdateTotalAmount(employeeID int64, newTotal float64) error {
	_, err := s.GetEmployee(employeeID)
	if err != nil {
		return err
	}

	err = db.ReallocateStepAmounts(employeeID, newTotal)
	if err != nil {
		return err
	}

	s.validateFlowConsistency(employeeID)
	return nil
}

func (s *FlowService) ProcessBackgroundCheck(employeeID int64, passed bool, note string) error {
	_, err := db.CreateBackgroundCheck(employeeID, passed, note)
	if err != nil {
		return err
	}

	passedVal := passed
	return s.CompleteStep(employeeID, model.StepBackgroundCheck, &passedVal)
}

func (s *FlowService) UpdateAccountSetup(employeeID int64, email, im, vpn, devEnv *bool) error {
	setup, err := db.GetOrCreateAccountSetup(employeeID)
	if err != nil {
		return err
	}

	err = db.UpdateAccountSetup(employeeID, email, im, vpn, devEnv)
	if err != nil {
		return err
	}

	setup, err = db.GetAccountSetupByID(setup.ID)
	if err != nil {
		return err
	}

	if setup.EmailDone && setup.IMDone && setup.VPNDone && setup.DevEnvDone {
		step, err := db.GetStepByEmployeeAndName(employeeID, model.StepAccountSetup)
		if err != nil {
			return err
		}
		if step.AccountSetupID == nil {
			id := setup.ID
			_, err = db.DB.Exec(`
				UPDATE steps SET account_setup_id = ?, updated_at = CURRENT_TIMESTAMP 
				WHERE employee_id = ? AND step_name = 'account_setup'
			`, id, employeeID)
			if err != nil {
				return err
			}
		}
		trueVal := true
		return s.CompleteStep(employeeID, model.StepAccountSetup, &trueVal)
	}

	return nil
}

func (s *FlowService) CreateMentor(name, employeeID, department string) (*model.Mentor, error) {
	return db.CreateMentor(name, employeeID, department)
}

func (s *FlowService) AssignMentor(employeeID int64, mentorID int64) error {
	mentor, err := db.GetMentorByID(mentorID)
	if err != nil {
		return err
	}

	emp, err := s.GetEmployee(employeeID)
	if err != nil {
		return err
	}

	if emp.Department != mentor.Department {
		return errors.New("mentor must be in the same department")
	}

	if emp.Status == model.StatusTerminated {
		return ErrProcessTerminated
	}

	if emp.CurrentStep != model.StepMentorAssign {
		return fmt.Errorf("%w: current step is %s", ErrStepNotReady, emp.CurrentStep)
	}

	currentTrainees, err := db.GetMentorCurrentTrainees(mentorID)
	if err != nil {
		return err
	}

	if currentTrainees >= 3 {
		return ErrMentorCapacityFull
	}

	step, err := db.GetStepByEmployeeAndName(employeeID, model.StepMentorAssign)
	if err != nil {
		return err
	}

	err = db.AssignMentorToStep(step.ID, mentorID)
	if err != nil {
		return err
	}

	trueVal := true
	return s.CompleteStep(employeeID, model.StepMentorAssign, &trueVal)
}

func (s *FlowService) SubmitExam(employeeID int64, examType string, score, maxScore int) error {
	if examType != "culture" && examType != "skill" {
		return errors.New("invalid exam type")
	}

	emp, err := s.GetEmployee(employeeID)
	if err != nil {
		return err
	}

	if emp.Status == model.StatusTerminated {
		return ErrProcessTerminated
	}

	if emp.CurrentStep != model.StepTrainingComplete {
		return fmt.Errorf("%w: current step is %s", ErrStepNotReady, emp.CurrentStep)
	}

	training, err := db.GetOrCreateTraining(employeeID)
	if err != nil {
		return err
	}

	latestExam, err := db.GetLatestExam(training.ID, examType)
	if err != nil {
		return err
	}

	attemptCount := 1
	if latestExam != nil {
		if latestExam.AttemptCount >= 3 {
			db.TerminateEmployee(employeeID, "Training exam failed after max attempts")
			return ErrExamMaxAttempts
		}
		attemptCount = latestExam.AttemptCount + 1
	}

	passed := score >= maxScore*60/100

	examID, err := db.CreateExam(training.ID, examType, score, maxScore, passed, attemptCount)
	if err != nil {
		return err
	}

	err = db.UpdateTrainingExam(training.ID, examType, examID)
	if err != nil {
		return err
	}

	if passed {
		cultureExam, err := db.GetLatestExam(training.ID, "culture")
		if err != nil {
			return err
		}
		skillExam, err := db.GetLatestExam(training.ID, "skill")
		if err != nil {
			return err
		}

		if cultureExam != nil && skillExam != nil && cultureExam.Passed && skillExam.Passed {
			err = db.UpdateTrainingStatus(training.ID, model.StatusCompleted)
			if err != nil {
				return err
			}
			trueVal := true
			return s.CompleteStep(employeeID, model.StepTrainingComplete, &trueVal)
		}
	}

	return nil
}

func (s *FlowService) SubmitMentorReview(employeeID int64, score int, comment string) error {
	return s.submitReview(employeeID, score, comment, "mentor")
}

func (s *FlowService) SubmitManagerReview(employeeID int64, score int, comment string) error {
	return s.submitReview(employeeID, score, comment, "manager")
}

func (s *FlowService) submitReview(employeeID int64, score int, comment string, reviewer string) error {
	emp, err := s.GetEmployee(employeeID)
	if err != nil {
		return err
	}

	if emp.Status == model.StatusTerminated {
		return ErrProcessTerminated
	}

	if emp.CurrentStep != model.StepProbationPass {
		return fmt.Errorf("%w: current step is %s", ErrStepNotReady, emp.CurrentStep)
	}

	review, err := db.GetOrCreateProbationReview(employeeID)
	if err != nil {
		return err
	}

	if reviewer == "mentor" {
		review.MentorScore = score
		review.MentorComment = comment
	} else {
		review.ManagerScore = score
		review.ManagerComment = comment
	}

	mentorHasScore := review.MentorScore > 0 || (reviewer == "mentor" && score > 0)
	managerHasScore := review.ManagerScore > 0 || (reviewer == "manager" && score > 0)

	if mentorHasScore && managerHasScore {
		review.TotalScore = review.MentorScore + review.ManagerScore
		review.Passed = review.TotalScore >= 120
		review.Completed = true
	}

	err = db.UpdateProbationReview(review)
	if err != nil {
		return err
	}

	if review.Completed {
		step, err := db.GetStepByEmployeeAndName(employeeID, model.StepProbationPass)
		if err != nil {
			return err
		}
		if step.ProbationReviewID == nil {
			id := review.ID
			_, err = db.DB.Exec(`
				UPDATE steps SET probation_review_id = ?, updated_at = CURRENT_TIMESTAMP
				WHERE employee_id = ? AND step_name = 'probation_pass'
			`, id, employeeID)
			if err != nil {
				return err
			}
		}
		return s.CompleteStep(employeeID, model.StepProbationPass, &review.Passed)
	}

	return nil
}
