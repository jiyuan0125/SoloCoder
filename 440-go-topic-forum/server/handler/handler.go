package handler

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"time"

	"forum/common"
	"forum/server/store"
)

type Handler struct {
	store *store.Store
}

func NewHandler(s *store.Store) *Handler {
	return &Handler{store: s}
}

func (h *Handler) jsonResponse(w http.ResponseWriter, success bool, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	resp := common.Response{
		Success: success,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) errorResponse(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	resp := common.Response{
		Success: false,
		Message: message,
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		h.errorResponse(w, "username is required", http.StatusBadRequest)
		return
	}

	user, err := h.store.CreateUser(req.Username, req.IsAdmin)
	if err != nil {
		h.errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.jsonResponse(w, true, "user created", h.toUserResponse(user))
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	username := r.URL.Query().Get("username")

	var user *common.User
	var exists bool

	if id != "" {
		user, exists = h.store.GetUser(id)
	} else if username != "" {
		user, exists = h.store.GetUserByUsername(username)
	} else {
		h.errorResponse(w, "id or username required", http.StatusBadRequest)
		return
	}

	if !exists {
		h.errorResponse(w, "user not found", http.StatusNotFound)
		return
	}

	h.jsonResponse(w, true, "", h.toUserResponse(user))
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	users := h.store.ListUsers()
	userResponses := make([]*common.UserResponse, 0, len(users))
	for _, u := range users {
		userResponses = append(userResponses, h.toUserResponse(u))
	}

	h.jsonResponse(w, true, "", userResponses)
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		h.errorResponse(w, "user_id is required", http.StatusBadRequest)
		return
	}

	user, exists := h.store.GetUser(req.UserID)
	if !exists {
		h.errorResponse(w, "user not found", http.StatusNotFound)
		return
	}

	dailyCount := h.store.GetDailyPostCount(req.UserID, time.Now())
	if dailyCount >= common.MaxPostsPerDay {
		h.errorResponse(w, "daily post limit reached", http.StatusBadRequest)
		return
	}

	post, err := h.store.CreatePost(&req, user)
	if err != nil {
		h.errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.store.IncrementDailyPostCount(req.UserID, time.Now())

	h.jsonResponse(w, true, "post created", h.toPostResponse(post))
}

func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	userID := r.URL.Query().Get("viewer_id")

	if id == "" {
		h.errorResponse(w, "id is required", http.StatusBadRequest)
		return
	}

	post, exists := h.store.GetPost(id)
	if !exists {
		h.errorResponse(w, "post not found", http.StatusNotFound)
		return
	}

	if userID != "" {
		h.store.RecordView(id, userID)
	}

	replies, _ := h.store.GetRepliesByPost(id)
	history := h.store.GetPostHistory(id)

	replyResponses := h.buildReplyTree(replies, post.BestReplyID)
	historyResponses := make([]*common.PostHistoryResponse, 0, len(history))
	for _, ph := range history {
		historyResponses = append(historyResponses, h.toPostHistoryResponse(ph))
	}

	detail := common.PostDetailResponse{
		Post:    *h.toPostResponse(post),
		Replies: replyResponses,
		History: historyResponses,
	}

	h.jsonResponse(w, true, "", detail)
}

func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		h.errorResponse(w, "id is required", http.StatusBadRequest)
		return
	}

	var req common.UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		h.errorResponse(w, "user_id is required", http.StatusBadRequest)
		return
	}

	user, exists := h.store.GetUser(req.UserID)
	if !exists {
		h.errorResponse(w, "user not found", http.StatusNotFound)
		return
	}

	post, err := h.store.UpdatePost(id, &req, user)
	if err != nil {
		h.errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.jsonResponse(w, true, "post updated", h.toPostResponse(post))
}

func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.DeletePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" || req.PostID == "" {
		h.errorResponse(w, "user_id and post_id are required", http.StatusBadRequest)
		return
	}

	user, exists := h.store.GetUser(req.UserID)
	if !exists {
		h.errorResponse(w, "user not found", http.StatusNotFound)
		return
	}

	err := h.store.DeletePost(req.PostID, user)
	if err != nil {
		h.errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.jsonResponse(w, true, "post deleted", nil)
}

func (h *Handler) ListPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	posts := h.store.ListAllPosts()
	postResponses := make([]*common.PostResponse, 0, len(posts))
	for _, p := range posts {
		postResponses = append(postResponses, h.toPostResponse(p))
	}

	h.jsonResponse(w, true, "", postResponses)
}

func (h *Handler) CreateReply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.CreateReplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" || req.PostID == "" {
		h.errorResponse(w, "user_id and post_id are required", http.StatusBadRequest)
		return
	}

	user, exists := h.store.GetUser(req.UserID)
	if !exists {
		h.errorResponse(w, "user not found", http.StatusNotFound)
		return
	}

	reply, err := h.store.CreateReply(&req, user)
	if err != nil {
		h.errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.jsonResponse(w, true, "reply created", h.toReplyResponse(reply))
}

func (h *Handler) SetBestReply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.SetBestReplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" || req.PostID == "" || req.ReplyID == "" {
		h.errorResponse(w, "user_id, post_id and reply_id are required", http.StatusBadRequest)
		return
	}

	user, exists := h.store.GetUser(req.UserID)
	if !exists {
		h.errorResponse(w, "user not found", http.StatusNotFound)
		return
	}

	err := h.store.SetBestReply(req.PostID, req.ReplyID, user)
	if err != nil {
		h.errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.jsonResponse(w, true, "best reply set", nil)
}

func (h *Handler) Like(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.LikeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" || req.TargetID == "" {
		h.errorResponse(w, "user_id and target_id are required", http.StatusBadRequest)
		return
	}

	user, exists := h.store.GetUser(req.UserID)
	if !exists {
		h.errorResponse(w, "user not found", http.StatusNotFound)
		return
	}

	err := h.store.Like(&req, user)
	if err != nil {
		h.errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.jsonResponse(w, true, "liked", nil)
}

func (h *Handler) Unlike(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.LikeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" || req.TargetID == "" {
		h.errorResponse(w, "user_id and target_id are required", http.StatusBadRequest)
		return
	}

	user, exists := h.store.GetUser(req.UserID)
	if !exists {
		h.errorResponse(w, "user not found", http.StatusNotFound)
		return
	}

	err := h.store.Unlike(&req, user)
	if err != nil {
		h.errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.jsonResponse(w, true, "unliked", nil)
}

func (h *Handler) SetTop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.SetTopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.AdminID == "" || req.PostID == "" {
		h.errorResponse(w, "admin_id and post_id are required", http.StatusBadRequest)
		return
	}

	err := h.store.SetTop(req.AdminID, req.PostID, req.IsTop)
	if err != nil {
		h.errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.jsonResponse(w, true, "top status updated", nil)
}

func (h *Handler) SetEssence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.SetEssenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.AdminID == "" || req.PostID == "" {
		h.errorResponse(w, "admin_id and post_id are required", http.StatusBadRequest)
		return
	}

	err := h.store.SetEssence(req.AdminID, req.PostID, req.IsEssence)
	if err != nil {
		h.errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.jsonResponse(w, true, "essence status updated", nil)
}

func (h *Handler) CreateReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.ReporterID == "" || req.PostID == "" {
		h.errorResponse(w, "reporter_id and post_id are required", http.StatusBadRequest)
		return
	}

	report, err := h.store.CreateReport(&req)
	if err != nil {
		h.errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.jsonResponse(w, true, "report created", h.toReportResponse(report))
}

func (h *Handler) ReviewReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ReviewReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.AdminID == "" || req.ReportID == "" {
		h.errorResponse(w, "admin_id and report_id are required", http.StatusBadRequest)
		return
	}

	err := h.store.ReviewReport(&req)
	if err != nil {
		h.errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.jsonResponse(w, true, "report reviewed", nil)
}

func (h *Handler) GetPendingReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reports := h.store.GetReportsByStatus(common.ReportStatusPending)
	reportResponses := make([]*common.ReportResponse, 0, len(reports))
	for _, r := range reports {
		reportResponses = append(reportResponses, h.toReportResponse(r))
	}

	h.jsonResponse(w, true, "", reportResponses)
}

func (h *Handler) SearchPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	req := common.SearchRequest{
		Keyword:  r.URL.Query().Get("keyword"),
		Category: r.URL.Query().Get("category"),
		Tag:      r.URL.Query().Get("tag"),
		UserID:   r.URL.Query().Get("user_id"),
		Limit:    limit,
		Offset:   offset,
	}

	posts := h.store.SearchPosts(&req)
	postResponses := make([]*common.PostResponse, 0, len(posts))
	for _, p := range posts {
		postResponses = append(postResponses, h.toPostResponse(p))
	}

	searchResp := common.SearchResponse{
		Total: len(postResponses),
		Posts: postResponses,
	}

	h.jsonResponse(w, true, "", searchResp)
}

func (h *Handler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	users := h.store.GetLeaderboard()
	ranks := make([]common.UserRank, 0, len(users))
	for i, u := range users {
		ranks = append(ranks, common.UserRank{
			Rank:      i + 1,
			UserID:    u.ID,
			Username:  u.Username,
			LikeCount: u.LikeCount,
		})
	}

	leaderboard := common.LeaderboardResponse{
		Users: ranks,
	}

	h.jsonResponse(w, true, "", leaderboard)
}

func (h *Handler) GetTagCloud(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}

	tags := h.store.GetTagCloud(limit)
	tagCloud := common.TagCloudResponse{
		Tags: tags,
	}

	h.jsonResponse(w, true, "", tagCloud)
}

func (h *Handler) CalculateHotPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hotPosts := h.store.CalculateHotPosts(time.Now())
	details := make([]common.HotPostWithDetail, 0, len(hotPosts))
	for _, hp := range hotPosts {
		post, exists := h.store.GetPost(hp.PostID)
		if exists {
			details = append(details, common.HotPostWithDetail{
				PostID:   hp.PostID,
				Title:    post.Title,
				Username: post.Username,
				Score:    hp.Score,
			})
		}
	}

	resp := common.HotPostsResponse{
		Date:  common.FormatDate(time.Now()),
		Posts: details,
	}

	h.jsonResponse(w, true, "", resp)
}

func (h *Handler) GetHotPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.errorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hotPosts := h.store.GetHotPosts(time.Now())
	if hotPosts == nil {
		hotPosts = h.store.CalculateHotPosts(time.Now())
	}

	details := make([]common.HotPostWithDetail, 0, len(hotPosts))
	for _, hp := range hotPosts {
		post, exists := h.store.GetPost(hp.PostID)
		if exists {
			details = append(details, common.HotPostWithDetail{
				PostID:   hp.PostID,
				Title:    post.Title,
				Username: post.Username,
				Score:    hp.Score,
			})
		}
	}

	resp := common.HotPostsResponse{
		Date:  common.FormatDate(time.Now()),
		Posts: details,
	}

	h.jsonResponse(w, true, "", resp)
}

func (h *Handler) toUserResponse(u *common.User) *common.UserResponse {
	return &common.UserResponse{
		ID:         u.ID,
		Username:   u.Username,
		IsAdmin:    u.IsAdmin,
		PostCount:  u.PostCount,
		ReplyCount: u.ReplyCount,
		LikeCount:  u.LikeCount,
		CreatedAt:  common.FormatTime(u.CreatedAt),
	}
}

func (h *Handler) toPostResponse(p *common.Post) *common.PostResponse {
	return &common.PostResponse{
		ID:          p.ID,
		UserID:      p.UserID,
		Username:    p.Username,
		Title:       p.Title,
		Content:     p.Content,
		Category:    p.Category,
		Tags:        append([]string{}, p.Tags...),
		ViewCount:   p.ViewCount,
		ReplyCount:  p.ReplyCount,
		LikeCount:   p.LikeCount,
		IsTop:       p.IsTop,
		IsEssence:   p.IsEssence,
		BestReplyID: p.BestReplyID,
		CreatedAt:   common.FormatTime(p.CreatedAt),
		UpdatedAt:   common.FormatTime(p.UpdatedAt),
	}
}

func (h *Handler) toReplyResponse(r *common.Reply) *common.ReplyResponse {
	return &common.ReplyResponse{
		ID:        r.ID,
		PostID:    r.PostID,
		UserID:    r.UserID,
		Username:  r.Username,
		Content:   r.Content,
		ParentID:  r.ParentID,
		Floor:     r.Floor,
		LikeCount: r.LikeCount,
		CreatedAt: common.FormatTime(r.CreatedAt),
	}
}

func (h *Handler) toReportResponse(r *common.Report) *common.ReportResponse {
	resp := &common.ReportResponse{
		ID:         r.ID,
		PostID:     r.PostID,
		ReporterID: r.ReporterID,
		Reason:     r.Reason,
		Status:     r.Status,
		CreatedAt:  common.FormatTime(r.CreatedAt),
	}
	if !r.ReviewedAt.IsZero() {
		resp.ReviewedAt = common.FormatTime(r.ReviewedAt)
		resp.ReviewerID = r.ReviewerID
	}
	return resp
}

func (h *Handler) buildReplyTree(replies []*common.Reply, bestReplyID string) []common.ReplyResponse {
	if len(replies) == 0 {
		return []common.ReplyResponse{}
	}

	replyMap := make(map[string]*common.ReplyResponse)
	childrenMap := make(map[string][]*common.ReplyResponse)
	var topLevel []*common.Reply

	sort.Slice(replies, func(i, j int) bool {
		return replies[i].CreatedAt.Before(replies[j].CreatedAt)
	})

	for _, r := range replies {
		resp := h.toReplyResponse(r)
		resp.IsBest = r.ID == bestReplyID
		replyMap[r.ID] = resp

		if r.ParentID == "" {
			topLevel = append(topLevel, r)
		} else {
			childrenMap[r.ParentID] = append(childrenMap[r.ParentID], resp)
		}
	}

	for id, children := range childrenMap {
		if parent, exists := replyMap[id]; exists {
			parent.Children = make([]common.ReplyResponse, len(children))
			for i, c := range children {
				parent.Children[i] = *c
			}
		}
	}

	var result []common.ReplyResponse
	var bestReply *common.ReplyResponse

	for _, r := range topLevel {
		resp := replyMap[r.ID]
		if resp.IsBest {
			bestReply = resp
		} else {
			result = append(result, *resp)
		}
	}

	if bestReply != nil {
		result = append([]common.ReplyResponse{*bestReply}, result...)
	}

	return result
}

func (h *Handler) toPostHistoryResponse(ph *common.PostHistory) *common.PostHistoryResponse {
	return &common.PostHistoryResponse{
		ID:       ph.ID,
		PostID:   ph.PostID,
		Title:    ph.Title,
		Content:  ph.Content,
		Tags:     append([]string{}, ph.Tags...),
		EditedAt: common.FormatTime(ph.EditedAt),
		EditorID: ph.EditorID,
	}
}
