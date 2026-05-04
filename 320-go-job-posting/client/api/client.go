package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"jobposting/common"
)

const baseURL = "http://localhost:8080"

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{},
	}
}

func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var errResp common.ErrorResponse
		json.Unmarshal(data, &errResp)
		if errResp.Error != "" {
			return nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	return data, nil
}

func (c *Client) CreateJob(req common.CreateJobRequest) (string, error) {
	data, err := c.doRequest(http.MethodPost, "/company/jobs", req)
	if err != nil {
		return "", err
	}

	var resp common.CreateJobResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	return resp.JobID, nil
}

func (c *Client) ListJobsForCompany() ([]common.Job, error) {
	data, err := c.doRequest(http.MethodGet, "/company/jobs", nil)
	if err != nil {
		return nil, err
	}

	var resp common.ListJobsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Jobs, nil
}

func (c *Client) OfflineJob(jobID string) error {
	_, err := c.doRequest(http.MethodPost, "/company/jobs/"+jobID, nil)
	return err
}

func (c *Client) ListApplications(jobID *string) ([]common.ApplicationDetail, error) {
	path := "/company/applications"
	if jobID != nil && *jobID != "" {
		path += "?job_id=" + url.QueryEscape(*jobID)
	}

	data, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp common.ListApplicationsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Applications, nil
}

func (c *Client) UpdateApplicationStatus(appID string, req common.UpdateApplicationStatusRequest) error {
	_, err := c.doRequest(http.MethodPost, "/company/applications/"+appID, req)
	return err
}

func (c *Client) ListJobsForJobseeker(minSalary *int, city *string, education *common.EducationLevel) ([]common.Job, error) {
	query := url.Values{}
	if minSalary != nil {
		query.Set("min_salary", strconv.Itoa(*minSalary))
	}
	if city != nil && *city != "" {
		query.Set("city", *city)
	}
	if education != nil && *education != "" {
		query.Set("education", string(*education))
	}

	path := "/jobs"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	data, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp common.ListJobsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Jobs, nil
}

func (c *Client) ApplyJob(jobID string, req common.ApplyJobRequest) (string, error) {
	data, err := c.doRequest(http.MethodPost, "/jobs/"+jobID, req)
	if err != nil {
		return "", err
	}

	var resp common.ApplyJobResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	return resp.ApplicationID, nil
}

func (c *Client) ListMyApplications(phone string) ([]common.ApplicationDetail, error) {
	path := "/my/applications?phone=" + url.QueryEscape(phone)
	data, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp common.ListMyApplicationsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Applications, nil
}
