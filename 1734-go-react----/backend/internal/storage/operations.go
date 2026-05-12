package storage

import (
	"school-connect/internal/models"
	"sort"
	"strings"
	"time"
)

func (s *MemoryStore) CreateUser(user *models.User) *models.User {
	s.mu.Lock()
	defer s.mu.Unlock()
	user.ID = generateID()
	user.CreatedAt = time.Now()
	s.Users[user.ID] = user
	return user
}

func (s *MemoryStore) GetUserByID(id string) *models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Users[id]
}

func (s *MemoryStore) GetUserByUsername(username string) *models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.Users {
		if u.Username == username {
			return u
		}
	}
	return nil
}

func (s *MemoryStore) CreateStudent(student *models.Student) *models.Student {
	s.mu.Lock()
	defer s.mu.Unlock()
	student.ID = generateID()
	s.Students[student.ID] = student
	return student
}

func (s *MemoryStore) GetStudent(id string) *models.Student {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Students[id]
}

func (s *MemoryStore) GetStudentsByClass(class string) []*models.Student {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*models.Student
	for _, st := range s.Students {
		if st.Class == class {
			result = append(result, st)
		}
	}
	return result
}

func (s *MemoryStore) GetStudentsByGrade(grade string) []*models.Student {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*models.Student
	for _, st := range s.Students {
		if st.Grade == grade {
			result = append(result, st)
		}
	}
	return result
}

func (s *MemoryStore) CreateParent(parent *models.Parent) *models.Parent {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Parents[parent.UserID] = parent
	return parent
}

func (s *MemoryStore) GetParent(userID string) *models.Parent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Parents[userID]
}

func (s *MemoryStore) GetParentsByClass(class string) []*models.Parent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	studentIDs := make(map[string]bool)
	for _, st := range s.Students {
		if st.Class == class {
			studentIDs[st.ID] = true
		}
	}
	var result []*models.Parent
	for _, p := range s.Parents {
		for _, cid := range p.ChildIDs {
			if studentIDs[cid] {
				result = append(result, p)
				break
			}
		}
	}
	return result
}

func (s *MemoryStore) GetParentsByGrade(grade string) []*models.Parent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	studentIDs := make(map[string]bool)
	for _, st := range s.Students {
		if st.Grade == grade {
			studentIDs[st.ID] = true
		}
	}
	var result []*models.Parent
	for _, p := range s.Parents {
		for _, cid := range p.ChildIDs {
			if studentIDs[cid] {
				result = append(result, p)
				break
			}
		}
	}
	return result
}

func (s *MemoryStore) GetAllParents() []*models.Parent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*models.Parent
	for _, p := range s.Parents {
		result = append(result, p)
	}
	return result
}

func (s *MemoryStore) CreateTeacher(teacher *models.Teacher) *models.Teacher {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Teachers[teacher.UserID] = teacher
	return teacher
}

func (s *MemoryStore) GetTeacher(userID string) *models.Teacher {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Teachers[userID]
}

func (s *MemoryStore) GetAllTeachers() []*models.Teacher {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*models.Teacher
	for _, t := range s.Teachers {
		result = append(result, t)
	}
	return result
}

func (s *MemoryStore) CreateAnnouncement(ann *models.Announcement) *models.Announcement {
	s.mu.Lock()
	defer s.mu.Unlock()
	ann.ID = generateID()
	ann.PublishedAt = time.Now()
	s.Announcements[ann.ID] = ann
	return ann
}

func (s *MemoryStore) GetAnnouncement(id string) *models.Announcement {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Announcements[id]
}

func (s *MemoryStore) GetAnnouncementsForParent(parentID string) []*models.Announcement {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	parent := s.Parents[parentID]
	if parent == nil {
		return nil
	}
	
	parentGrades := make(map[string]bool)
	parentClasses := make(map[string]bool)
	for _, childID := range parent.ChildIDs {
		if student, ok := s.Students[childID]; ok {
			parentGrades[student.Grade] = true
			parentClasses[student.Class] = true
		}
	}
	
	var result []*models.Announcement
	for _, ann := range s.Announcements {
		include := false
		switch ann.ScopeType {
		case models.ScopeAll:
			include = true
		case models.ScopeGrade:
			include = parentGrades[ann.ScopeValue]
		case models.ScopeClass:
			include = parentClasses[ann.ScopeValue]
		}
		if include {
			result = append(result, ann)
		}
	}
	
	sort.Slice(result, func(i, j int) bool {
		urgencyOrder := map[models.Urgency]int{
			models.UrgencyUrgent:    3,
			models.UrgencyImportant: 2,
			models.UrgencyNormal:    1,
		}
		if urgencyOrder[result[i].Urgency] != urgencyOrder[result[j].Urgency] {
			return urgencyOrder[result[i].Urgency] > urgencyOrder[result[j].Urgency]
		}
		return result[i].PublishedAt.After(result[j].PublishedAt)
	})
	
	return result
}

func (s *MemoryStore) GetAllAnnouncements() []*models.Announcement {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*models.Announcement
	for _, ann := range s.Announcements {
		result = append(result, ann)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].PublishedAt.After(result[j].PublishedAt)
	})
	return result
}

func (s *MemoryStore) MarkAnnouncementRead(announcementID, parentID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	reads := s.AnnouncementReads[announcementID]
	for _, r := range reads {
		if r.ParentID == parentID {
			return false
		}
	}
	
	reads = append(reads, &models.AnnouncementRead{
		AnnouncementID: announcementID,
		ParentID:       parentID,
		ReadAt:         time.Now(),
	})
	s.AnnouncementReads[announcementID] = reads
	return true
}

func (s *MemoryStore) IsAnnouncementRead(announcementID, parentID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	reads := s.AnnouncementReads[announcementID]
	for _, r := range reads {
		if r.ParentID == parentID {
			return true
		}
	}
	return false
}

func (s *MemoryStore) GetAnnouncementStats(announcementID string) *models.AnnouncementStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	ann := s.Announcements[announcementID]
	if ann == nil {
		return nil
	}
	
	var targetParents []*models.Parent
	switch ann.ScopeType {
	case models.ScopeAll:
		targetParents = s.GetAllParents()
	case models.ScopeGrade:
		targetParents = s.GetParentsByGrade(ann.ScopeValue)
	case models.ScopeClass:
		targetParents = s.GetParentsByClass(ann.ScopeValue)
	}
	
	parentIDs := make(map[string]bool)
	for _, p := range targetParents {
		parentIDs[p.UserID] = true
	}
	
	total := len(parentIDs)
	readCount := 0
	reads := s.AnnouncementReads[announcementID]
	for _, r := range reads {
		if parentIDs[r.ParentID] {
			readCount++
		}
	}
	
	return &models.AnnouncementStats{
		AnnouncementID: announcementID,
		TotalParents:   total,
		ReadCount:      readCount,
		UnreadCount:    total - readCount,
	}
}

func (s *MemoryStore) RecordScore(score *models.ExamScore) *models.ExamScore {
	s.mu.Lock()
	defer s.mu.Unlock()
	score.ID = generateID()
	score.RecordedAt = time.Now()
	s.Scores[score.ID] = score
	return score
}

func (s *MemoryStore) GetStudentLatestScores(studentID, semester string) []*models.ExamScore {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	latestBySubject := make(map[string]*models.ExamScore)
	for _, sc := range s.Scores {
		if sc.StudentID == studentID {
			if semester == "" || sc.Semester == semester {
				existing, ok := latestBySubject[sc.Subject]
				if !ok || sc.RecordedAt.After(existing.RecordedAt) {
					latestBySubject[sc.Subject] = sc
				}
			}
		}
	}
	
	var result []*models.ExamScore
	for _, sc := range latestBySubject {
		result = append(result, sc)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Subject < result[j].Subject
	})
	return result
}

func (s *MemoryStore) GetStudentSubjectHistory(studentID, subject string) []*models.ExamScore {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	latestBySemester := make(map[string]*models.ExamScore)
	for _, sc := range s.Scores {
		if sc.StudentID == studentID && sc.Subject == subject {
			existing, ok := latestBySemester[sc.Semester]
			if !ok || sc.RecordedAt.After(existing.RecordedAt) {
				latestBySemester[sc.Semester] = sc
			}
		}
	}
	
	var result []*models.ExamScore
	for _, sc := range latestBySemester {
		result = append(result, sc)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ExamDate.Before(result[j].ExamDate)
	})
	return result
}

func (s *MemoryStore) GetClassScores(class, subject, semester string) []*models.ExamScore {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	latestByStudent := make(map[string]*models.ExamScore)
	for _, sc := range s.Scores {
		if sc.Class == class && sc.Subject == subject && sc.Semester == semester {
			existing, ok := latestByStudent[sc.StudentID]
			if !ok || sc.RecordedAt.After(existing.RecordedAt) {
				latestByStudent[sc.StudentID] = sc
			}
		}
	}
	
	var result []*models.ExamScore
	for _, sc := range latestByStudent {
		result = append(result, sc)
	}
	return result
}

func (s *MemoryStore) SendMessage(msg *models.Message) *models.Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	msg.ID = generateID()
	msg.Timestamp = time.Now()
	msg.IsRead = false
	msg.Content = truncateContent(msg.Content)
	s.Messages[msg.ID] = msg
	
	convID := getConversationID(msg.SenderID, msg.ReceiverID)
	conv, ok := s.Conversations[convID]
	if !ok {
		conv = &models.Conversation{
			ID:             convID,
			Participant1ID: msg.SenderID,
			Participant2ID: msg.ReceiverID,
		}
	}
	conv.LastMessage = msg.Content
	conv.LastMessageTime = msg.Timestamp
	s.Conversations[convID] = conv
	
	return msg
}

func (s *MemoryStore) GetMessages(conversationID, userID string, markAsRead bool) []*models.Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	conv := s.Conversations[conversationID]
	if conv == nil {
		return nil
	}
	
	var result []*models.Message
	for _, msg := range s.Messages {
		if (msg.SenderID == conv.Participant1ID && msg.ReceiverID == conv.Participant2ID) ||
		   (msg.SenderID == conv.Participant2ID && msg.ReceiverID == conv.Participant1ID) {
			result = append(result, msg)
			if markAsRead && msg.ReceiverID == userID && !msg.IsRead {
				now := time.Now()
				msg.IsRead = true
				msg.ReadAt = &now
			}
		}
	}
	
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})
	return result
}

func (s *MemoryStore) GetConversations(userID string) []*models.ConversationListItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var result []*models.ConversationListItem
	for _, conv := range s.Conversations {
		if conv.Participant1ID == userID || conv.Participant2ID == userID {
			otherID := conv.Participant1ID
			if otherID == userID {
				otherID = conv.Participant2ID
			}
			
			unreadCount := 0
			for _, msg := range s.Messages {
				if msg.ReceiverID == userID &&
				   ((msg.SenderID == conv.Participant1ID && msg.ReceiverID == conv.Participant2ID) ||
				    (msg.SenderID == conv.Participant2ID && msg.ReceiverID == conv.Participant1ID)) &&
				   !msg.IsRead {
					unreadCount++
				}
			}
			
			otherUser := s.Users[otherID]
			otherName := ""
			if otherUser != nil {
				otherName = otherUser.Name
			}
			
			result = append(result, &models.ConversationListItem{
				ConversationID:  conv.ID,
				OtherUserID:     otherID,
				OtherUserName:   otherName,
				LastMessage:     conv.LastMessage,
				LastMessageTime: conv.LastMessageTime,
				UnreadCount:     unreadCount,
			})
		}
	}
	
	sort.Slice(result, func(i, j int) bool {
		return result[i].LastMessageTime.After(result[j].LastMessageTime)
	})
	return result
}

func (s *MemoryStore) GetTotalUnreadCount(userID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	count := 0
	for _, msg := range s.Messages {
		if msg.ReceiverID == userID && !msg.IsRead {
			count++
		}
	}
	return count
}

func (s *MemoryStore) MarkConversationRead(conversationID, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	conv := s.Conversations[conversationID]
	if conv == nil {
		return
	}
	
	now := time.Now()
	for _, msg := range s.Messages {
		if msg.ReceiverID == userID &&
		   ((msg.SenderID == conv.Participant1ID && msg.ReceiverID == conv.Participant2ID) ||
		    (msg.SenderID == conv.Participant2ID && msg.ReceiverID == conv.Participant1ID)) &&
		   !msg.IsRead {
			msg.IsRead = true
			msg.ReadAt = &now
		}
	}
}

func truncateContent(content string) string {
	if len([]rune(content)) > 500 {
		return string([]rune(content)[:500])
	}
	return content
}

func getConversationID(user1, user2 string) string {
	if strings.Compare(user1, user2) < 0 {
		return user1 + "_" + user2
	}
	return user2 + "_" + user1
}

func (s *MemoryStore) SeedDemoData() {
	admin := s.CreateUser(&models.User{
		Username: "admin",
		Password: "admin123",
		Role:     models.RoleAdmin,
		Name:     "管理员",
	})
	
	teacher1 := s.CreateUser(&models.User{
		Username: "teacher1",
		Password: "123456",
		Role:     models.RoleTeacher,
		Name:     "张老师",
	})
	s.CreateTeacher(&models.Teacher{
		UserID:   teacher1.ID,
		Subject:  "数学",
		ClassIDs: []string{"一年级1班", "一年级2班"},
	})
	
	teacher2 := s.CreateUser(&models.User{
		Username: "teacher2",
		Password: "123456",
		Role:     models.RoleTeacher,
		Name:     "李老师",
	})
	s.CreateTeacher(&models.Teacher{
		UserID:   teacher2.ID,
		Subject:  "语文",
		ClassIDs: []string{"一年级1班"},
	})
	
	student1 := s.CreateStudent(&models.Student{
		Name:      "小明",
		Grade:     "一年级",
		Class:     "一年级1班",
		StudentNo: "2024001",
	})
	student2 := s.CreateStudent(&models.Student{
		Name:      "小红",
		Grade:     "一年级",
		Class:     "一年级1班",
		StudentNo: "2024002",
	})
	student3 := s.CreateStudent(&models.Student{
		Name:      "小刚",
		Grade:     "一年级",
		Class:     "一年级2班",
		StudentNo: "2024003",
	})
	
	parent1 := s.CreateUser(&models.User{
		Username: "parent1",
		Password: "123456",
		Role:     models.RoleParent,
		Name:     "小明爸爸",
	})
	s.CreateParent(&models.Parent{
		UserID:   parent1.ID,
		ChildIDs: []string{student1.ID},
	})
	
	parent2 := s.CreateUser(&models.User{
		Username: "parent2",
		Password: "123456",
		Role:     models.RoleParent,
		Name:     "小红妈妈",
	})
	s.CreateParent(&models.Parent{
		UserID:   parent2.ID,
		ChildIDs: []string{student2.ID},
	})
	
	s.CreateAnnouncement(&models.Announcement{
		Title:       "开学通知",
		Content:     "新学期将于9月1日正式开学，请各位家长提前做好准备。",
		ScopeType:   models.ScopeAll,
		ScopeValue:  "",
		Urgency:     models.UrgencyImportant,
		PublishedBy: admin.ID,
	})
	
	s.CreateAnnouncement(&models.Announcement{
		Title:       "紧急：台风预警",
		Content:     "受台风影响，明天停课一天，请各位家长注意孩子安全。",
		ScopeType:   models.ScopeAll,
		ScopeValue:  "",
		Urgency:     models.UrgencyUrgent,
		PublishedBy: admin.ID,
	})
	
	s.CreateAnnouncement(&models.Announcement{
		Title:       "一年级家长会",
		Content:     "一年级家长会定于周五下午3点在各班教室召开。",
		ScopeType:   models.ScopeGrade,
		ScopeValue:  "一年级",
		Urgency:     models.UrgencyNormal,
		PublishedBy: teacher1.ID,
	})
}
