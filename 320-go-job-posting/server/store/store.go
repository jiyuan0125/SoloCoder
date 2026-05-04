package store

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"jobposting/common"
)

type DataStore struct {
	Jobs         map[string]*common.Job
	Resumes      map[string]*common.Resume
	Applications map[string]*common.Application
	mu           sync.RWMutex
	dataFile     string
}

type storedData struct {
	Jobs         map[string]*common.Job         `json:"jobs"`
	Resumes      map[string]*common.Resume      `json:"resumes"`
	Applications map[string]*common.Application `json:"applications"`
}

func NewStore(dataFile string) *DataStore {
	s := &DataStore{
		Jobs:         make(map[string]*common.Job),
		Resumes:      make(map[string]*common.Resume),
		Applications: make(map[string]*common.Application),
		dataFile:     dataFile,
	}
	s.load()
	return s
}

func (s *DataStore) load() {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		return
	}

	var stored storedData
	if err := json.Unmarshal(data, &stored); err != nil {
		return
	}

	if stored.Jobs != nil {
		s.Jobs = stored.Jobs
	}
	if stored.Resumes != nil {
		s.Resumes = stored.Resumes
	}
	if stored.Applications != nil {
		s.Applications = stored.Applications
	}
}

func (s *DataStore) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.saveLocked()
}

func (s *DataStore) saveLocked() error {
	stored := storedData{
		Jobs:         s.Jobs,
		Resumes:      s.Resumes,
		Applications: s.Applications,
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.dataFile, data, 0644)
}

func (s *DataStore) CreateJob(req common.CreateJobRequest) (*common.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := &common.Job{
		ID:          common.GenerateID(),
		Title:       req.Title,
		MinSalary:   req.MinSalary,
		MaxSalary:   req.MaxSalary,
		City:        req.City,
		Education:   req.Education,
		Experience:  req.Experience,
		Description: req.Description,
		IsOffline:   false,
		CreatedAt:   time.Now(),
	}

	s.Jobs[job.ID] = job
	if err := s.saveLocked(); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *DataStore) GetJob(id string) (*common.Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.Jobs[id]
	return job, ok
}

func (s *DataStore) ListJobsForJobseeker(filter common.ListJobsRequest) []common.Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var jobs []common.Job
	for _, job := range s.Jobs {
		if job.IsOffline {
			continue
		}
		if filter.MinSalary != nil && *filter.MinSalary > job.MaxSalary {
			continue
		}
		if filter.City != nil && *filter.City != "" && job.City != *filter.City {
			continue
		}
		if filter.Education != nil && *filter.Education != "" && job.Education != *filter.Education {
			if job.Education != common.EducationAny {
				continue
			}
		}
		jobs = append(jobs, *job)
	}
	return jobs
}

func (s *DataStore) ListJobsForCompany() []common.Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var jobs []common.Job
	for _, job := range s.Jobs {
		jobs = append(jobs, *job)
	}
	return jobs
}

func (s *DataStore) OfflineJob(jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.Jobs[jobID]
	if !ok {
		return nil
	}
	job.IsOffline = true
	return s.saveLocked()
}

func (s *DataStore) ApplyJob(jobID string, req common.ApplyJobRequest) (*common.Application, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.Jobs[jobID]
	if !ok {
		return nil, nil
	}
	if job.IsOffline {
		return nil, nil
	}

	for _, app := range s.Applications {
		if app.JobID == jobID {
			resume := s.Resumes[app.ResumeID]
			if resume != nil && resume.Phone == req.Phone {
				return nil, nil
			}
		}
	}

	resume := &common.Resume{
		ID:        common.GenerateID(),
		Name:      req.Name,
		Phone:     req.Phone,
		Summary:   req.Summary,
		CreatedAt: time.Now(),
	}
	s.Resumes[resume.ID] = resume

	application := &common.Application{
		ID:        common.GenerateID(),
		JobID:     jobID,
		ResumeID:  resume.ID,
		Status:    common.StatusUnderReview,
		CreatedAt: time.Now(),
	}
	s.Applications[application.ID] = application

	if err := s.saveLocked(); err != nil {
		return nil, err
	}
	return application, nil
}

func (s *DataStore) GetApplication(id string) (*common.Application, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	app, ok := s.Applications[id]
	return app, ok
}

func (s *DataStore) ListApplications(jobID *string) []common.ApplicationDetail {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var apps []common.ApplicationDetail
	for _, app := range s.Applications {
		if jobID != nil && *jobID != "" && app.JobID != *jobID {
			continue
		}
		detail := common.ApplicationDetail{
			Application: *app,
		}
		if job := s.Jobs[app.JobID]; job != nil {
			detail.JobTitle = job.Title
		}
		if resume := s.Resumes[app.ResumeID]; resume != nil {
			detail.ResumeName = resume.Name
			detail.ResumePhone = resume.Phone
			detail.ResumeSummary = resume.Summary
		}
		apps = append(apps, detail)
	}
	return apps
}

func (s *DataStore) ListApplicationsByPhone(phone string) []common.ApplicationDetail {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var apps []common.ApplicationDetail
	for _, app := range s.Applications {
		resume := s.Resumes[app.ResumeID]
		if resume == nil || resume.Phone != phone {
			continue
		}
		job := s.Jobs[app.JobID]
		detail := common.ApplicationDetail{
			Application:   *app,
			JobTitle:      "",
			ResumeName:    resume.Name,
			ResumePhone:   resume.Phone,
			ResumeSummary: resume.Summary,
		}
		if job != nil {
			detail.JobTitle = job.Title
		}
		apps = append(apps, detail)
	}
	return apps
}

func (s *DataStore) UpdateApplicationStatus(appID string, req common.UpdateApplicationStatusRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	app, ok := s.Applications[appID]
	if !ok {
		return nil
	}

	app.Status = req.Status
	if req.Status == common.StatusInterview {
		app.InterviewTime = req.InterviewTime
		app.InterviewLocation = req.InterviewLocation
	} else {
		app.InterviewTime = nil
		app.InterviewLocation = nil
	}

	return s.saveLocked()
}
