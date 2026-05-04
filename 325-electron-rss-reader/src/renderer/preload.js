const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  getAllFeeds: () => ipcRenderer.invoke('get-all-feeds'),
  getFeedsByCategory: (categoryId) => ipcRenderer.invoke('get-feeds-by-category', categoryId),
  addFeed: (url, categoryId) => ipcRenderer.invoke('add-feed', url, categoryId),
  updateFeed: (feed) => ipcRenderer.invoke('update-feed', feed),
  deleteFeed: (feedId) => ipcRenderer.invoke('delete-feed', feedId),
  
  getAllCategories: () => ipcRenderer.invoke('get-all-categories'),
  addCategory: (name) => ipcRenderer.invoke('add-category', name),
  updateCategory: (category) => ipcRenderer.invoke('update-category', category),
  deleteCategory: (categoryId) => ipcRenderer.invoke('delete-category', categoryId),
  
  getArticlesByFeed: (feedId) => ipcRenderer.invoke('get-articles-by-feed', feedId),
  getArticle: (articleId) => ipcRenderer.invoke('get-article', articleId),
  markArticleRead: (articleId, read) => ipcRenderer.invoke('mark-article-read', articleId, read),
  markAllRead: (feedId) => ipcRenderer.invoke('mark-all-read', feedId),
  toggleFavorite: (articleId) => ipcRenderer.invoke('toggle-favorite', articleId),
  getFavorites: () => ipcRenderer.invoke('get-favorites'),
  searchArticles: (keyword) => ipcRenderer.invoke('search-articles', keyword),
  
  getSettings: () => ipcRenderer.invoke('get-settings'),
  updateSettings: (settings) => ipcRenderer.invoke('update-settings', settings),
  
  refreshFeed: (feedId) => ipcRenderer.invoke('refresh-feed', feedId),
  refreshAllFeeds: () => ipcRenderer.invoke('refresh-all-feeds'),
  
  exportOpml: () => ipcRenderer.invoke('export-opml'),
  importOpml: () => ipcRenderer.invoke('import-opml'),
  
  onFeedsUpdated: (callback) => {
    ipcRenderer.on('feeds-updated', (event, ...args) => callback(...args));
  },
  
  removeOnFeedsUpdated: (callback) => {
    ipcRenderer.removeListener('feeds-updated', callback);
  }
});
