package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-faq-service/pkg/protocol"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type APIClient struct {
	BaseURL string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{BaseURL: baseURL}
}

func (c *APIClient) doRequest(method, path string, body interface{}) (*protocol.Response, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
	}

	fullURL := c.BaseURL + path

	req, err := http.NewRequest(method, fullURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response protocol.Response
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

func (c *APIClient) HealthCheck() (*protocol.Response, error) {
	return c.doRequest("GET", "/api/v1/health", nil)
}

func (c *APIClient) CreateFAQ(categoryID string, content map[protocol.Language]protocol.FAQContent) (*protocol.Response, error) {
	req := protocol.CreateFAQRequest{
		CategoryID: categoryID,
		Content:    content,
	}
	return c.doRequest("POST", "/api/v1/faqs", req)
}

func (c *APIClient) GetFAQ(id string, lang protocol.Language) (*protocol.Response, error) {
	path := fmt.Sprintf("/api/v1/faqs/%s?lang=%s", id, lang)
	return c.doRequest("GET", path, nil)
}

func (c *APIClient) UpdateFAQ(id string, content map[protocol.Language]protocol.FAQContent) (*protocol.Response, error) {
	req := protocol.UpdateFAQRequest{
		Content: content,
	}
	path := fmt.Sprintf("/api/v1/faqs/%s", id)
	return c.doRequest("PUT", path, req)
}

func (c *APIClient) DeleteFAQ(id string) (*protocol.Response, error) {
	path := fmt.Sprintf("/api/v1/faqs/%s", id)
	return c.doRequest("DELETE", path, nil)
}

func (c *APIClient) SetFAQEnabled(id string, isEnabled bool) (*protocol.Response, error) {
	req := protocol.EnableFAQRequest{IsEnabled: isEnabled}
	path := fmt.Sprintf("/api/v1/faqs/%s/enable", id)
	return c.doRequest("POST", path, req)
}

func (c *APIClient) SetFAQPinned(id string, isPinned bool) (*protocol.Response, error) {
	req := protocol.PinFAQRequest{IsPinned: isPinned}
	path := fmt.Sprintf("/api/v1/faqs/%s/pin", id)
	return c.doRequest("POST", path, req)
}

func (c *APIClient) SearchFAQ(keyword string, lang protocol.Language, page, pageSize int) (*protocol.Response, error) {
	params := url.Values{}
	params.Set("keyword", keyword)
	params.Set("lang", string(lang))
	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		params.Set("page_size", strconv.Itoa(pageSize))
	}
	path := "/api/v1/faqs/search?" + params.Encode()
	return c.doRequest("GET", path, nil)
}

func (c *APIClient) BatchImportFAQ(items []protocol.BatchImportFAQItem) (*protocol.Response, error) {
	req := protocol.BatchImportFAQRequest{Items: items}
	return c.doRequest("POST", "/api/v1/faqs/batch-import", req)
}

func (c *APIClient) RecordClick(id string, userID string, isHelpful bool) (*protocol.Response, error) {
	req := protocol.RecordClickRequest{
		UserID:    userID,
		IsHelpful: isHelpful,
	}
	path := fmt.Sprintf("/api/v1/faqs/%s/click", id)
	return c.doRequest("POST", path, req)
}

func (c *APIClient) GetFAQHistory(id string) (*protocol.Response, error) {
	path := fmt.Sprintf("/api/v1/faqs/%s/history", id)
	return c.doRequest("GET", path, nil)
}

func (c *APIClient) CreateCategory(name string, parentID string) (*protocol.Response, error) {
	req := protocol.CreateCategoryRequest{
		ParentID: parentID,
		Name:     name,
	}
	return c.doRequest("POST", "/api/v1/categories", req)
}

func (c *APIClient) GetCategoryTree() (*protocol.Response, error) {
	return c.doRequest("GET", "/api/v1/categories/tree", nil)
}

func (c *APIClient) UpdateCategory(id string, name string) (*protocol.Response, error) {
	req := protocol.UpdateCategoryRequest{Name: name}
	path := fmt.Sprintf("/api/v1/categories/%s", id)
	return c.doRequest("PUT", path, req)
}

func (c *APIClient) DeleteCategory(id string) (*protocol.Response, error) {
	path := fmt.Sprintf("/api/v1/categories/%s", id)
	return c.doRequest("DELETE", path, nil)
}

func (c *APIClient) GetStatistics() (*protocol.Response, error) {
	return c.doRequest("GET", "/api/v1/statistics", nil)
}

func (c *APIClient) ListAllFAQs() (*protocol.Response, error) {
	return c.doRequest("GET", "/api/v1/faqs", nil)
}

func (c *APIClient) ListAllCategories() (*protocol.Response, error) {
	return c.doRequest("GET", "/api/v1/categories", nil)
}

func ParseLanguage(lang string) protocol.Language {
	switch strings.ToLower(lang) {
	case "en", "english":
		return protocol.LanguageEnglish
	default:
		return protocol.LanguageChinese
	}
}

func ParseContentMap(questions, answers []string, langs []string) (map[protocol.Language]protocol.FAQContent, error) {
	if len(questions) == 0 {
		return nil, fmt.Errorf("at least one question is required")
	}
	if len(answers) == 0 {
		return nil, fmt.Errorf("at least one answer is required")
	}

	if len(langs) == 0 {
		langs = []string{"zh"}
	}

	content := make(map[protocol.Language]protocol.FAQContent)

	for i, lang := range langs {
		l := ParseLanguage(lang)
		q := ""
		if i < len(questions) {
			q = questions[i]
		} else if len(questions) > 0 {
			q = questions[0]
		}

		a := ""
		if i < len(answers) {
			a = answers[i]
		} else if len(answers) > 0 {
			a = answers[0]
		}

		content[l] = protocol.FAQContent{
			Question: q,
			Answer:   a,
		}
	}

	return content, nil
}
