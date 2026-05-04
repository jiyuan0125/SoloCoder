package common

type CreateJobRequest struct {
	Title       string         `json:"title"`
	MinSalary   int            `json:"min_salary"`
	MaxSalary   int            `json:"max_salary"`
	City        string         `json:"city"`
	Education   EducationLevel `json:"education"`
	Experience  string         `json:"experience"`
	Description string         `json:"description"`
}

type CreateJobResponse struct {
	JobID string `json:"job_id"`
}

type ListJobsRequest struct {
	MinSalary   *int           `json:"min_salary,omitempty"`
	City        *string        `json:"city,omitempty"`
	Education   *EducationLevel `json:"education,omitempty"`
}

type ListJobsResponse struct {
	Jobs []Job `json:"jobs"`
}

type ApplyJobRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Summary string `json:"summary"`
}

type ApplyJobResponse struct {
	ApplicationID string `json:"application_id"`
}

type UpdateApplicationStatusRequest struct {
	Status            ApplicationStatus `json:"status"`
	InterviewTime     *string           `json:"interview_time,omitempty"`
	InterviewLocation *string           `json:"interview_location,omitempty"`
}

type ListApplicationsRequest struct {
	JobID *string `json:"job_id,omitempty"`
}

type ListApplicationsResponse struct {
	Applications []ApplicationDetail `json:"applications"`
}

type ListMyApplicationsRequest struct {
	Phone string `json:"phone"`
}

type ListMyApplicationsResponse struct {
	Applications []ApplicationDetail `json:"applications"`
}

type OfflineJobRequest struct {
	JobID string `json:"job_id"`
}

type OfflineJobResponse struct {
	Success bool `json:"success"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
