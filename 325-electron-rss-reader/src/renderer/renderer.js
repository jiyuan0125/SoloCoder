class App {
  constructor() {
    this.feeds = [];
    this.categories = [];
    this.articles = [];
    this.currentFeedId = null;
    this.currentArticleId = null;
    this.currentView = 'all';
    this.sortOrder = 'date-desc';
    this.searchKeyword = '';
    this.init();
  }

  async init() {
    this.setupEventListeners();
    await this.loadData();
    this.setupFeedsUpdatedListener();
  }

  setupFeedsUpdatedListener() {
    window.electronAPI.onFeedsUpdated(async () => {
      await this.loadData();
    });
  }

  setupEventListeners() {
    document.getElementById('btn-add-feed').addEventListener('click', () => this.showModal('modal-add-feed'));
    document.getElementById('btn-add-category').addEventListener('click', () => this.showModal('modal-add-category'));
    document.getElementById('btn-settings').addEventListener('click', () => this.showModal('modal-settings'));
    
    document.getElementById('btn-refresh').addEventListener('click', () => this.refreshCurrentFeed());
    document.getElementById('btn-mark-all-read').addEventListener('click', () => this.markAllRead());
    
    document.getElementById('btn-import-opml').addEventListener('click', () => this.importOpml());
    document.getElementById('btn-export-opml').addEventListener('click', () => this.exportOpml());

    document.getElementById('search-input').addEventListener('input', (e) => {
      this.searchKeyword = e.target.value;
      this.renderArticleList();
    });

    document.getElementById('sort-select').addEventListener('change', (e) => {
      this.sortOrder = e.target.value;
      this.renderArticleList();
    });

    document.querySelectorAll('.modal-close, .modal-cancel').forEach(btn => {
      btn.addEventListener('click', () => this.closeModal(btn.closest('.modal')));
    });

    document.querySelectorAll('.modal').forEach(modal => {
      modal.addEventListener('click', (e) => {
        if (e.target === modal) this.closeModal(modal);
      });
    });

    document.getElementById('form-add-feed').addEventListener('submit', (e) => this.handleAddFeed(e));
    document.getElementById('form-add-category').addEventListener('submit', (e) => this.handleAddCategory(e));
    document.getElementById('form-settings').addEventListener('submit', (e) => this.handleSaveSettings(e));

    document.querySelectorAll('.feed-item.special').forEach(item => {
      item.addEventListener('click', () => this.selectSpecialView(item.dataset.view));
    });

    document.getElementById('btn-open-link').addEventListener('click', () => this.openCurrentArticleLink());
    document.getElementById('btn-toggle-read').addEventListener('click', () => this.toggleCurrentArticleRead());
    document.getElementById('btn-toggle-favorite').addEventListener('click', () => this.toggleCurrentArticleFavorite());
  }

  async loadData() {
    try {
      this.feeds = await window.electronAPI.getAllFeeds();
      this.categories = await window.electronAPI.getAllCategories();
      this.renderFeedList();
      this.renderCategorySelect();
      await this.loadSettings();
    } catch (error) {
      console.error('Failed to load data:', error);
    }
  }

  async loadSettings() {
    try {
      const settings = await window.electronAPI.getSettings();
      if (settings.refreshInterval) {
        document.getElementById('refresh-interval').value = settings.refreshInterval;
      }
    } catch (error) {
      console.error('Failed to load settings:', error);
    }
  }

  renderFeedList() {
    const container = document.getElementById('categories-list');
    container.innerHTML = '';

    const uncategorizedFeeds = this.feeds.filter(f => !f.categoryId);
    uncategorizedFeeds.forEach(feed => {
      container.appendChild(this.createFeedItem(feed));
    });

    this.categories.forEach(category => {
      const categoryFeeds = this.feeds.filter(f => f.categoryId === category.id);
      if (categoryFeeds.length > 0) {
        const categoryGroup = document.createElement('div');
        categoryGroup.className = 'category-group';
        
        const header = document.createElement('div');
        header.className = 'category-header';
        header.textContent = category.name;
        
        const feedsContainer = document.createElement('div');
        feedsContainer.className = 'category-feeds';
        
        categoryFeeds.forEach(feed => {
          feedsContainer.appendChild(this.createFeedItem(feed));
        });

        categoryGroup.appendChild(header);
        categoryGroup.appendChild(feedsContainer);
        container.appendChild(categoryGroup);
      }
    });

    this.updateUnreadCounts();
  }

  createFeedItem(feed) {
    const item = document.createElement('div');
    item.className = 'feed-item';
    item.dataset.feedId = feed.id;
    if (this.currentFeedId === feed.id) {
      item.classList.add('active');
    }

    item.innerHTML = `
      <span class="feed-title">${this.escapeHtml(feed.title)}</span>
      <span class="unread-badge" data-feed-id="${feed.id}">0</span>
    `;

    item.addEventListener('click', () => this.selectFeed(feed.id));
    
    return item;
  }

  async updateUnreadCounts() {
    try {
      const allArticles = [];
      for (const feed of this.feeds) {
        const articles = await window.electronAPI.getArticlesByFeed(feed.id);
        articles.forEach(a => allArticles.push(a));
        
        const unreadCount = articles.filter(a => !a.isRead).length;
        const badge = document.querySelector(`.unread-badge[data-feed-id="${feed.id}"]`);
        if (badge) {
          badge.textContent = unreadCount > 0 ? unreadCount : '';
        }
      }

      const totalUnread = allArticles.filter(a => !a.isRead).length;
      document.getElementById('unread-all').textContent = totalUnread > 0 ? totalUnread : '';

      const favorites = await window.electronAPI.getFavorites();
      const favUnread = favorites.filter(a => !a.isRead).length;
      document.getElementById('unread-favorites').textContent = favUnread > 0 ? favUnread : '';
    } catch (error) {
      console.error('Failed to update unread counts:', error);
    }
  }

  renderCategorySelect() {
    const select = document.getElementById('feed-category');
    select.innerHTML = '<option value="">无分类</option>';
    
    this.categories.forEach(category => {
      const option = document.createElement('option');
      option.value = category.id;
      option.textContent = this.escapeHtml(category.name);
      select.appendChild(option);
    });
  }

  async selectFeed(feedId) {
    this.currentFeedId = feedId;
    this.currentView = null;
    this.searchKeyword = '';
    document.getElementById('search-input').value = '';

    document.querySelectorAll('.feed-item').forEach(item => {
      item.classList.remove('active');
    });
    document.querySelector(`[data-feed-id="${feedId}"]`)?.classList.add('active');

    const feed = this.feeds.find(f => f.id === feedId);
    if (feed) {
      document.getElementById('article-list-title').textContent = feed.title;
    }

    this.articles = await window.electronAPI.getArticlesByFeed(feedId);
    this.renderArticleList();
    this.clearArticleDetail();
  }

  async selectSpecialView(view) {
    this.currentView = view;
    this.currentFeedId = null;

    document.querySelectorAll('.feed-item').forEach(item => {
      item.classList.remove('active');
    });
    document.querySelector(`[data-view="${view}"]`)?.classList.add('active');

    if (view === 'all') {
      document.getElementById('article-list-title').textContent = '全部文章';
      this.articles = [];
      for (const feed of this.feeds) {
        const articles = await window.electronAPI.getArticlesByFeed(feed.id);
        this.articles.push(...articles);
      }
    } else if (view === 'favorites') {
      document.getElementById('article-list-title').textContent = '收藏夹';
      this.articles = await window.electronAPI.getFavorites();
    }

    this.renderArticleList();
    this.clearArticleDetail();
  }

  renderArticleList() {
    const container = document.getElementById('article-list');
    container.innerHTML = '';

    let articles = [...this.articles];

    if (this.searchKeyword) {
      const keyword = this.searchKeyword.toLowerCase();
      articles = articles.filter(a => 
        a.title?.toLowerCase().includes(keyword) ||
        a.summary?.toLowerCase().includes(keyword)
      );
    }

    this.sortArticles(articles);

    if (articles.length === 0) {
      container.innerHTML = '<p class="empty-message">暂无文章</p>';
      return;
    }

    articles.forEach(article => {
      container.appendChild(this.createArticleItem(article));
    });
  }

  sortArticles(articles) {
    switch (this.sortOrder) {
      case 'date-desc':
        articles.sort((a, b) => new Date(b.pubDate) - new Date(a.pubDate));
        break;
      case 'date-asc':
        articles.sort((a, b) => new Date(a.pubDate) - new Date(b.pubDate));
        break;
      case 'title-asc':
        articles.sort((a, b) => (a.title || '').localeCompare(b.title || ''));
        break;
      case 'title-desc':
        articles.sort((a, b) => (b.title || '').localeCompare(a.title || ''));
        break;
    }
  }

  createArticleItem(article) {
    const item = document.createElement('div');
    item.className = `article-item ${article.isRead ? 'read' : 'unread'}`;
    item.dataset.articleId = article.id;
    
    if (this.currentArticleId === article.id) {
      item.classList.add('active');
    }

    const feed = this.feeds.find(f => f.id === article.feedId);
    const dateStr = this.formatDate(article.pubDate);

    item.innerHTML = `
      <div class="article-item-title">${this.escapeHtml(article.title || '无标题')}</div>
      <div class="article-item-meta">
        <span>${feed ? this.escapeHtml(feed.title) : ''}</span>
        <span class="article-item-date">${dateStr}</span>
        ${article.isFavorite ? '<span class="article-item-favorite">★</span>' : ''}
      </div>
    `;

    item.addEventListener('click', () => this.selectArticle(article.id));
    
    return item;
  }

  async selectArticle(articleId) {
    this.currentArticleId = articleId;

    document.querySelectorAll('.article-item').forEach(item => {
      item.classList.remove('active');
    });
    document.querySelector(`.article-item[data-article-id="${articleId}"]`)?.classList.add('active');

    const article = this.articles.find(a => a.id === articleId);
    if (!article) return;

    if (!article.isRead) {
      await window.electronAPI.markArticleRead(articleId, true);
      article.isRead = true;
      
      const item = document.querySelector(`.article-item[data-article-id="${articleId}"]`);
      if (item) {
        item.classList.remove('unread');
        item.classList.add('read');
      }
      
      this.updateUnreadCounts();
    }

    this.renderArticleDetail(article);
  }

  renderArticleDetail(article) {
    document.getElementById('article-title').textContent = article.title || '无标题';
    
    const feed = this.feeds.find(f => f.id === article.feedId);
    document.getElementById('article-source').textContent = feed ? feed.title : '';
    document.getElementById('article-date').textContent = this.formatDate(article.pubDate);

    document.getElementById('btn-toggle-read').textContent = article.isRead ? '标记未读' : '标记已读';
    document.getElementById('btn-toggle-favorite').textContent = article.isFavorite ? '取消收藏' : '收藏';

    const contentContainer = document.getElementById('article-content');
    contentContainer.innerHTML = article.content || article.summary || '<p>暂无内容</p>';
  }

  clearArticleDetail() {
    this.currentArticleId = null;
    document.getElementById('article-title').textContent = '请选择一篇文章';
    document.getElementById('article-source').textContent = '';
    document.getElementById('article-date').textContent = '';
    document.getElementById('article-content').innerHTML = '<p class="empty-message">从左侧选择一篇文章查看详情</p>';
  }

  openCurrentArticleLink() {
    const article = this.articles.find(a => a.id === this.currentArticleId);
    if (article && article.link) {
      window.open(article.link, '_blank');
    }
  }

  async toggleCurrentArticleRead() {
    if (!this.currentArticleId) return;
    
    const article = this.articles.find(a => a.id === this.currentArticleId);
    if (!article) return;

    const newReadStatus = !article.isRead;
    await window.electronAPI.markArticleRead(this.currentArticleId, newReadStatus);
    article.isRead = newReadStatus;

    document.getElementById('btn-toggle-read').textContent = newReadStatus ? '标记未读' : '标记已读';
    
    const item = document.querySelector(`.article-item[data-article-id="${this.currentArticleId}"]`);
    if (item) {
      item.classList.remove('read', 'unread');
      item.classList.add(newReadStatus ? 'read' : 'unread');
    }

    this.updateUnreadCounts();
  }

  async toggleCurrentArticleFavorite() {
    if (!this.currentArticleId) return;
    
    const article = this.articles.find(a => a.id === this.currentArticleId);
    if (!article) return;

    const updated = await window.electronAPI.toggleFavorite(this.currentArticleId);
    if (updated) {
      article.isFavorite = updated.isFavorite;
      document.getElementById('btn-toggle-favorite').textContent = updated.isFavorite ? '取消收藏' : '收藏';
      
      this.renderArticleList();
    }
  }

  async markAllRead() {
    if (this.currentFeedId) {
      await window.electronAPI.markAllRead(this.currentFeedId);
      this.articles.forEach(a => a.isRead = true);
    } else if (this.currentView === 'all') {
      for (const feed of this.feeds) {
        await window.electronAPI.markAllRead(feed.id);
      }
      this.articles.forEach(a => a.isRead = true);
    } else if (this.currentView === 'favorites') {
      const favorites = await window.electronAPI.getFavorites();
      for (const article of favorites) {
        await window.electronAPI.markArticleRead(article.id, true);
      }
      this.articles.forEach(a => a.isRead = true);
    }

    this.renderArticleList();
    this.updateUnreadCounts();
  }

  async refreshCurrentFeed() {
    if (this.currentFeedId) {
      try {
        await window.electronAPI.refreshFeed(this.currentFeedId);
        this.articles = await window.electronAPI.getArticlesByFeed(this.currentFeedId);
        this.renderArticleList();
        this.updateUnreadCounts();
      } catch (error) {
        console.error('Refresh failed:', error);
      }
    } else if (this.currentView === 'all') {
      try {
        await window.electronAPI.refreshAllFeeds();
        this.articles = [];
        for (const feed of this.feeds) {
          const articles = await window.electronAPI.getArticlesByFeed(feed.id);
          this.articles.push(...articles);
        }
        this.renderArticleList();
        this.updateUnreadCounts();
      } catch (error) {
        console.error('Refresh all failed:', error);
      }
    }
  }

  async handleAddFeed(e) {
    e.preventDefault();
    
    const url = document.getElementById('feed-url').value.trim();
    const categoryId = document.getElementById('feed-category').value || null;
    const errorDiv = document.getElementById('feed-error');
    
    if (!url) {
      errorDiv.textContent = '请输入 URL';
      errorDiv.style.display = 'block';
      return;
    }

    try {
      const result = await window.electronAPI.addFeed(url, categoryId);
      
      if (result.success) {
        await this.loadData();
        this.closeModal(document.getElementById('modal-add-feed'));
        document.getElementById('feed-url').value = '';
        errorDiv.style.display = 'none';
      } else {
        errorDiv.textContent = result.error || '添加失败';
        errorDiv.style.display = 'block';
      }
    } catch (error) {
      errorDiv.textContent = '添加失败: ' + error.message;
      errorDiv.style.display = 'block';
    }
  }

  async handleAddCategory(e) {
    e.preventDefault();
    
    const name = document.getElementById('category-name').value.trim();
    
    if (!name) return;

    try {
      await window.electronAPI.addCategory(name);
      await this.loadData();
      this.closeModal(document.getElementById('modal-add-category'));
      document.getElementById('category-name').value = '';
    } catch (error) {
      console.error('Add category failed:', error);
    }
  }

  async handleSaveSettings(e) {
    e.preventDefault();
    
    const refreshInterval = parseInt(document.getElementById('refresh-interval').value);
    
    try {
      await window.electronAPI.updateSettings({ refreshInterval });
      this.closeModal(document.getElementById('modal-settings'));
    } catch (error) {
      console.error('Save settings failed:', error);
    }
  }

  async importOpml() {
    try {
      const result = await window.electronAPI.importOpml();
      if (result.success) {
        await this.loadData();
      }
    } catch (error) {
      console.error('Import OPML failed:', error);
    }
  }

  async exportOpml() {
    try {
      await window.electronAPI.exportOpml();
    } catch (error) {
      console.error('Export OPML failed:', error);
    }
  }

  showModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
      modal.classList.add('active');
    }
  }

  closeModal(modal) {
    if (modal) {
      modal.classList.remove('active');
    }
  }

  escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  formatDate(dateStr) {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    if (isNaN(date.getTime())) return '';
    
    const now = new Date();
    const diff = now - date;
    
    if (diff < 60000) return '刚刚';
    if (diff < 3600000) return Math.floor(diff / 60000) + '分钟前';
    if (diff < 86400000) return Math.floor(diff / 3600000) + '小时前';
    if (diff < 604800000) return Math.floor(diff / 86400000) + '天前';
    
    return date.toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit'
    });
  }
}

document.addEventListener('DOMContentLoaded', () => {
  new App();
});
