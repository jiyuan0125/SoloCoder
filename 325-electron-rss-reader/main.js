const { app, BrowserWindow, ipcMain, dialog } = require('electron');
const path = require('path');
const DataStore = require('./src/main/data-store');
const RssFetcher = require('./src/main/rss-fetcher');
const OpmlHandler = require('./src/main/opml-handler');

let mainWindow;
let dataStore;
let rssFetcher;
let opmlHandler;

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1400,
    height: 900,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'src/renderer/preload.js')
    }
  });

  mainWindow.loadFile(path.join(__dirname, 'src/renderer/index.html'));

  if (process.env.NODE_ENV === 'development') {
    mainWindow.webContents.openDevTools();
  }
}

function initializeModules() {
  const userDataPath = app.getPath('userData');
  dataStore = new DataStore(userDataPath);
  rssFetcher = new RssFetcher(dataStore);
  opmlHandler = new OpmlHandler();

  startAutoRefresh();
}

function startAutoRefresh() {
  const settings = dataStore.getSettings();
  const refreshInterval = settings.refreshInterval || 30;

  setInterval(() => {
    refreshAllFeeds();
  }, refreshInterval * 60 * 1000);
}

async function refreshAllFeeds() {
  const feeds = dataStore.getAllFeeds();
  for (const feed of feeds) {
    try {
      await rssFetcher.fetchAndUpdateFeed(feed.url);
    } catch (error) {
      console.error(`Failed to refresh feed ${feed.url}:`, error.message);
    }
  }
  if (mainWindow) {
    mainWindow.webContents.send('feeds-updated');
  }
}

app.whenReady().then(() => {
  initializeModules();
  createWindow();

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

ipcMain.handle('get-all-feeds', () => {
  return dataStore.getAllFeeds();
});

ipcMain.handle('get-feeds-by-category', (event, categoryId) => {
  return dataStore.getFeedsByCategory(categoryId);
});

ipcMain.handle('add-feed', async (event, url, categoryId) => {
  const validation = await rssFetcher.validateFeedUrl(url);
  if (!validation.valid) {
    return { success: false, error: validation.error };
  }

  if (dataStore.feedExists(url)) {
    return { success: false, error: '该订阅源已存在' };
  }

  const feed = dataStore.addFeed({
    url,
    title: validation.title || url,
    categoryId: categoryId || null
  });

  try {
    await rssFetcher.fetchAndUpdateFeed(url);
  } catch (error) {
    console.error('Initial fetch failed:', error.message);
  }

  return { success: true, feed };
});

ipcMain.handle('update-feed', (event, feed) => {
  return dataStore.updateFeed(feed);
});

ipcMain.handle('delete-feed', (event, feedId) => {
  return dataStore.deleteFeed(feedId);
});

ipcMain.handle('get-all-categories', () => {
  return dataStore.getAllCategories();
});

ipcMain.handle('add-category', (event, name) => {
  return dataStore.addCategory({ name });
});

ipcMain.handle('update-category', (event, category) => {
  return dataStore.updateCategory(category);
});

ipcMain.handle('delete-category', (event, categoryId) => {
  return dataStore.deleteCategory(categoryId);
});

ipcMain.handle('get-articles-by-feed', (event, feedId) => {
  return dataStore.getArticlesByFeed(feedId);
});

ipcMain.handle('get-article', (event, articleId) => {
  return dataStore.getArticle(articleId);
});

ipcMain.handle('mark-article-read', (event, articleId, read) => {
  return dataStore.markArticleRead(articleId, read);
});

ipcMain.handle('mark-all-read', (event, feedId) => {
  return dataStore.markAllRead(feedId);
});

ipcMain.handle('toggle-favorite', (event, articleId) => {
  return dataStore.toggleFavorite(articleId);
});

ipcMain.handle('get-favorites', () => {
  return dataStore.getFavorites();
});

ipcMain.handle('search-articles', (event, keyword) => {
  return dataStore.searchArticles(keyword);
});

ipcMain.handle('get-settings', () => {
  return dataStore.getSettings();
});

ipcMain.handle('update-settings', (event, settings) => {
  return dataStore.updateSettings(settings);
});

ipcMain.handle('refresh-feed', async (event, feedId) => {
  const feed = dataStore.getFeed(feedId);
  if (!feed) {
    return { success: false, error: '订阅源不存在' };
  }

  try {
    await rssFetcher.fetchAndUpdateFeed(feed.url);
    return { success: true };
  } catch (error) {
    return { success: false, error: error.message };
  }
});

ipcMain.handle('refresh-all-feeds', async () => {
  await refreshAllFeeds();
  return { success: true };
});

ipcMain.handle('export-opml', async () => {
  const result = await dialog.showSaveDialog(mainWindow, {
    title: '导出 OPML',
    defaultPath: 'feeds.opml',
    filters: [{ name: 'OPML 文件', extensions: ['opml', 'xml'] }]
  });

  if (result.canceled) {
    return { success: false };
  }

  const feeds = dataStore.getAllFeeds();
  const categories = dataStore.getAllCategories();
  const opmlContent = opmlHandler.exportOpml(feeds, categories);

  const fs = require('fs');
  fs.writeFileSync(result.filePath, opmlContent, 'utf-8');

  return { success: true, path: result.filePath };
});

ipcMain.handle('import-opml', async () => {
  const result = await dialog.showOpenDialog(mainWindow, {
    title: '导入 OPML',
    properties: ['openFile'],
    filters: [{ name: 'OPML 文件', extensions: ['opml', 'xml'] }]
  });

  if (result.canceled || result.filePaths.length === 0) {
    return { success: false };
  }

  const fs = require('fs');
  const opmlContent = fs.readFileSync(result.filePaths[0], 'utf-8');

  try {
    const parsed = opmlHandler.importOpml(opmlContent);
    
    for (const item of parsed) {
      if (!dataStore.feedExists(item.url)) {
        let categoryId = null;
        if (item.category) {
          let category = dataStore.getCategoryByName(item.category);
          if (!category) {
            category = dataStore.addCategory({ name: item.category });
          }
          categoryId = category.id;
        }

        dataStore.addFeed({
          url: item.url,
          title: item.title || item.url,
          categoryId
        });

        try {
          await rssFetcher.fetchAndUpdateFeed(item.url);
        } catch (error) {
          console.error(`Failed to fetch feed ${item.url}:`, error.message);
        }
      }
    }

    return { success: true };
  } catch (error) {
    return { success: false, error: error.message };
  }
});
