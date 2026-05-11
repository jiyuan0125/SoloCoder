package core

import (
	"errors"
	"fmt"

	"github.com/drivingschool/common"
)

func (s *Store) RegisterStudent(name, idCard, phone string, vehicleType common.VehicleType) (*common.Student, error) {
	if name == "" || idCard == "" || phone == "" {
		return nil, errors.New("姓名、身份证号和手机号不能为空")
	}

	if _, exists := common.VehicleTypeNames[vehicleType]; !exists {
		return nil, errors.New("无效的报考车型")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	studentID := fmt.Sprintf("STU_%s", idCard)
	if _, exists := s.students[studentID]; exists {
		return nil, errors.New("该身份证号已注册")
	}

	student := &common.Student{
		ID:          studentID,
		Name:        name,
		IDCard:      idCard,
		Phone:       phone,
		VehicleType: vehicleType,
		SubjectStatus: map[common.Subject]common.SubjectStatus{
			common.Subject1: common.SubjectStatusNotStarted,
			common.Subject2: common.SubjectStatusNotStarted,
			common.Subject3: common.SubjectStatusNotStarted,
			common.Subject4: common.SubjectStatusNotStarted,
		},
		StudyHours: map[common.Subject]int{
			common.Subject1: 0,
			common.Subject2: 0,
			common.Subject3: 0,
			common.Subject4: 0,
		},
	}

	s.students[studentID] = student
	return student, nil
}

func (s *Store) GetStudent(studentID string) (*common.Student, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	student, exists := s.students[studentID]
	if !exists {
		return nil, errors.New("学员不存在")
	}
	return student, nil
}

func (s *Store) ListStudents() []*common.Student {
	s.mu.RLock()
	defer s.mu.RUnlock()

	students := make([]*common.Student, 0, len(s.students))
	for _, student := range s.students {
		students = append(students, student)
	}
	return students
}

func (s *Store) UpdateSubjectStatus(studentID string, subject common.Subject, status common.SubjectStatus) error {
	if subject < common.Subject1 || subject > common.Subject4 {
		return errors.New("无效的科目")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	student, exists := s.students[studentID]
	if !exists {
		return errors.New("学员不存在")
	}

	if status == common.SubjectStatusPassed && subject > common.Subject1 {
		prevSubject := subject - 1
		if student.SubjectStatus[prevSubject] != common.SubjectStatusPassed {
			return errors.New("必须先通过上一科目")
		}
	}

	student.SubjectStatus[subject] = status
	return nil
}

func (s *Store) AddStudyHours(studentID string, subject common.Subject, hours int) error {
	if subject < common.Subject1 || subject > common.Subject4 {
		return errors.New("无效的科目")
	}
	if hours <= 0 {
		return errors.New("学时必须为正数")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	student, exists := s.students[studentID]
	if !exists {
		return errors.New("学员不存在")
	}

	student.StudyHours[subject] += hours
	return nil
}

func (s *Store) GetNextSubject(studentID string) (common.Subject, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	student, exists := s.students[studentID]
	if !exists {
		return 0, errors.New("学员不存在")
	}

	for subject := common.Subject1; subject <= common.Subject4; subject++ {
		if student.SubjectStatus[subject] != common.SubjectStatusPassed {
			return subject, nil
		}
	}

	return 0, errors.New("所有科目已通过")
}

func (s *Store) CheckStudyHoursRequirement(studentID string, subject common.Subject) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	student, exists := s.students[studentID]
	if !exists {
		return errors.New("学员不存在")
	}

	if student.VehicleType == common.VehicleTypeC1 {
		switch subject {
		case common.Subject2:
			if student.StudyHours[subject] < 24 {
				return errors.New("C1车型科目二需要至少24学时")
			}
		case common.Subject3:
			if student.StudyHours[subject] < 16 {
				return errors.New("C1车型科目三需要至少16学时")
			}
		}
	}

	return nil
}
