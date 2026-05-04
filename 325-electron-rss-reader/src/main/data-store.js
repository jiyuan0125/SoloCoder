const fs = require('fs');
const path = require('path');
const { v4: uuidv4 } = require('uuid');

class DataStore {
  constructor(dataPath) {
    this.dataPath = dataPath;
    this.feedsFile = path.join(dataPath, 'feeds.json');
    this.articlesFile = path.join(dataPath, 'articles.json');
    this.categoriesFile = path.join(dataPath, 'categories.json');
    this.settingsFile = path.join(dataPath, 'settings.json');
    this.ensureDataDirectory();
  }

  ensureDataDirectory() {
    if (!fs.existsSync(this.dataPath)) {
      fs.mkdirSync(this.dataPath, { recursive: true });
    }
  }

  loadJsonFile(filePath, defaultData = []) {
    if (!fs.existsSync(filePath)) {
      return defaultData;
    }
    try {
      const content = fs.readFileSync(filePath, 'utf-8');
      return JSON.parse(content);
    } catch (error) {
      console.error(`Error loading ${filePath}:`, error.message);
      return defaultData;
    }
  }

  saveJsonFile(filePath, data) {
    try {
      fs.writeFileSync(filePath, JSON.stringify(data, null, 2), 'utf-8');
    } catch (error) {
      console.error(`Error saving ${filePath}:`, error.message);
    }
  }

  getAllFeeds() {
    return this.loadJsonFile(this.feedsFile, []);
  }

  getFeed(feedId) {
    const feeds = this.getAllFeeds();
    return feeds.find(f => f.id === feedId);
  }

  getFeedsByCategory(categoryId) {
    const feeds = this.getAllFeeds();
    if (categoryId === null) {
      return feeds.filter(f => !f.categoryId);
    }
    return feeds.filter(f => f.categoryId === categoryId);
  }

  feedExists(url) {
    const feeds = this.getAllFeeds();
    return feeds.some(f => f.url.toLowerCase() === url.toLowerCase());
  }

  addFeed(feed) {
    const feeds = this.getAllFeeds();
    const newFeed = {
      id: uuidv4(),
      ...feed,
      createdAt: new Date().toISOString()
    };
    feeds.push(newFeed);
    this.saveJsonFile(this.feedsFile, feeds);
    return newFeed;
  }

  updateFeed(updatedFeed) {
    const feeds = this.getAllFeeds();
    const index = feeds.findIndex(f => f.id === updatedFeed.id);
    if (index !== -1) {
      feeds[index] = { ...feeds[index], ...updatedFeed };
      this.saveJsonFile(this.feedsFile, feeds);
      return feeds[index];
    }
    return null;
  }

  deleteFeed(feedId) {
    let feeds = this.getAllFeeds();
    feeds = feeds.filter(f => f.id !== feedId);
    this.saveJsonFile(this.feedsFile, feeds);

    let articles = this.getAllArticles();
    articles = articles.filter(a => a.feedId !== feedId);
    this.saveJsonFile(this.articlesFile, articles);
  }

  getAllCategories() {
    return this.loadJsonFile(this.categoriesFile, []);
  }

  getCategory(categoryId) {
    const categories = this.getAllCategories();
    return categories.find(c => c.id === categoryId);
  }

  getCategoryByName(name) {
    const categories = this.getAllCategories();
    return categories.find(c => c.name.toLowerCase() === name.toLowerCase());
  }

  addCategory(category) {
    const categories = this.getAllCategories();
    const newCategory = {
      id: uuidv4(),
      ...category,
      createdAt: new Date().toISOString()
    };
    categories.push(newCategory);
    this.saveJsonFile(this.categoriesFile, categories);
    return newCategory;
  }

  updateCategory(updatedCategory) {
    const categories = this.getAllCategories();
    const index = categories.findIndex(c => c.id === updatedCategory.id);
    if (index !== -1) {
      categories[index] = { ...categories[index], ...updatedCategory };
      this.saveJsonFile(this.categoriesFile, categories);
      return categories[index];
    }
    return null;
  }

  deleteCategory(categoryId) {
    let categories = this.getAllCategories();
    categories = categories.filter(c => c.id !== categoryId);
    this.saveJsonFile(this.categoriesFile, categories);

    const feeds = this.getAllFeeds();
    feeds.forEach(feed => {
      if (feed.categoryId === categoryId) {
        feed.categoryId = null;
      }
    });
    this.saveJsonFile(this.feedsFile, feeds);
  }

  getAllArticles() {
    return this.loadJsonFile(this.articlesFile, []);
  }

  getArticle(articleId) {
    const articles = this.getAllArticles();
    return articles.find(a => a.id === articleId);
  }

  getArticlesByFeed(feedId) {
    const articles = this.getAllArticles();
    return articles
      .filter(a => a.feedId === feedId)
      .sort((a, b) => new Date(b.pubDate) - new Date(a.pubDate));
  }

  articleExists(guid) {
    const articles = this.getAllArticles();
    return articles.some(a => a.guid === guid);
  }

  addArticle(article) {
    const articles = this.getAllArticles();
    const newArticle = {
      id: uuidv4(),
      ...article,
      isRead: false,
      isFavorite: false,
      createdAt: new Date().toISOString()
    };
    articles.push(newArticle);
    this.saveJsonFile(this.articlesFile, articles);
    return newArticle;
  }

  updateArticle(updatedArticle) {
    const articles = this.getAllArticles();
    const index = articles.findIndex(a => a.id === updatedArticle.id);
    if (index !== -1) {
      articles[index] = { ...articles[index], ...updatedArticle };
      this.saveJsonFile(this.articlesFile, articles);
      return articles[index];
    }
    return null;
  }

  markArticleRead(articleId, read) {
    const articles = this.getAllArticles();
    const index = articles.findIndex(a => a.id === articleId);
    if (index !== -1) {
      articles[index].isRead = read;
      this.saveJsonFile(this.articlesFile, articles);
      return articles[index];
    }
    return null;
  }

  markAllRead(feedId) {
    const articles = this.getAllArticles();
    articles.forEach(article => {
      if (article.feedId === feedId) {
        article.isRead = true;
      }
    });
    this.saveJsonFile(this.articlesFile, articles);
  }

  toggleFavorite(articleId) {
    const articles = this.getAllArticles();
    const index = articles.findIndex(a => a.id === articleId);
    if (index !== -1) {
      articles[index].isFavorite = !articles[index].isFavorite;
      this.saveJsonFile(this.articlesFile, articles);
      return articles[index];
    }
    return null;
  }

  getFavorites() {
    const articles = this.getAllArticles();
    return articles
      .filter(a => a.isFavorite)
      .sort((a, b) => new Date(b.pubDate) - new Date(a.pubDate));
  }

  searchArticles(keyword) {
    if (!keyword || keyword.trim() === '') {
      return [];
    }
    const lowerKeyword = keyword.toLowerCase();
    const articles = this.getAllArticles();
    return articles
      .filter(a => 
        a.title?.toLowerCase().includes(lowerKeyword) ||
        a.content?.toLowerCase().includes(lowerKeyword) ||
        a.summary?.toLowerCase().includes(lowerKeyword)
      )
      .sort((a, b) => new Date(b.pubDate) - new Date(a.pubDate));
  }

  getSettings() {
    const defaultSettings = {
      refreshInterval: 30
    };
    return this.loadJsonFile(this.settingsFile, defaultSettings);
  }

  updateSettings(settings) {
    const currentSettings = this.getSettings();
    const updatedSettings = { ...currentSettings, ...settings };
    this.saveJsonFile(this.settingsFile, updatedSettings);
    return updatedSettings;
  }

  getUnreadCount(feedId) {
    const articles = this.getAllArticles();
    return articles.filter(a => a.feedId === feedId && !a.isRead).length;
  }

  getTotalUnreadCount() {
    const articles = this.getAllArticles();
    return articles.filter(a => !a.isRead).length;
  }
}

module.exports = DataStore;
