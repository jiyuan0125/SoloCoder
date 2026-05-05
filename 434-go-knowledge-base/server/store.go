package main

import (
	"kb/common"
	"strings"
	"sync"
	"time"
)

type Store struct {
	mu            sync.RWMutex
	articles      map[string]*common.Article
	versions      map[string][]*common.ArticleVersion
	favorites     map[string][]*common.Favorite
	references    map[string][]*common.Reference
	articleByUser map[string][]string
	articleByTag  map[string][]string
	articleByCat  map[string][]string
	nextID        int64
}

func NewStore() *Store {
	return &Store{
		articles:      make(map[string]*common.Article),
		versions:      make(map[string][]*common.ArticleVersion),
		favorites:     make(map[string][]*common.Favorite),
		references:    make(map[string][]*common.Reference),
		articleByUser: make(map[string][]string),
		articleByTag:  make(map[string][]string),
		articleByCat:  make(map[string][]string),
		nextID:        1,
	}
}

func (s *Store) generateIDLocked() string {
	id := s.nextID
	s.nextID++
	return "id_" + time.Now().Format("20060102150405") + "_" + string(rune('0'+id%10))
}

func (s *Store) GenerateID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.generateIDLocked()
}

func (s *Store) CreateArticle(req *common.CreateArticleRequest) (*common.Article, *common.ArticleVersion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	articleID := s.generateIDLocked()

	article := &common.Article{
		ID:                 articleID,
		Title:              req.Title,
		Content:            req.Content,
		Category:           req.Category,
		Tags:               req.Tags,
		Status:             common.StatusDraft,
		AuthorID:           req.AuthorID,
		CreatedAt:          now,
		UpdatedAt:          now,
		CurrentVersion:     1,
		AccessLevel:        req.AccessLevel,
		DepartmentIDs:      req.DepartmentIDs,
		ViewCount:          0,
		ReferenceCount:     0,
		RelatedArticleIDs:  []string{},
	}

	version := &common.ArticleVersion{
		ID:         s.generateIDLocked(),
		ArticleID:  articleID,
		VersionNum: 1,
		Title:      req.Title,
		Content:    req.Content,
		Category:   req.Category,
		Tags:       req.Tags,
		CreatedBy:  req.AuthorID,
		CreatedAt:  now,
	}

	s.articles[articleID] = article
	s.versions[articleID] = []*common.ArticleVersion{version}
	s.articleByUser[req.AuthorID] = append(s.articleByUser[req.AuthorID], articleID)

	for _, tag := range req.Tags {
		s.articleByTag[tag] = append(s.articleByTag[tag], articleID)
	}

	if req.Category != "" {
		s.articleByCat[req.Category] = append(s.articleByCat[req.Category], articleID)
	}

	return article, version, nil
}

func (s *Store) GetArticle(articleID string) (*common.Article, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	article, exists := s.articles[articleID]
	return article, exists
}

func (s *Store) UpdateArticle(articleID string, req *common.UpdateArticleRequest) (*common.Article, *common.ArticleVersion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	article, exists := s.articles[articleID]
	if !exists {
		return nil, nil, nil
	}

	if article.Status == common.StatusArchived {
		return nil, nil, nil
	}

	if article.AuthorID != req.AuthorID {
		return nil, nil, nil
	}

	wasPublished := (article.Status == common.StatusPublished)

	now := time.Now()
	newVersionNum := article.CurrentVersion + 1

	version := &common.ArticleVersion{
		ID:         s.generateIDLocked(),
		ArticleID:  articleID,
		VersionNum: newVersionNum,
		Title:      req.Title,
		Content:    req.Content,
		Category:   req.Category,
		Tags:       req.Tags,
		CreatedBy:  req.AuthorID,
		CreatedAt:  now,
	}

	article.Title = req.Title
	article.Content = req.Content
	article.Category = req.Category
	article.Tags = req.Tags
	article.UpdatedAt = now
	article.CurrentVersion = newVersionNum
	article.AccessLevel = req.AccessLevel
	article.DepartmentIDs = req.DepartmentIDs
	if req.RelatedArticleIDs != nil {
		article.RelatedArticleIDs = req.RelatedArticleIDs
	}

	if wasPublished {
		article.Status = common.StatusDraft
	}

	s.versions[articleID] = append(s.versions[articleID], version)

	return article, version, nil
}

func (s *Store) PublishArticle(articleID string, authorID string) (*common.Article, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	article, exists := s.articles[articleID]
	if !exists {
		return nil, nil
	}

	if article.Status == common.StatusArchived {
		return nil, nil
	}

	if article.AuthorID != authorID {
		return nil, nil
	}

	if article.Status == common.StatusPublished {
		return article, nil
	}

	article.Status = common.StatusPublished
	article.UpdatedAt = time.Now()

	return article, nil
}

func (s *Store) ArchiveArticle(articleID string, operatorID string) (*common.Article, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	article, exists := s.articles[articleID]
	if !exists {
		return nil, nil
	}

	if article.Status == common.StatusArchived {
		return nil, nil
	}

	if article.AuthorID != operatorID {
		return nil, nil
	}

	article.Status = common.StatusArchived
	article.UpdatedAt = time.Now()

	return article, nil
}

func (s *Store) GetVersions(articleID string) ([]*common.ArticleVersion, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	versions, exists := s.versions[articleID]
	if versions == nil {
		return []*common.ArticleVersion{}, exists
	}
	return versions, exists
}

func (s *Store) GetVersion(articleID string, versionNum int) (*common.ArticleVersion, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	versions, exists := s.versions[articleID]
	if !exists {
		return nil, false
	}

	for _, v := range versions {
		if v.VersionNum == versionNum {
			return v, true
		}
	}
	return nil, false
}

func (s *Store) RollbackToVersion(articleID string, versionNum int, authorID string) (*common.Article, *common.ArticleVersion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	article, exists := s.articles[articleID]
	if !exists {
		return nil, nil, nil
	}

	if article.Status == common.StatusArchived {
		return nil, nil, nil
	}

	if article.AuthorID != authorID {
		return nil, nil, nil
	}

	versions := s.versions[articleID]
	var targetVersion *common.ArticleVersion
	for _, v := range versions {
		if v.VersionNum == versionNum {
			targetVersion = v
			break
		}
	}

	if targetVersion == nil {
		return nil, nil, nil
	}

	wasPublished := (article.Status == common.StatusPublished)

	now := time.Now()
	newVersionNum := article.CurrentVersion + 1

	newVersion := &common.ArticleVersion{
		ID:         s.generateIDLocked(),
		ArticleID:  articleID,
		VersionNum: newVersionNum,
		Title:      targetVersion.Title,
		Content:    targetVersion.Content,
		Category:   targetVersion.Category,
		Tags:       targetVersion.Tags,
		CreatedBy:  authorID,
		CreatedAt:  now,
	}

	article.Title = targetVersion.Title
	article.Content = targetVersion.Content
	article.Category = targetVersion.Category
	article.Tags = targetVersion.Tags
	article.UpdatedAt = now
	article.CurrentVersion = newVersionNum

	if wasPublished {
		article.Status = common.StatusDraft
	}

	s.versions[articleID] = append(s.versions[articleID], newVersion)

	return article, newVersion, nil
}

func (s *Store) Search(keyword string, userID string, department string, isLoggedIn bool) []*common.SearchResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*common.SearchResult, 0)
	keywordLower := strings.ToLower(keyword)

	for _, article := range s.articles {
		if !s.canAccessLocked(article, userID, department, isLoggedIn) {
			continue
		}

		if article.Status == common.StatusDraft && article.AuthorID != userID {
			continue
		}

		relevance := 0.0

		if strings.Contains(strings.ToLower(article.Title), keywordLower) {
			relevance += 5.0
		}
		if strings.Contains(strings.ToLower(article.Content), keywordLower) {
			relevance += 3.0
		}
		if strings.Contains(strings.ToLower(article.Category), keywordLower) {
			relevance += 2.0
		}
		for _, tag := range article.Tags {
			if strings.Contains(strings.ToLower(tag), keywordLower) {
				relevance += 2.0
				break
			}
		}

		if relevance > 0 {
			viewBoost := float64(article.ViewCount) / 100.0
			relevance += viewBoost

			results = append(results, &common.SearchResult{
				Article:    article,
				IsArchived: article.Status == common.StatusArchived,
				Relevance:  relevance,
			})
		}
	}

	return results
}

func (s *Store) canAccessLocked(article *common.Article, userID string, department string, isLoggedIn bool) bool {
	switch article.AccessLevel {
	case common.AccessPublic:
		return true
	case common.AccessLoggedIn:
		return isLoggedIn
	case common.AccessDepartment:
		if !isLoggedIn {
			return false
		}
		for _, deptID := range article.DepartmentIDs {
			if deptID == department {
				return true
			}
		}
		return false
	default:
		return true
	}
}

func (s *Store) AddFavorite(userID string, articleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	favs := s.favorites[userID]
	for _, f := range favs {
		if f.ArticleID == articleID {
			return nil
		}
	}

	fav := &common.Favorite{
		ID:        s.generateIDLocked(),
		UserID:    userID,
		ArticleID: articleID,
		CreatedAt: time.Now(),
	}

	s.favorites[userID] = append(s.favorites[userID], fav)
	return nil
}

func (s *Store) RemoveFavorite(userID string, articleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	favs := s.favorites[userID]
	newFavs := make([]*common.Favorite, 0)
	for _, f := range favs {
		if f.ArticleID != articleID {
			newFavs = append(newFavs, f)
		}
	}
	s.favorites[userID] = newFavs
	return nil
}

func (s *Store) GetFavorites(userID string) []*common.Favorite {
	s.mu.RLock()
	defer s.mu.RUnlock()

	favs := s.favorites[userID]
	if favs == nil {
		return []*common.Favorite{}
	}
	return favs
}

func (s *Store) IncrementViewCount(articleID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if article, exists := s.articles[articleID]; exists {
		article.ViewCount++
	}
}

func (s *Store) GetHotArticles(limit int) []*common.HotArticle {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var articles []*common.Article
	for _, a := range s.articles {
		if a.Status == common.StatusPublished {
			articles = append(articles, a)
		}
	}

	for i := 0; i < len(articles); i++ {
		for j := i + 1; j < len(articles); j++ {
			scoreI := articles[i].ViewCount*2 + articles[i].ReferenceCount*5
			scoreJ := articles[j].ViewCount*2 + articles[j].ReferenceCount*5
			if scoreJ > scoreI {
				articles[i], articles[j] = articles[j], articles[i]
			}
		}
	}

	hot := make([]*common.HotArticle, 0)
	for i := 0; i < limit && i < len(articles); i++ {
		hot = append(hot, &common.HotArticle{
			ArticleID:      articles[i].ID,
			Title:          articles[i].Title,
			ViewCount:      articles[i].ViewCount,
			ReferenceCount: articles[i].ReferenceCount,
		})
	}

	return hot
}

func (s *Store) GetStats() *common.Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var totalSize int64
	draftCount := 0
	publishedCount := 0
	archivedCount := 0

	for _, a := range s.articles {
		totalSize += int64(len(a.Title) + len(a.Content))
		switch a.Status {
		case common.StatusDraft:
			draftCount++
		case common.StatusPublished:
			publishedCount++
		case common.StatusArchived:
			archivedCount++
		}
	}

	for _, versions := range s.versions {
		for _, v := range versions {
			totalSize += int64(len(v.Title) + len(v.Content))
		}
	}

	return &common.Stats{
		TotalArticles:  int64(len(s.articles)),
		TotalSize:      totalSize,
		DraftCount:     draftCount,
		PublishedCount: publishedCount,
		ArchivedCount:  archivedCount,
	}
}

func (s *Store) AddReference(fromArticleID string, toArticleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	refs := s.references[toArticleID]
	for _, r := range refs {
		if r.FromArticleID == fromArticleID {
			return nil
		}
	}

	ref := &common.Reference{
		ID:            s.generateIDLocked(),
		FromArticleID: fromArticleID,
		ToArticleID:   toArticleID,
		CreatedAt:     time.Now(),
	}

	s.references[toArticleID] = append(s.references[toArticleID], ref)

	if article, exists := s.articles[toArticleID]; exists {
		article.ReferenceCount++
	}

	return nil
}

func (s *Store) GetRelatedArticles(articleID string) []*common.Article {
	s.mu.RLock()
	defer s.mu.RUnlock()

	article, exists := s.articles[articleID]
	if !exists {
		return []*common.Article{}
	}

	related := make([]*common.Article, 0)
	for _, id := range article.RelatedArticleIDs {
		if a, ok := s.articles[id]; ok {
			if a.Status == common.StatusPublished || a.Status == common.StatusArchived {
				related = append(related, a)
			}
		}
	}

	return related
}
