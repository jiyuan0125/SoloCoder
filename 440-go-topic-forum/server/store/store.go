package store

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"forum/common"
)

type Store struct {
	mu sync.RWMutex

	users           map[string]*common.User
	usernames       map[string]string

	posts           map[string]*common.Post
	postsByUser     map[string][]*common.Post
	postsByCategory map[string][]*common.Post
	postsByTag      map[string][]*common.Post
	postHistories   map[string][]*common.PostHistory

	replies         map[string]*common.Reply
	repliesByPost   map[string][]*common.Reply
	repliesByParent map[string][]*common.Reply
	repliesByUser   map[string][]*common.Reply

	likes           map[string]*common.Like
	likesByTarget   map[string][]*common.Like
	likesByUser     map[string][]*common.Like

	reports         map[string]*common.Report
	reportsByPost   map[string][]*common.Report
	reportsByStatus map[string][]*common.Report

	viewRecords     map[string]*common.ViewRecord
	viewsByPostUser map[string]bool

	tagUsage        map[string]int

	hotPosts        map[string][]*common.HotPost

	dailyPostCount  map[string]map[string]int
}

func NewStore() *Store {
	return &Store{
		users:           make(map[string]*common.User),
		usernames:       make(map[string]string),
		posts:           make(map[string]*common.Post),
		postsByUser:     make(map[string][]*common.Post),
		postsByCategory: make(map[string][]*common.Post),
		postsByTag:      make(map[string][]*common.Post),
		postHistories:   make(map[string][]*common.PostHistory),
		replies:         make(map[string]*common.Reply),
		repliesByPost:   make(map[string][]*common.Reply),
		repliesByParent: make(map[string][]*common.Reply),
		repliesByUser:   make(map[string][]*common.Reply),
		likes:           make(map[string]*common.Like),
		likesByTarget:   make(map[string][]*common.Like),
		likesByUser:     make(map[string][]*common.Like),
		reports:         make(map[string]*common.Report),
		reportsByPost:   make(map[string][]*common.Report),
		reportsByStatus: make(map[string][]*common.Report),
		viewRecords:     make(map[string]*common.ViewRecord),
		viewsByPostUser: make(map[string]bool),
		tagUsage:        make(map[string]int),
		hotPosts:        make(map[string][]*common.HotPost),
		dailyPostCount:  make(map[string]map[string]int),
	}
}

func (s *Store) CreateUser(username string, isAdmin bool) (*common.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.usernames[username]; exists {
		return nil, fmt.Errorf("username already exists")
	}

	user := &common.User{
		ID:         common.GenerateID(),
		Username:   username,
		IsAdmin:    isAdmin,
		CreatedAt:  time.Now(),
	}

	s.users[user.ID] = user
	s.usernames[username] = user.ID

	return user, nil
}

func (s *Store) GetUser(id string) (*common.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	return user, exists
}

func (s *Store) GetUserByUsername(username string) (*common.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if id, exists := s.usernames[username]; exists {
		return s.users[id], true
	}
	return nil, false
}

func (s *Store) ListUsers() []*common.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*common.User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}
	return users
}

func (s *Store) GetDailyPostCount(userID string, date time.Time) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dateKey := common.FormatDate(date)
	if userCounts, exists := s.dailyPostCount[dateKey]; exists {
		return userCounts[userID]
	}
	return 0
}

func (s *Store) IncrementDailyPostCount(userID string, date time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dateKey := common.FormatDate(date)
	if _, exists := s.dailyPostCount[dateKey]; !exists {
		s.dailyPostCount[dateKey] = make(map[string]int)
	}
	s.dailyPostCount[dateKey][userID]++
}

func (s *Store) CreatePost(req *common.CreatePostRequest, user *common.User) (*common.Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(req.Content) > common.MaxPostLength {
		return nil, fmt.Errorf("post content too long, max %d characters", common.MaxPostLength)
	}

	if common.IsBlank(req.Title) {
		return nil, fmt.Errorf("post title cannot be blank")
	}

	if common.IsBlank(req.Content) {
		return nil, fmt.Errorf("post content cannot be blank")
	}

	post := &common.Post{
		ID:         common.GenerateID(),
		UserID:     user.ID,
		Username:   user.Username,
		Title:      req.Title,
		Content:    req.Content,
		Category:   req.Category,
		Tags:       req.Tags,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	s.posts[post.ID] = post
	s.postsByUser[user.ID] = append(s.postsByUser[user.ID], post)

	if req.Category != "" {
		s.postsByCategory[req.Category] = append(s.postsByCategory[req.Category], post)
	}

	for _, tag := range req.Tags {
		if tag != "" {
			s.postsByTag[tag] = append(s.postsByTag[tag], post)
			s.tagUsage[tag]++
		}
	}

	user.PostCount++

	return post, nil
}

func (s *Store) GetPost(id string) (*common.Post, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	post, exists := s.posts[id]
	return post, exists
}

func (s *Store) UpdatePost(postID string, req *common.UpdatePostRequest, user *common.User) (*common.Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	post, exists := s.posts[postID]
	if !exists {
		return nil, fmt.Errorf("post not found")
	}

	if post.UserID != user.ID && !user.IsAdmin {
		return nil, fmt.Errorf("permission denied")
	}

	history := &common.PostHistory{
		ID:        common.GenerateID(),
		PostID:    post.ID,
		Title:     post.Title,
		Content:   post.Content,
		Tags:      append([]string{}, post.Tags...),
		EditedAt:  time.Now(),
		EditorID:  user.ID,
	}
	s.postHistories[postID] = append(s.postHistories[postID], history)

	if req.Title != "" {
		post.Title = req.Title
	}
	if req.Content != "" {
		if len(req.Content) > common.MaxPostLength {
			return nil, fmt.Errorf("post content too long, max %d characters", common.MaxPostLength)
		}
		if common.IsBlank(req.Content) {
			return nil, fmt.Errorf("post content cannot be blank")
		}
		post.Content = req.Content
	}
	if req.Category != "" {
		post.Category = req.Category
	}
	if req.Tags != nil {
		for _, oldTag := range post.Tags {
			s.tagUsage[oldTag]--
		}
		for _, newTag := range req.Tags {
			s.tagUsage[newTag]++
		}
		post.Tags = req.Tags
	}

	post.UpdatedAt = time.Now()

	return post, nil
}

func (s *Store) DeletePost(postID string, user *common.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	post, exists := s.posts[postID]
	if !exists {
		return fmt.Errorf("post not found")
	}

	if post.UserID != user.ID && !user.IsAdmin {
		return fmt.Errorf("permission denied")
	}

	for _, tag := range post.Tags {
		s.tagUsage[tag]--
		if s.tagUsage[tag] <= 0 {
			delete(s.tagUsage, tag)
		}
	}

	repliesToDelete := s.repliesByPost[postID]
	for _, reply := range repliesToDelete {
		delete(s.replies, reply.ID)
		for i, r := range s.repliesByUser[reply.UserID] {
			if r.ID == reply.ID {
				s.repliesByUser[reply.UserID] = append(s.repliesByUser[reply.UserID][:i], s.repliesByUser[reply.UserID][i+1:]...)
				break
			}
		}
		if replyAuthor, exists := s.users[reply.UserID]; exists {
			replyAuthor.ReplyCount--
		}
		if reply.ParentID != "" {
			for i, r := range s.repliesByParent[reply.ParentID] {
				if r.ID == reply.ID {
					s.repliesByParent[reply.ParentID] = append(s.repliesByParent[reply.ParentID][:i], s.repliesByParent[reply.ParentID][i+1:]...)
					break
				}
			}
		}
	}

	likesToDelete := s.likesByTarget[postID]
	for _, like := range likesToDelete {
		delete(s.likes, like.ID)
		for i, l := range s.likesByUser[like.UserID] {
			if l.ID == like.ID {
				s.likesByUser[like.UserID] = append(s.likesByUser[like.UserID][:i], s.likesByUser[like.UserID][i+1:]...)
				break
			}
		}
		if postAuthor, exists := s.users[post.UserID]; exists {
			postAuthor.LikeCount--
		}
	}

	for _, reply := range repliesToDelete {
		replyLikes := s.likesByTarget[reply.ID]
		for _, like := range replyLikes {
			delete(s.likes, like.ID)
			for i, l := range s.likesByUser[like.UserID] {
				if l.ID == like.ID {
					s.likesByUser[like.UserID] = append(s.likesByUser[like.UserID][:i], s.likesByUser[like.UserID][i+1:]...)
					break
				}
			}
			if replyAuthor, exists := s.users[reply.UserID]; exists {
				replyAuthor.LikeCount--
			}
		}
		delete(s.likesByTarget, reply.ID)
	}

	delete(s.repliesByPost, postID)
	delete(s.likesByTarget, postID)
	delete(s.posts, postID)
	delete(s.postHistories, postID)

	for i, p := range s.postsByUser[post.UserID] {
		if p.ID == postID {
			s.postsByUser[post.UserID] = append(s.postsByUser[post.UserID][:i], s.postsByUser[post.UserID][i+1:]...)
			break
		}
	}

	if post.Category != "" {
		for i, p := range s.postsByCategory[post.Category] {
			if p.ID == postID {
				s.postsByCategory[post.Category] = append(s.postsByCategory[post.Category][:i], s.postsByCategory[post.Category][i+1:]...)
				break
			}
		}
	}

	for _, tag := range post.Tags {
		if tagPosts, exists := s.postsByTag[tag]; exists {
			for i, p := range tagPosts {
				if p.ID == postID {
					s.postsByTag[tag] = append(s.postsByTag[tag][:i], s.postsByTag[tag][i+1:]...)
					break
				}
			}
		}
	}

	postAuthor := s.users[post.UserID]
	if postAuthor != nil {
		postAuthor.PostCount--
	}

	return nil
}

func (s *Store) GetPostHistory(postID string) []*common.PostHistory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.postHistories[postID]
}

func (s *Store) CreateReply(req *common.CreateReplyRequest, user *common.User) (*common.Reply, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	post, exists := s.posts[req.PostID]
	if !exists {
		return nil, fmt.Errorf("post not found")
	}

	if common.IsBlank(req.Content) {
		return nil, fmt.Errorf("reply content cannot be blank")
	}

	floor := len(s.repliesByPost[req.PostID]) + 1
	if req.ParentID != "" {
		parentReply, exists := s.replies[req.ParentID]
		if !exists {
			return nil, fmt.Errorf("parent reply not found")
		}
		if parentReply.PostID != req.PostID {
			return nil, fmt.Errorf("parent reply does not belong to this post")
		}
		floor = parentReply.Floor
	}

	reply := &common.Reply{
		ID:        common.GenerateID(),
		PostID:    req.PostID,
		UserID:    user.ID,
		Username:  user.Username,
		Content:   req.Content,
		ParentID:  req.ParentID,
		Floor:     floor,
		CreatedAt: time.Now(),
	}

	s.replies[reply.ID] = reply
	s.repliesByPost[req.PostID] = append(s.repliesByPost[req.PostID], reply)
	s.repliesByUser[user.ID] = append(s.repliesByUser[user.ID], reply)

	if req.ParentID != "" {
		s.repliesByParent[req.ParentID] = append(s.repliesByParent[req.ParentID], reply)
	}

	post.ReplyCount++
	user.ReplyCount++

	return reply, nil
}

func (s *Store) GetRepliesByPost(postID string) ([]*common.Reply, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	replies, exists := s.repliesByPost[postID]
	result := make([]*common.Reply, len(replies))
	copy(result, replies)
	return result, exists
}

func (s *Store) GetReply(id string) (*common.Reply, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reply, exists := s.replies[id]
	return reply, exists
}

func (s *Store) SetBestReply(postID, replyID string, user *common.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	post, exists := s.posts[postID]
	if !exists {
		return fmt.Errorf("post not found")
	}

	if post.UserID != user.ID {
		return fmt.Errorf("only the post author can set best reply")
	}

	reply, exists := s.replies[replyID]
	if !exists {
		return fmt.Errorf("reply not found")
	}

	if reply.PostID != postID {
		return fmt.Errorf("reply does not belong to this post")
	}

	post.BestReplyID = replyID
	return nil
}

func (s *Store) Like(req *common.LikeRequest, user *common.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.TargetType != common.TargetTypePost && req.TargetType != common.TargetTypeReply {
		return fmt.Errorf("invalid target type")
	}

	if req.TargetType == common.TargetTypePost {
		post, exists := s.posts[req.TargetID]
		if !exists {
			return fmt.Errorf("post not found")
		}
		if post.UserID == user.ID {
			return fmt.Errorf("cannot like your own post")
		}
	} else {
		reply, exists := s.replies[req.TargetID]
		if !exists {
			return fmt.Errorf("reply not found")
		}
		if reply.UserID == user.ID {
			return fmt.Errorf("cannot like your own reply")
		}
	}

	for _, like := range s.likesByUser[user.ID] {
		if like.TargetID == req.TargetID && like.TargetType == req.TargetType {
			return fmt.Errorf("already liked")
		}
	}

	like := &common.Like{
		ID:         common.GenerateID(),
		UserID:     user.ID,
		TargetID:   req.TargetID,
		TargetType: req.TargetType,
		CreatedAt:  time.Now(),
	}

	s.likes[like.ID] = like
	s.likesByTarget[req.TargetID] = append(s.likesByTarget[req.TargetID], like)
	s.likesByUser[user.ID] = append(s.likesByUser[user.ID], like)

	if req.TargetType == common.TargetTypePost {
		s.posts[req.TargetID].LikeCount++
		if targetUser, exists := s.users[s.posts[req.TargetID].UserID]; exists {
			targetUser.LikeCount++
		}
	} else {
		s.replies[req.TargetID].LikeCount++
		if targetUser, exists := s.users[s.replies[req.TargetID].UserID]; exists {
			targetUser.LikeCount++
		}
	}

	return nil
}

func (s *Store) Unlike(req *common.LikeRequest, user *common.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var likeToDelete *common.Like
	var likeIndex int

	for i, like := range s.likesByUser[user.ID] {
		if like.TargetID == req.TargetID && like.TargetType == req.TargetType {
			likeToDelete = like
			likeIndex = i
			break
		}
	}

	if likeToDelete == nil {
		return fmt.Errorf("like not found")
	}

	delete(s.likes, likeToDelete.ID)
	s.likesByUser[user.ID] = append(s.likesByUser[user.ID][:likeIndex], s.likesByUser[user.ID][likeIndex+1:]...)

	for i, like := range s.likesByTarget[req.TargetID] {
		if like.ID == likeToDelete.ID {
			s.likesByTarget[req.TargetID] = append(s.likesByTarget[req.TargetID][:i], s.likesByTarget[req.TargetID][i+1:]...)
			break
		}
	}

	if req.TargetType == common.TargetTypePost {
		s.posts[req.TargetID].LikeCount--
		if targetUser, exists := s.users[s.posts[req.TargetID].UserID]; exists {
			targetUser.LikeCount--
		}
	} else {
		s.replies[req.TargetID].LikeCount--
		if targetUser, exists := s.users[s.replies[req.TargetID].UserID]; exists {
			targetUser.LikeCount--
		}
	}

	return nil
}

func (s *Store) SetTop(adminID, postID string, isTop bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	admin, exists := s.users[adminID]
	if !exists || !admin.IsAdmin {
		return fmt.Errorf("permission denied")
	}

	post, exists := s.posts[postID]
	if !exists {
		return fmt.Errorf("post not found")
	}

	post.IsTop = isTop
	return nil
}

func (s *Store) SetEssence(adminID, postID string, isEssence bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	admin, exists := s.users[adminID]
	if !exists || !admin.IsAdmin {
		return fmt.Errorf("permission denied")
	}

	post, exists := s.posts[postID]
	if !exists {
		return fmt.Errorf("post not found")
	}

	post.IsEssence = isEssence
	return nil
}

func (s *Store) CreateReport(req *common.ReportRequest) (*common.Report, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.posts[req.PostID]; !exists {
		return nil, fmt.Errorf("post not found")
	}

	if common.IsBlank(req.Reason) {
		return nil, fmt.Errorf("report reason cannot be blank")
	}

	report := &common.Report{
		ID:         common.GenerateID(),
		PostID:     req.PostID,
		ReporterID: req.ReporterID,
		Reason:     req.Reason,
		Status:     common.ReportStatusPending,
		CreatedAt:  time.Now(),
	}

	s.reports[report.ID] = report
	s.reportsByPost[req.PostID] = append(s.reportsByPost[req.PostID], report)
	s.reportsByStatus[common.ReportStatusPending] = append(s.reportsByStatus[common.ReportStatusPending], report)

	return report, nil
}

func (s *Store) ReviewReport(req *common.ReviewReportRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	admin, exists := s.users[req.AdminID]
	if !exists || !admin.IsAdmin {
		return fmt.Errorf("permission denied")
	}

	report, exists := s.reports[req.ReportID]
	if !exists {
		return fmt.Errorf("report not found")
	}

	if req.Status != common.ReportStatusResolved && req.Status != common.ReportStatusRejected {
		return fmt.Errorf("invalid status")
	}

	oldStatus := report.Status
	for i, r := range s.reportsByStatus[oldStatus] {
		if r.ID == report.ID {
			s.reportsByStatus[oldStatus] = append(s.reportsByStatus[oldStatus][:i], s.reportsByStatus[oldStatus][i+1:]...)
			break
		}
	}

	report.Status = req.Status
	report.ReviewedAt = time.Now()
	report.ReviewerID = req.AdminID

	s.reportsByStatus[req.Status] = append(s.reportsByStatus[req.Status], report)

	return nil
}

func (s *Store) GetReportsByStatus(status string) []*common.Report {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.reportsByStatus[status]
}

func (s *Store) RecordView(postID, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := postID + ":" + userID
	if s.viewsByPostUser[key] {
		return
	}

	post, exists := s.posts[postID]
	if !exists {
		return
	}

	viewRecord := &common.ViewRecord{
		PostID:   postID,
		UserID:   userID,
		ViewedAt: time.Now(),
	}

	s.viewRecords[common.GenerateID()] = viewRecord
	s.viewsByPostUser[key] = true
	post.ViewCount++
}

func (s *Store) SearchPosts(req *common.SearchRequest) []*common.Post {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*common.Post

	if req.Keyword != "" {
		keyword := strings.ToLower(req.Keyword)
		for _, post := range s.posts {
			if strings.Contains(strings.ToLower(post.Title), keyword) ||
				strings.Contains(strings.ToLower(post.Content), keyword) ||
				common.ContainsString(post.Tags, req.Keyword) {
				result = append(result, post)
			}
		}
	} else if req.Tag != "" {
		result = s.postsByTag[req.Tag]
	} else if req.Category != "" {
		result = s.postsByCategory[req.Category]
	} else if req.UserID != "" {
		result = s.postsByUser[req.UserID]
	} else {
		for _, post := range s.posts {
			result = append(result, post)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].IsTop != result[j].IsTop {
			return result[i].IsTop
		}
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	offset := req.Offset
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	if offset >= len(result) {
		return []*common.Post{}
	}

	end := offset + limit
	if end > len(result) {
		end = len(result)
	}

	return result[offset:end]
}

func (s *Store) GetLeaderboard() []*common.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*common.User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].LikeCount > users[j].LikeCount
	})

	return users
}

func (s *Store) GetTagCloud(limit int) []common.TagUsage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tags := make([]common.TagUsage, 0, len(s.tagUsage))
	for name, count := range s.tagUsage {
		if count > 0 {
			tags = append(tags, common.TagUsage{Name: name, PostCount: count})
		}
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].PostCount > tags[j].PostCount
	})

	if limit > 0 && len(tags) > limit {
		tags = tags[:limit]
	}

	return tags
}

func (s *Store) CalculateHotPosts(date time.Time) []*common.HotPost {
	s.mu.Lock()
	defer s.mu.Unlock()

	dateKey := common.FormatDate(date)

	var posts []*common.Post
	for _, post := range s.posts {
		posts = append(posts, post)
	}

	hotPosts := make([]*common.HotPost, 0, len(posts))
	for _, post := range posts {
		score := float64(post.ViewCount)*common.ViewWeight + float64(post.ReplyCount)*common.ReplyWeight
		hotPosts = append(hotPosts, &common.HotPost{
			PostID: post.ID,
			Score:  score,
			Date:   dateKey,
		})
	}

	sort.Slice(hotPosts, func(i, j int) bool {
		return hotPosts[i].Score > hotPosts[j].Score
	})

	if len(hotPosts) > common.HotPostCount {
		hotPosts = hotPosts[:common.HotPostCount]
	}

	s.hotPosts[dateKey] = hotPosts

	return hotPosts
}

func (s *Store) GetHotPosts(date time.Time) []*common.HotPost {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dateKey := common.FormatDate(date)
	return s.hotPosts[dateKey]
}

func (s *Store) ListAllPosts() []*common.Post {
	s.mu.RLock()
	defer s.mu.RUnlock()

	posts := make([]*common.Post, 0, len(s.posts))
	for _, p := range s.posts {
		posts = append(posts, p)
	}

	sort.Slice(posts, func(i, j int) bool {
		if posts[i].IsTop != posts[j].IsTop {
			return posts[i].IsTop
		}
		return posts[i].CreatedAt.After(posts[j].CreatedAt)
	})

	return posts
}
