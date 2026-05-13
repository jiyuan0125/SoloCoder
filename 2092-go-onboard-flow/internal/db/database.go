package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"onboard-flow/internal/model"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_fk=1&_journal=WAL")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	DB.SetMaxOpenConns(1)

	if err = createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

func createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS employees (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			department TEXT NOT NULL,
			hire_date DATETIME NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			current_step TEXT NOT NULL DEFAULT 'offer_confirm',
			total_amount REAL NOT NULL DEFAULT 0,
			termination_note TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS steps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			employee_id INTEGER NOT NULL,
			step_name TEXT NOT NULL,
			step_order INTEGER NOT NULL,
			responsible_dept TEXT NOT NULL,
			planned_start_date DATETIME NOT NULL,
			planned_end_date DATETIME NOT NULL,
			actual_start_date DATETIME,
			actual_end_date DATETIME,
			status TEXT NOT NULL DEFAULT 'pending',
			planned_amount REAL NOT NULL DEFAULT 0,
			actual_amount REAL NOT NULL DEFAULT 0,
			delay_reason TEXT,
			passed INTEGER,
			background_check_id INTEGER,
			account_setup_id INTEGER,
			mentor_id INTEGER,
			training_id INTEGER,
			probation_review_id INTEGER,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
			FOREIGN KEY (mentor_id) REFERENCES mentors(id),
			UNIQUE(employee_id, step_name)
		)`,

		`CREATE TABLE IF NOT EXISTS mentors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			employee_id TEXT NOT NULL UNIQUE,
			department TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS background_checks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			employee_id INTEGER NOT NULL,
			passed INTEGER NOT NULL,
			note TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS account_setups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			employee_id INTEGER NOT NULL UNIQUE,
			email_done INTEGER NOT NULL DEFAULT 0,
			im_done INTEGER NOT NULL DEFAULT 0,
			vpn_done INTEGER NOT NULL DEFAULT 0,
			dev_env_done INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS trainings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			employee_id INTEGER NOT NULL UNIQUE,
			culture_exam_id INTEGER,
			skill_exam_id INTEGER,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS exams (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			training_id INTEGER NOT NULL,
			exam_type TEXT NOT NULL,
			score INTEGER NOT NULL,
			max_score INTEGER NOT NULL,
			passed INTEGER NOT NULL,
			attempt_count INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (training_id) REFERENCES trainings(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS probation_reviews (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			employee_id INTEGER NOT NULL UNIQUE,
			mentor_score INTEGER NOT NULL DEFAULT 0,
			mentor_comment TEXT,
			manager_score INTEGER NOT NULL DEFAULT 0,
			manager_comment TEXT,
			total_score INTEGER NOT NULL DEFAULT 0,
			passed INTEGER NOT NULL DEFAULT 0,
			completed INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS delay_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			step_id INTEGER NOT NULL,
			employee_id INTEGER NOT NULL,
			step_name TEXT NOT NULL,
			reason TEXT NOT NULL,
			notified INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (step_id) REFERENCES steps(id) ON DELETE CASCADE,
			FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE
		)`,

		`CREATE INDEX IF NOT EXISTS idx_steps_employee ON steps(employee_id)`,
		`CREATE INDEX IF NOT EXISTS idx_steps_employee_step ON steps(employee_id, step_name)`,
		`CREATE INDEX IF NOT EXISTS idx_delay_employee ON delay_records(employee_id)`,
		`CREATE INDEX IF NOT EXISTS idx_employees_status ON employees(status)`,
	}

	for _, query := range queries {
		_, err := DB.Exec(query)
		if err != nil {
			return fmt.Errorf("query failed: %s, error: %w", query, err)
		}
	}

	return nil
}

func CreateEmployee(name, email, department string, hireDate time.Time, totalAmount float64) (*model.Employee, error) {
	tx, err := DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO employees (name, email, department, hire_date, total_amount, status, current_step)
		VALUES (?, ?, ?, ?, ?, 'pending', 'offer_confirm')
	`, name, email, department, hireDate, totalAmount)
	if err != nil {
		return nil, err
	}

	employeeID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	plannedEnds := map[model.StepName]int{
		model.StepOfferConfirm:     1,
		model.StepBackgroundCheck:  3,
		model.StepContractSign:     2,
		model.StepWorkstationPrep:  3,
		model.StepAccountSetup:     5,
		model.StepMentorAssign:     2,
		model.StepTrainingComplete: 10,
		model.StepProbationPeriod:  90,
		model.StepProbationPass:    5,
	}

	perStepAmount := totalAmount / 9.0
	for _, stepName := range model.StepsInOrder {
		days := plannedEnds[stepName]
		startDate := now
		if stepName != model.StepOfferConfirm {
			startDate = now.AddDate(0, 0, model.StepOrder[stepName]*2)
		}
		endDate := startDate.AddDate(0, 0, days)

		_, err = tx.Exec(`
			INSERT INTO steps (employee_id, step_name, step_order, responsible_dept, 
				planned_start_date, planned_end_date, planned_amount, status)
			VALUES (?, ?, ?, ?, ?, ?, ?, 'pending')
		`, employeeID, stepName, model.StepOrder[stepName], model.StepDepartment[stepName],
			startDate, endDate, perStepAmount)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return GetEmployeeByID(employeeID)
}

func GetEmployeeByID(id int64) (*model.Employee, error) {
	row := DB.QueryRow(`
		SELECT id, name, email, department, hire_date, status, current_step, 
			total_amount, termination_note, created_at, updated_at
		FROM employees WHERE id = ?
	`, id)

	emp := &model.Employee{}
	err := row.Scan(&emp.ID, &emp.Name, &emp.Email, &emp.Department, &emp.HireDate,
		&emp.Status, &emp.CurrentStep, &emp.TotalAmount, &emp.TerminationNote,
		&emp.CreatedAt, &emp.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return emp, nil
}

func GetStepByEmployeeAndName(employeeID int64, stepName model.StepName) (*model.Step, error) {
	row := DB.QueryRow(`
		SELECT id, employee_id, step_name, step_order, responsible_dept,
			planned_start_date, planned_end_date, actual_start_date, actual_end_date,
			status, planned_amount, actual_amount, delay_reason, passed,
			background_check_id, account_setup_id, mentor_id, training_id, probation_review_id,
			created_at, updated_at
		FROM steps WHERE employee_id = ? AND step_name = ?
	`, employeeID, stepName)

	step := &model.Step{}
	err := row.Scan(&step.ID, &step.EmployeeID, &step.StepName, &step.StepOrder,
		&step.ResponsibleDept, &step.PlannedStartDate, &step.PlannedEndDate,
		&step.ActualStartDate, &step.ActualEndDate, &step.Status,
		&step.PlannedAmount, &step.ActualAmount, &step.DelayReason,
		&step.Passed, &step.BackgroundCheckID, &step.AccountSetupID,
		&step.MentorID, &step.TrainingID, &step.ProbationReviewID,
		&step.CreatedAt, &step.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return step, nil
}

func GetAllStepsForEmployee(employeeID int64) ([]*model.Step, error) {
	rows, err := DB.Query(`
		SELECT id, employee_id, step_name, step_order, responsible_dept,
			planned_start_date, planned_end_date, actual_start_date, actual_end_date,
			status, planned_amount, actual_amount, delay_reason, passed,
			background_check_id, account_setup_id, mentor_id, training_id, probation_review_id,
			created_at, updated_at
		FROM steps WHERE employee_id = ? ORDER BY step_order ASC
	`, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	steps := []*model.Step{}
	for rows.Next() {
		step := &model.Step{}
		err := rows.Scan(&step.ID, &step.EmployeeID, &step.StepName, &step.StepOrder,
			&step.ResponsibleDept, &step.PlannedStartDate, &step.PlannedEndDate,
			&step.ActualStartDate, &step.ActualEndDate, &step.Status,
			&step.PlannedAmount, &step.ActualAmount, &step.DelayReason,
			&step.Passed, &step.BackgroundCheckID, &step.AccountSetupID,
			&step.MentorID, &step.TrainingID, &step.ProbationReviewID,
			&step.CreatedAt, &step.UpdatedAt)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	return steps, nil
}

func UpdateStepStatus(employeeID int64, stepName model.StepName, status model.StepStatus, passed *bool) error {
	now := time.Now()

	var startDate *time.Time
	if status == model.StatusInProgress {
		startDate = &now
	}

	var endDate *time.Time
	if status == model.StatusCompleted {
		endDate = &now
	}

	query := `
		UPDATE steps SET status = ?, updated_at = ?
	`
	args := []interface{}{status, now}

	if startDate != nil {
		query += `, actual_start_date = ?`
		args = append(args, *startDate)
	}

	if endDate != nil {
		query += `, actual_end_date = ?`
		args = append(args, *endDate)
	}

	if passed != nil {
		query += `, passed = ?`
		args = append(args, *passed)
	}

	query += ` WHERE employee_id = ? AND step_name = ?`
	args = append(args, employeeID, stepName)

	_, err := DB.Exec(query, args...)
	return err
}

func UpdateEmployeeCurrentStep(employeeID int64, stepName model.StepName) error {
	_, err := DB.Exec(`
		UPDATE employees SET current_step = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, stepName, employeeID)
	return err
}

func TerminateEmployee(employeeID int64, note string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE employees SET status = 'terminated', termination_note = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, note, employeeID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE steps SET status = 'terminated', updated_at = CURRENT_TIMESTAMP WHERE employee_id = ? AND status NOT IN ('completed', 'terminated')
	`, employeeID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func CreateMentor(name, employeeID, department string) (*model.Mentor, error) {
	result, err := DB.Exec(`
		INSERT INTO mentors (name, employee_id, department)
		VALUES (?, ?, ?)
	`, name, employeeID, department)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetMentorByID(id)
}

func GetMentorByID(id int64) (*model.Mentor, error) {
	row := DB.QueryRow(`
		SELECT id, name, employee_id, department, created_at FROM mentors WHERE id = ?
	`, id)

	mentor := &model.Mentor{}
	err := row.Scan(&mentor.ID, &mentor.Name, &mentor.EmployeeID, &mentor.Department, &mentor.CreatedAt)
	if err != nil {
		return nil, err
	}
	return mentor, nil
}

func GetMentorCurrentTrainees(mentorID int64) (int, error) {
	var count int
	row := DB.QueryRow(`
		SELECT COUNT(*) FROM steps 
		WHERE mentor_id = ? AND status = 'in_progress'
		AND step_name = 'mentor_assign'
	`, mentorID)

	err := row.Scan(&count)
	return count, err
}

func AssignMentorToStep(stepID int64, mentorID int64) error {
	_, err := DB.Exec(`
		UPDATE steps SET mentor_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, mentorID, stepID)
	return err
}

func CreateBackgroundCheck(employeeID int64, passed bool, note string) (int64, error) {
	result, err := DB.Exec(`
		INSERT INTO background_checks (employee_id, passed, note)
		VALUES (?, ?, ?)
	`, employeeID, passed, note)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	_, err = DB.Exec(`
		UPDATE steps SET background_check_id = ? WHERE employee_id = ? AND step_name = 'background_check'
	`, id, employeeID)
	return id, err
}

func GetOrCreateAccountSetup(employeeID int64) (*model.AccountSetup, error) {
	setup := &model.AccountSetup{}
	err := DB.QueryRow(`
		SELECT id, employee_id, email_done, im_done, vpn_done, dev_env_done, created_at, updated_at
		FROM account_setups WHERE employee_id = ?
	`, employeeID).Scan(&setup.ID, &setup.EmployeeID, &setup.EmailDone, &setup.IMDone,
		&setup.VPNDone, &setup.DevEnvDone, &setup.CreatedAt, &setup.UpdatedAt)

	if err == sql.ErrNoRows {
		result, err := DB.Exec(`
			INSERT INTO account_setups (employee_id) VALUES (?)
		`, employeeID)
		if err != nil {
			return nil, err
		}
		id, _ := result.LastInsertId()
		return GetAccountSetupByID(id)
	}

	return setup, err
}

func GetAccountSetupByID(id int64) (*model.AccountSetup, error) {
	setup := &model.AccountSetup{}
	err := DB.QueryRow(`
		SELECT id, employee_id, email_done, im_done, vpn_done, dev_env_done, created_at, updated_at
		FROM account_setups WHERE id = ?
	`, id).Scan(&setup.ID, &setup.EmployeeID, &setup.EmailDone, &setup.IMDone,
		&setup.VPNDone, &setup.DevEnvDone, &setup.CreatedAt, &setup.UpdatedAt)
	return setup, err
}

func UpdateAccountSetup(employeeID int64, email, im, vpn, devEnv *bool) error {
	query := `UPDATE account_setups SET updated_at = CURRENT_TIMESTAMP`
	args := []interface{}{}

	if email != nil {
		query += `, email_done = ?`
		args = append(args, *email)
	}
	if im != nil {
		query += `, im_done = ?`
		args = append(args, *im)
	}
	if vpn != nil {
		query += `, vpn_done = ?`
		args = append(args, *vpn)
	}
	if devEnv != nil {
		query += `, dev_env_done = ?`
		args = append(args, *devEnv)
	}

	query += ` WHERE employee_id = ?`
	args = append(args, employeeID)

	_, err := DB.Exec(query, args...)
	return err
}

func GetOrCreateTraining(employeeID int64) (*model.Training, error) {
	training := &model.Training{}
	err := DB.QueryRow(`
		SELECT id, employee_id, culture_exam_id, skill_exam_id, status, created_at, updated_at
		FROM trainings WHERE employee_id = ?
	`, employeeID).Scan(&training.ID, &training.EmployeeID, &training.CultureExamID,
		&training.SkillExamID, &training.Status, &training.CreatedAt, &training.UpdatedAt)

	if err == sql.ErrNoRows {
		result, err := DB.Exec(`
			INSERT INTO trainings (employee_id, status) VALUES (?, 'in_progress')
		`, employeeID)
		if err != nil {
			return nil, err
		}
		id, _ := result.LastInsertId()
		return GetTrainingByID(id)
	}

	return training, err
}

func GetTrainingByID(id int64) (*model.Training, error) {
	training := &model.Training{}
	err := DB.QueryRow(`
		SELECT id, employee_id, culture_exam_id, skill_exam_id, status, created_at, updated_at
		FROM trainings WHERE id = ?
	`, id).Scan(&training.ID, &training.EmployeeID, &training.CultureExamID,
		&training.SkillExamID, &training.Status, &training.CreatedAt, &training.UpdatedAt)
	return training, err
}

func GetLatestExam(trainingID int64, examType string) (*model.Exam, error) {
	exam := &model.Exam{}
	err := DB.QueryRow(`
		SELECT id, training_id, exam_type, score, max_score, passed, attempt_count, created_at
		FROM exams 
		WHERE training_id = ? AND exam_type = ?
		ORDER BY attempt_count DESC, id DESC LIMIT 1
	`, trainingID, examType).Scan(&exam.ID, &exam.TrainingID, &exam.ExamType,
		&exam.Score, &exam.MaxScore, &exam.Passed, &exam.AttemptCount, &exam.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return exam, err
}

func CreateExam(trainingID int64, examType string, score, maxScore int, passed bool, attemptCount int) (int64, error) {
	result, err := DB.Exec(`
		INSERT INTO exams (training_id, exam_type, score, max_score, passed, attempt_count)
		VALUES (?, ?, ?, ?, ?, ?)
	`, trainingID, examType, score, maxScore, passed, attemptCount)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func UpdateTrainingExam(trainingID int64, examType string, examID int64) error {
	if examType == "culture" {
		_, err := DB.Exec(`
			UPDATE trainings SET culture_exam_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
		`, examID, trainingID)
		return err
	}
	_, err := DB.Exec(`
		UPDATE trainings SET skill_exam_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, examID, trainingID)
	return err
}

func UpdateTrainingStatus(trainingID int64, status model.StepStatus) error {
	_, err := DB.Exec(`
		UPDATE trainings SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, status, trainingID)
	return err
}

func GetOrCreateProbationReview(employeeID int64) (*model.ProbationReview, error) {
	review := &model.ProbationReview{}
	err := DB.QueryRow(`
		SELECT id, employee_id, mentor_score, mentor_comment, manager_score, manager_comment,
			total_score, passed, completed, created_at, updated_at
		FROM probation_reviews WHERE employee_id = ?
	`, employeeID).Scan(&review.ID, &review.EmployeeID, &review.MentorScore,
		&review.MentorComment, &review.ManagerScore, &review.ManagerComment,
		&review.TotalScore, &review.Passed, &review.Completed,
		&review.CreatedAt, &review.UpdatedAt)

	if err == sql.ErrNoRows {
		result, err := DB.Exec(`
			INSERT INTO probation_reviews (employee_id) VALUES (?)
		`, employeeID)
		if err != nil {
			return nil, err
		}
		id, _ := result.LastInsertId()
		return GetProbationReviewByID(id)
	}

	return review, err
}

func GetProbationReviewByID(id int64) (*model.ProbationReview, error) {
	review := &model.ProbationReview{}
	err := DB.QueryRow(`
		SELECT id, employee_id, mentor_score, mentor_comment, manager_score, manager_comment,
			total_score, passed, completed, created_at, updated_at
		FROM probation_reviews WHERE id = ?
	`, id).Scan(&review.ID, &review.EmployeeID, &review.MentorScore,
		&review.MentorComment, &review.ManagerScore, &review.ManagerComment,
		&review.TotalScore, &review.Passed, &review.Completed,
		&review.CreatedAt, &review.UpdatedAt)
	return review, err
}

func UpdateProbationReview(review *model.ProbationReview) error {
	_, err := DB.Exec(`
		UPDATE probation_reviews SET
			mentor_score = ?, mentor_comment = ?,
			manager_score = ?, manager_comment = ?,
			total_score = ?, passed = ?, completed = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, review.MentorScore, review.MentorComment,
		review.ManagerScore, review.ManagerComment,
		review.TotalScore, review.Passed, review.Completed, review.ID)
	return err
}

func RecordDelay(stepID, employeeID int64, stepName model.StepName, reason string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO delay_records (step_id, employee_id, step_name, reason)
		VALUES (?, ?, ?, ?)
	`, stepID, employeeID, stepName, reason)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE steps SET delay_reason = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, reason, stepID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetDelayRecords(employeeID int64) ([]*model.DelayRecord, error) {
	rows, err := DB.Query(`
		SELECT id, step_id, employee_id, step_name, reason, notified, created_at
		FROM delay_records WHERE employee_id = ? ORDER BY created_at DESC
	`, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := []*model.DelayRecord{}
	for rows.Next() {
		rec := &model.DelayRecord{}
		err := rows.Scan(&rec.ID, &rec.StepID, &rec.EmployeeID, &rec.StepName,
			&rec.Reason, &rec.Notified, &rec.CreatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}

func ReallocateStepAmounts(employeeID int64, newTotal float64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE employees SET total_amount = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, newTotal, employeeID)
	if err != nil {
		return err
	}

	perStep := newTotal / 9.0
	_, err = tx.Exec(`
		UPDATE steps SET planned_amount = ?, updated_at = CURRENT_TIMESTAMP WHERE employee_id = ?
	`, perStep, employeeID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetEmployeesForProbationReminder() ([]*model.Employee, error) {
	oneWeekLater := time.Now().AddDate(0, 0, 7)
	threeMonthsAgo := oneWeekLater.AddDate(0, -3, 0)

	rows, err := DB.Query(`
		SELECT id, name, email, department, hire_date, status, current_step, 
			total_amount, termination_note, created_at, updated_at
		FROM employees 
		WHERE status = 'pending' 
		AND current_step = 'probation_period'
		AND DATE(hire_date) = DATE(?)
	`, threeMonthsAgo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	employees := []*model.Employee{}
	for rows.Next() {
		emp := &model.Employee{}
		err := rows.Scan(&emp.ID, &emp.Name, &emp.Email, &emp.Department, &emp.HireDate,
			&emp.Status, &emp.CurrentStep, &emp.TotalAmount, &emp.TerminationNote,
			&emp.CreatedAt, &emp.UpdatedAt)
		if err != nil {
			return nil, err
		}
		employees = append(employees, emp)
	}
	return employees, nil
}
