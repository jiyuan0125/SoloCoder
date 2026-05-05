package main

import (
	"sync"
	"time"

	"tag-manager/common"
)

type Store struct {
	tags         map[string]*common.Tag
	tagNames     map[string]string
	tagGroups    map[string]*common.TagGroup
	userTags     map[string]map[string]*common.UserTag
	auditLogs    []*common.AuditLog

	tagMutex     sync.RWMutex
	groupMutex     sync.RWMutex
	userTagMutex   sync.RWMutex
	auditMutex     sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		tags:       make(map[string]*common.Tag),
		tagNames:   make(map[string]string),
		tagGroups:  make(map[string]*common.TagGroup),
		userTags:   make(map[string]map[string]*common.UserTag),
		auditLogs:  []*common.AuditLog{},
	}
}

func (s *Store) GetTagByID(id string) *common.Tag {
	s.tagMutex.RLock()
	defer s.tagMutex.RUnlock()
	return s.tags[id]
}

func (s *Store) GetTagByName(name string) *common.Tag {
	s.tagMutex.RLock()
	defer s.tagMutex.RUnlock()
	if id, exists := s.tagNames[name]; exists {
		return s.tags[id]
	}
	return nil
}

func (s *Store) CreateTag(tag *common.Tag) error {
	s.tagMutex.Lock()
	defer s.tagMutex.Unlock()

	if _, exists := s.tagNames[tag.Name]; exists {
		return &DuplicateNameError{Name: tag.Name}
	}

	s.tags[tag.ID] = tag
	s.tagNames[tag.Name] = tag.ID
	return nil
}

func (s *Store) UpdateTag(tag *common.Tag, oldName string) error {
	s.tagMutex.Lock()
	defer s.tagMutex.Unlock()

	if oldName != tag.Name {
		if _, exists := s.tagNames[tag.Name]; exists {
			return &DuplicateNameError{Name: tag.Name}
		}
		delete(s.tagNames, oldName)
		s.tagNames[tag.Name] = tag.ID
	}

	s.tags[tag.ID] = tag
	return nil
}

func (s *Store) DeleteTag(id string) *common.Tag {
	s.tagMutex.Lock()
	defer s.tagMutex.Unlock()

	tag := s.tags[id]
	if tag == nil {
		return nil
	}

	delete(s.tagNames, tag.Name)
	delete(s.tags, id)
	return tag
}

func (s *Store) ListTags() []*common.Tag {
	s.tagMutex.RLock()
	defer s.tagMutex.RUnlock()

	tags := make([]*common.Tag, 0, len(s.tags))
	for _, tag := range s.tags {
		tags = append(tags, tag)
	}
	return tags
}

func (s *Store) ListTagsByGroup(groupID string) []*common.Tag {
	s.tagMutex.RLock()
	defer s.tagMutex.RUnlock()

	var tags []*common.Tag
	for _, tag := range s.tags {
		if tag.GroupID == groupID {
			tags = append(tags, tag)
		}
	}
	return tags
}

func (s *Store) GetTagGroupByID(id string) *common.TagGroup {
	s.groupMutex.RLock()
	defer s.groupMutex.RUnlock()
	return s.tagGroups[id]
}

func (s *Store) CreateTagGroup(group *common.TagGroup) {
	s.groupMutex.Lock()
	defer s.groupMutex.Unlock()
	s.tagGroups[group.ID] = group
}

func (s *Store) ListTagGroups() []*common.TagGroup {
	s.groupMutex.RLock()
	defer s.groupMutex.RUnlock()

	groups := make([]*common.TagGroup, 0, len(s.tagGroups))
	for _, group := range s.tagGroups {
		groups = append(groups, group)
	}
	return groups
}

func (s *Store) GetUserTag(userID, tagID string) *common.UserTag {
	s.userTagMutex.RLock()
	defer s.userTagMutex.RUnlock()

	userTags := s.userTags[userID]
	if userTags == nil {
		return nil
	}
	return userTags[tagID]
}

func (s *Store) ApplyTag(userID, tagID string, value interface{}) {
	s.userTagMutex.Lock()
	defer s.userTagMutex.Unlock()

	if s.userTags[userID] == nil {
		s.userTags[userID] = make(map[string]*common.UserTag)
	}

	s.userTags[userID][tagID] = &common.UserTag{
		UserID:    userID,
		TagID:     tagID,
		TagValue:  value,
		CreatedAt: time.Now().Unix(),
	}
}

func (s *Store) RemoveUserTag(userID, tagID string) {
	s.userTagMutex.Lock()
	defer s.userTagMutex.Unlock()

	if userTags := s.userTags[userID]; userTags != nil {
		delete(userTags, tagID)
	}
}

func (s *Store) DeleteAllTagRelations(tagID string) int64 {
	s.userTagMutex.Lock()
	defer s.userTagMutex.Unlock()

	var count int64
	for userID, userTags := range s.userTags {
		if userTags[tagID] != nil {
			delete(userTags, tagID)
			count++
			if len(userTags) == 0 {
				delete(s.userTags, userID)
			}
		}
	}
	return count
}

func (s *Store) GetTagImpactCount(tagID string) int64 {
	s.userTagMutex.RLock()
	defer s.userTagMutex.RUnlock()

	var count int64
	for _, userTags := range s.userTags {
		if userTags[tagID] != nil {
			count++
		}
	}
	return count
}

func (s *Store) ListUsersWithTag(tagID string) []string {
	s.userTagMutex.RLock()
	defer s.userTagMutex.RUnlock()

	var userIDs []string
	for userID, userTags := range s.userTags {
		if userTags[tagID] != nil {
			userIDs = append(userIDs, userID)
		}
	}
	return userIDs
}

func (s *Store) GetUserTags(userID string) []*common.UserTag {
	s.userTagMutex.RLock()
	defer s.userTagMutex.RUnlock()

	userTags := s.userTags[userID]
	if userTags == nil {
		return nil
	}

	tags := make([]*common.UserTag, 0, len(userTags))
	for _, ut := range userTags {
		tags = append(tags, ut)
	}
	return tags
}

func (s *Store) GetTagUsageStats() []common.TagUsageStats {
	s.userTagMutex.RLock()
	defer s.userTagMutex.RUnlock()

	tagCounts := make(map[string]int64)

	for _, userTags := range s.userTags {
		for tagID := range userTags {
			tagCounts[tagID]++
		}
	}

	s.tagMutex.RLock()
	defer s.tagMutex.RUnlock()

	var stats []common.TagUsageStats
	for tagID, count := range tagCounts {
		tag := s.tags[tagID]
		if tag != nil {
			stats = append(stats, common.TagUsageStats{
				TagID:    tagID,
				TagName:  tag.Name,
				UserCount: count,
			})
		}
	}
	return stats
}

func (s *Store) GetTagTrend(days int) []common.TagTrend {
	s.userTagMutex.RLock()
	defer s.userTagMutex.RUnlock()

	now := time.Now().Unix()
	threshold := now - int64(days*24*60*60)

	tagCounts := make(map[string]int64)

	for _, userTags := range s.userTags {
		for _, ut := range userTags {
			if ut.CreatedAt >= threshold {
				tagCounts[ut.TagID]++
			}
		}
	}

	s.tagMutex.RLock()
	defer s.tagMutex.RUnlock()

	var trends []common.TagTrend
	for tagID, count := range tagCounts {
		tag := s.tags[tagID]
		if tag != nil {
			trends = append(trends, common.TagTrend{
				TagID:        tagID,
				TagName:      tag.Name,
				NewUserCount: count,
			})
		}
	}
	return trends
}

func (s *Store) AddAuditLog(log *common.AuditLog) {
	s.auditMutex.Lock()
	defer s.auditMutex.Unlock()
	s.auditLogs = append(s.auditLogs, log)
}

func (s *Store) ListAuditLogs() []*common.AuditLog {
	s.auditMutex.RLock()
	defer s.auditMutex.RUnlock()

	logs := make([]*common.AuditLog, len(s.auditLogs))
	copy(logs, s.auditLogs)
	return logs
}

func (s *Store) GetAllUserTags() map[string]map[string]*common.UserTag {
	s.userTagMutex.RLock()
	defer s.userTagMutex.RUnlock()

	result := make(map[string]map[string]*common.UserTag)
	for userID, tags := range s.userTags {
		result[userID] = make(map[string]*common.UserTag)
		for tagID, ut := range tags {
			result[userID][tagID] = ut
		}
	}
	return result
}
