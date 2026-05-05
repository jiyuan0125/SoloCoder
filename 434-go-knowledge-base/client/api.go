package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"kb/common"
	"net/http"
)

const serverURL = "http://localhost:8080"

type APIClient struct {
	UserID     string
	Department string
	IsLoggedIn bool
}

func (c *APIClient) doRequest(method, path string, body interface{}) (*common.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, serverURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *APIClient) CreateArticle(title, content, category string, tags []string, accessLevel common.AccessLevel, deptIDs []string) (*common.Response, error) {
	req := common.CreateArticleRequest{
		Title:         title,
		Content:       content,
		Category:      category,
		Tags:          tags,
		AuthorID:      c.UserID,
		AccessLevel:   accessLevel,
		DepartmentIDs: deptIDs,
	}
	return c.doRequest("POST", "/articles", req)
}

func (c *APIClient) GetArticle(articleID string) (*common.Response, error) {
	path := fmt.Sprintf("/articles/%s?user_id=%s&department=%s&is_logged_in=%t",
		articleID, c.UserID, c.Department, c.IsLoggedIn)
	return c.doRequest("GET", path, nil)
}

func (c *APIClient) UpdateArticle(articleID, title, content, category string, tags []string, accessLevel common.AccessLevel, deptIDs, relatedIDs []string) (*common.Response, error) {
	req := common.UpdateArticleRequest{
		Title:              title,
		Content:            content,
		Category:           category,
		Tags:               tags,
		AuthorID:           c.UserID,
		AccessLevel:        accessLevel,
		DepartmentIDs:      deptIDs,
		RelatedArticleIDs:  relatedIDs,
	}
	path := fmt.Sprintf("/articles/%s/update", articleID)
	return c.doRequest("POST", path, req)
}

func (c *APIClient) PublishArticle(articleID string) (*common.Response, error) {
	req := common.PublishArticleRequest{AuthorID: c.UserID}
	path := fmt.Sprintf("/articles/%s/publish", articleID)
	return c.doRequest("POST", path, req)
}

func (c *APIClient) ArchiveArticle(articleID string) (*common.Response, error) {
	req := common.ArchiveArticleRequest{OperatorID: c.UserID}
	path := fmt.Sprintf("/articles/%s/archive", articleID)
	return c.doRequest("POST", path, req)
}

func (c *APIClient) GetVersions(articleID string) (*common.Response, error) {
	path := fmt.Sprintf("/articles/%s/versions?user_id=%s", articleID, c.UserID)
	return c.doRequest("GET", path, nil)
}

func (c *APIClient) GetVersion(articleID string, versionNum int) (*common.Response, error) {
	path := fmt.Sprintf("/articles/%s/versions/%d?user_id=%s", articleID, versionNum, c.UserID)
	return c.doRequest("GET", path, nil)
}

func (c *APIClient) RollbackArticle(articleID string, versionNum int) (*common.Response, error) {
	req := common.RollbackRequest{
		VersionNum: versionNum,
		AuthorID:   c.UserID,
	}
	path := fmt.Sprintf("/articles/%s/rollback", articleID)
	return c.doRequest("POST", path, req)
}

func (c *APIClient) CompareVersions(articleID string, v1, v2 int) (*common.Response, error) {
	path := fmt.Sprintf("/articles/%s/versions/%d/compare/%d?user_id=%s", articleID, v1, v2, c.UserID)
	return c.doRequest("GET", path, nil)
}

func (c *APIClient) Search(keyword string) (*common.Response, error) {
	req := common.SearchRequest{
		Keyword:    keyword,
		UserID:     c.UserID,
		Department: c.Department,
		IsLoggedIn: c.IsLoggedIn,
	}
	return c.doRequest("POST", "/search", req)
}

func (c *APIClient) AddFavorite(articleID string) (*common.Response, error) {
	path := fmt.Sprintf("/favorites/%s?user_id=%s", articleID, c.UserID)
	return c.doRequest("POST", path, nil)
}

func (c *APIClient) RemoveFavorite(articleID string) (*common.Response, error) {
	path := fmt.Sprintf("/favorites/%s?user_id=%s", articleID, c.UserID)
	return c.doRequest("DELETE", path, nil)
}

func (c *APIClient) GetFavorites() (*common.Response, error) {
	path := fmt.Sprintf("/favorites/%s", c.UserID)
	return c.doRequest("GET", path, nil)
}

func (c *APIClient) GetHotArticles(limit int) (*common.Response, error) {
	path := fmt.Sprintf("/hot?limit=%d", limit)
	return c.doRequest("GET", path, nil)
}

func (c *APIClient) GetStats() (*common.Response, error) {
	return c.doRequest("GET", "/stats", nil)
}
