package core

import (
	"errors"
	"fmt"
	"time"

	"housekeeping-training/common"
)

func (s *Store) CreateEmployer(name, contact, phone string) (*common.Employer, error) {
	if name == "" {
		return nil, errors.New("employer name is required")
	}
	if contact == "" {
		return nil, errors.New("contact person is required")
	}
	if phone == "" {
		return nil, errors.New("phone is required")
	}

	s.Lock()
	defer s.Unlock()

	s.nextEmployerID++
	employer := &common.Employer{
		ID:      fmt.Sprintf("EM%04d", s.nextEmployerID),
		Name:    name,
		Contact: contact,
		Phone:   phone,
	}
	s.employers[employer.ID] = employer
	return employer, nil
}

func (s *Store) GetEmployer(id string) (*common.Employer, error) {
	s.RLock()
	defer s.RUnlock()

	employer, ok := s.employers[id]
	if !ok {
		return nil, errors.New("employer not found")
	}
	return employer, nil
}

func (s *Store) ListEmployers() []*common.Employer {
	s.RLock()
	defer s.RUnlock()

	result := make([]*common.Employer, 0, len(s.employers))
	for _, e := range s.employers {
		result = append(result, e)
	}
	return result
}

func (s *Store) CreateJobPosting(employerID, position, requiredCert string, minSalary, maxSalary int, workLocation string) (*common.JobPosting, error) {
	if employerID == "" {
		return nil, errors.New("employer ID is required")
	}
	if position == "" {
		return nil, errors.New("position is required")
	}
	if requiredCert == "" {
		return nil, errors.New("required certificate is required")
	}
	if minSalary < 0 {
		return nil, errors.New("min salary cannot be negative")
	}
	if maxSalary < minSalary {
		return nil, errors.New("max salary cannot be less than min salary")
	}
	if workLocation == "" {
		return nil, errors.New("work location is required")
	}

	s.Lock()
	defer s.Unlock()

	if _, ok := s.employers[employerID]; !ok {
		return nil, errors.New("employer not found")
	}

	s.nextJobPostingID++
	posting := &common.JobPosting{
		ID:           fmt.Sprintf("J%04d", s.nextJobPostingID),
		EmployerID:   employerID,
		Position:     position,
		RequiredCert: requiredCert,
		MinSalary:    minSalary,
		MaxSalary:    maxSalary,
		WorkLocation: workLocation,
		PostDate:     time.Now(),
	}
	s.jobPostings[posting.ID] = posting
	return posting, nil
}

func (s *Store) GetJobPosting(id string) (*common.JobPosting, error) {
	s.RLock()
	defer s.RUnlock()

	posting, ok := s.jobPostings[id]
	if !ok {
		return nil, errors.New("job posting not found")
	}
	return posting, nil
}

func (s *Store) ListJobPostings() []*common.JobPosting {
	s.RLock()
	defer s.RUnlock()

	result := make([]*common.JobPosting, 0, len(s.jobPostings))
	for _, j := range s.jobPostings {
		result = append(result, j)
	}
	return result
}

func (s *Store) GenerateRecommendations(jobPostingID string) ([]*common.Recommendation, error) {
	if jobPostingID == "" {
		return nil, errors.New("job posting ID is required")
	}

	s.Lock()
	defer s.Unlock()

	posting, ok := s.jobPostings[jobPostingID]
	if !ok {
		return nil, errors.New("job posting not found")
	}

	var matchedStudents []string
	for _, cert := range s.certificates {
		if cert.CourseID == posting.RequiredCert {
			matchedStudents = append(matchedStudents, cert.StudentID)
		}
	}

	recommendations := make([]*common.Recommendation, 0, len(matchedStudents))
	for _, studentID := range matchedStudents {
		alreadyRecommended := false
		for _, r := range s.recommendations {
			if r.JobPostingID == jobPostingID && r.StudentID == studentID {
				alreadyRecommended = true
				break
			}
		}
		if alreadyRecommended {
			continue
		}

		s.nextRecommendationID++
		rec := &common.Recommendation{
			ID:           fmt.Sprintf("R%04d", s.nextRecommendationID),
			JobPostingID: jobPostingID,
			StudentID:    studentID,
			Consultant:   "未分配",
			Result:       common.InterviewPending,
		}
		s.recommendations[rec.ID] = rec
		recommendations = append(recommendations, rec)
	}

	return recommendations, nil
}

func (s *Store) ProcessRecommendation(recommendationID, interviewDateStr string, result common.InterviewResult, notes string) (*common.Recommendation, error) {
	if recommendationID == "" {
		return nil, errors.New("recommendation ID is required")
	}

	s.Lock()
	defer s.Unlock()

	rec, ok := s.recommendations[recommendationID]
	if !ok {
		return nil, errors.New("recommendation not found")
	}

	if interviewDateStr != "" {
		interviewDate, err := time.Parse("2006-01-02", interviewDateStr)
		if err != nil {
			return nil, errors.New("invalid interview date format, use YYYY-MM-DD")
		}
		rec.InterviewDate = interviewDate
	}

	if result != "" {
		rec.Result = result
	}

	rec.Notes = notes

	return rec, nil
}

func (s *Store) GetRecommendation(id string) (*common.Recommendation, error) {
	s.RLock()
	defer s.RUnlock()

	rec, ok := s.recommendations[id]
	if !ok {
		return nil, errors.New("recommendation not found")
	}
	return rec, nil
}

func (s *Store) ListRecommendations() []*common.Recommendation {
	s.RLock()
	defer s.RUnlock()

	result := make([]*common.Recommendation, 0, len(s.recommendations))
	for _, r := range s.recommendations {
		result = append(result, r)
	}
	return result
}

func (s *Store) ListPendingRecommendations() []*common.Recommendation {
	s.RLock()
	defer s.RUnlock()

	result := make([]*common.Recommendation, 0)
	for _, r := range s.recommendations {
		if r.Result == common.InterviewPending {
			result = append(result, r)
		}
	}
	return result
}
