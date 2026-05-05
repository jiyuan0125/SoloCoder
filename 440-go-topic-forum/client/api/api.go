package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"forum/common"
)

const DefaultServerURL = "http://localhost:8080"

var ServerURL = DefaultServerURL

func httpGet(path string, query url.Values) (*common.Response, error) {
	u, err := url.Parse(ServerURL + path)
	if err != nil {
		return nil, err
	}
	if query != nil {
		u.RawQuery = query.Encode()
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.Response
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("response parse error: %s", string(body))
	}

	return &result, nil
}

func httpPost(path string, body interface{}) (*common.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(ServerURL+path, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.Response
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("response parse error: %s", string(respBody))
	}

	return &result, nil
}

func CreateUser(username string, isAdmin bool) (*common.UserResponse, error) {
	req := common.CreateUserRequest{
		Username: username,
		IsAdmin:  isAdmin,
	}

	resp, err := httpPost("/users", req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var user common.UserResponse
	json.Unmarshal(dataBytes, &user)

	return &user, nil
}

func GetUser(id string, username string) (*common.UserResponse, error) {
	query := url.Values{}
	if id != "" {
		query.Set("id", id)
	}
	if username != "" {
		query.Set("username", username)
	}

	resp, err := httpGet("/users", query)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var user common.UserResponse
	json.Unmarshal(dataBytes, &user)

	return &user, nil
}

func ListUsers() ([]*common.UserResponse, error) {
	resp, err := httpGet("/users", nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var users []*common.UserResponse
	json.Unmarshal(dataBytes, &users)

	return users, nil
}

func CreatePost(userID, title, content, category string, tags []string) (*common.PostResponse, error) {
	req := common.CreatePostRequest{
		UserID:   userID,
		Title:    title,
		Content:  content,
		Category: category,
		Tags:     tags,
	}

	resp, err := httpPost("/posts", req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var post common.PostResponse
	json.Unmarshal(dataBytes, &post)

	return &post, nil
}

func GetPost(postID, viewerID string) (*common.PostDetailResponse, error) {
	query := url.Values{}
	query.Set("id", postID)
	if viewerID != "" {
		query.Set("viewer_id", viewerID)
	}

	resp, err := httpGet("/posts", query)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var detail common.PostDetailResponse
	json.Unmarshal(dataBytes, &detail)

	return &detail, nil
}

func ListPosts() ([]*common.PostResponse, error) {
	resp, err := httpGet("/posts", nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var posts []*common.PostResponse
	json.Unmarshal(dataBytes, &posts)

	return posts, nil
}

func UpdatePost(postID, userID, title, content, category string, tags []string) (*common.PostResponse, error) {
	query := url.Values{}
	query.Set("id", postID)

	req := common.UpdatePostRequest{
		UserID:   userID,
		Title:    title,
		Content:  content,
		Category: category,
		Tags:     tags,
	}

	u, _ := url.Parse(ServerURL + "/posts/update")
	u.RawQuery = query.Encode()

	jsonBody, _ := json.Marshal(req)
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result common.Response
	json.Unmarshal(body, &result)

	if !result.Success {
		return nil, fmt.Errorf(result.Message)
	}

	dataBytes, _ := json.Marshal(result.Data)
	var post common.PostResponse
	json.Unmarshal(dataBytes, &post)

	return &post, nil
}

func DeletePost(userID, postID string) error {
	req := common.DeletePostRequest{
		UserID: userID,
		PostID: postID,
	}

	resp, err := httpPost("/posts/delete", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}

	return nil
}

func SearchPosts(keyword, category, tag, userID string, limit, offset int) (*common.SearchResponse, error) {
	query := url.Values{}
	if keyword != "" {
		query.Set("keyword", keyword)
	}
	if category != "" {
		query.Set("category", category)
	}
	if tag != "" {
		query.Set("tag", tag)
	}
	if userID != "" {
		query.Set("user_id", userID)
	}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		query.Set("offset", fmt.Sprintf("%d", offset))
	}

	resp, err := httpGet("/posts/search", query)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var result common.SearchResponse
	json.Unmarshal(dataBytes, &result)

	return &result, nil
}

func CreateReply(userID, postID, content, parentID string) (*common.ReplyResponse, error) {
	req := common.CreateReplyRequest{
		UserID:   userID,
		PostID:   postID,
		Content:  content,
		ParentID: parentID,
	}

	resp, err := httpPost("/replies", req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var reply common.ReplyResponse
	json.Unmarshal(dataBytes, &reply)

	return &reply, nil
}

func SetBestReply(userID, postID, replyID string) error {
	req := common.SetBestReplyRequest{
		UserID:  userID,
		PostID:  postID,
		ReplyID: replyID,
	}

	resp, err := httpPost("/replies/best", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}

	return nil
}

func Like(userID, targetID, targetType string) error {
	req := common.LikeRequest{
		UserID:     userID,
		TargetID:   targetID,
		TargetType: targetType,
	}

	resp, err := httpPost("/likes", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}

	return nil
}

func Unlike(userID, targetID, targetType string) error {
	req := common.LikeRequest{
		UserID:     userID,
		TargetID:   targetID,
		TargetType: targetType,
	}

	resp, err := httpPost("/unlikes", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}

	return nil
}

func SetTop(adminID, postID string, isTop bool) error {
	req := common.SetTopRequest{
		AdminID: adminID,
		PostID:  postID,
		IsTop:   isTop,
	}

	resp, err := httpPost("/admin/top", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}

	return nil
}

func SetEssence(adminID, postID string, isEssence bool) error {
	req := common.SetEssenceRequest{
		AdminID:   adminID,
		PostID:    postID,
		IsEssence: isEssence,
	}

	resp, err := httpPost("/admin/essence", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}

	return nil
}

func CreateReport(reporterID, postID, reason string) (*common.ReportResponse, error) {
	req := common.ReportRequest{
		ReporterID: reporterID,
		PostID:     postID,
		Reason:     reason,
	}

	resp, err := httpPost("/reports", req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var report common.ReportResponse
	json.Unmarshal(dataBytes, &report)

	return &report, nil
}

func GetPendingReports() ([]*common.ReportResponse, error) {
	resp, err := httpGet("/reports", nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var reports []*common.ReportResponse
	json.Unmarshal(dataBytes, &reports)

	return reports, nil
}

func ReviewReport(adminID, reportID, status string) error {
	req := common.ReviewReportRequest{
		AdminID:  adminID,
		ReportID: reportID,
		Status:   status,
	}

	resp, err := httpPost("/reports/review", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}

	return nil
}

func GetLeaderboard() (*common.LeaderboardResponse, error) {
	resp, err := httpGet("/leaderboard", nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var result common.LeaderboardResponse
	json.Unmarshal(dataBytes, &result)

	return &result, nil
}

func GetTagCloud(limit int) (*common.TagCloudResponse, error) {
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}

	resp, err := httpGet("/tags", query)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var result common.TagCloudResponse
	json.Unmarshal(dataBytes, &result)

	return &result, nil
}

func CalculateHotPosts() (*common.HotPostsResponse, error) {
	resp, err := httpPost("/hotposts", map[string]string{})
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var result common.HotPostsResponse
	json.Unmarshal(dataBytes, &result)

	return &result, nil
}

func GetHotPosts() (*common.HotPostsResponse, error) {
	resp, err := httpGet("/hotposts", nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var result common.HotPostsResponse
	json.Unmarshal(dataBytes, &result)

	return &result, nil
}
