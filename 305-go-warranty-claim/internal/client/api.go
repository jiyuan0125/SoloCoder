package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"warranty-claim/pkg/common"
)

type APIClient struct {
	baseURL string
	userID  string
}

func NewAPIClient(baseURL, userID string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		userID:  userID,
	}
}

func (c *APIClient) SubmitApplication(req common.SubmitApplicationRequest) (*common.SubmitApplicationResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.baseURL+"/submit", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.SubmitApplicationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *APIClient) GetApplication(id string) (*common.WarrantyApplication, error) {
	resp, err := http.Get(fmt.Sprintf("%s/applications/%s", c.baseURL, id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("application not found")
	}

	var app common.WarrantyApplication
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return nil, err
	}

	return &app, nil
}

func (c *APIClient) GetMyApplications() ([]common.WarrantyApplication, error) {
	resp, err := http.Get(fmt.Sprintf("%s/my-applications?user_id=%s", c.baseURL, url.QueryEscape(c.userID)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.GetApplicationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Applications, nil
}

func (c *APIClient) GetAllApplications() ([]common.WarrantyApplication, error) {
	resp, err := http.Get(c.baseURL + "/all-applications")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.GetApplicationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Applications, nil
}

func (c *APIClient) GetPendingApplications() ([]common.WarrantyApplication, error) {
	resp, err := http.Get(c.baseURL + "/pending-applications")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.GetApplicationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Applications, nil
}

func (c *APIClient) ReviewApplication(req common.ReviewApplicationRequest) (bool, string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return false, "", err
	}

	resp, err := http.Post(c.baseURL+"/review", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, "", err
	}

	success, _ := result["success"].(bool)
	message, _ := result["message"].(string)
	return success, message, nil
}

func (c *APIClient) SubmitAppeal(req common.SubmitAppealRequest) (bool, string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return false, "", err
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/appeal", bytes.NewBuffer(body))
	if err != nil {
		return false, "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-ID", c.userID)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, "", err
	}

	success, _ := result["success"].(bool)
	message, _ := result["message"].(string)
	return success, message, nil
}

func (c *APIClient) GetPendingAppeals() ([]common.AppealWithApplication, error) {
	resp, err := http.Get(c.baseURL + "/appeals")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.GetAppealsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Appeals, nil
}

func (c *APIClient) GetAppealsByApplication(appID string) ([]common.AppealWithApplication, error) {
	resp, err := http.Get(fmt.Sprintf("%s/appeals?application_id=%s", c.baseURL, url.QueryEscape(appID)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.GetAppealsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Appeals, nil
}

func (c *APIClient) ResolveAppeal(appealID string, resolved bool, adminNote string) (bool, string, error) {
	req := map[string]interface{}{
		"appeal_id":  appealID,
		"resolved":   resolved,
		"admin_note": adminNote,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return false, "", err
	}

	resp, err := http.Post(c.baseURL+"/resolve-appeal", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, "", err
	}

	success, _ := result["success"].(bool)
	message, _ := result["message"].(string)
	return success, message, nil
}

func (c *APIClient) GetStatistics() (*common.StatisticsResponse, error) {
	resp, err := http.Get(c.baseURL + "/statistics")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	var stats common.StatisticsResponse
	if err := json.Unmarshal(bodyBytes, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}
