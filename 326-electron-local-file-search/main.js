const { app, BrowserWindow, ipcMain, dialog, shell } = require('electron');
const path = require('path');
const FileIndexer = require('./src/indexer');

let mainWindow;
let indexer = null;

function initIndexer() {
  if (!indexer) {
    indexer = new FileIndexer();
    
    indexer.on('progress', (progress) => {
      if (mainWindow && !mainWindow.isDestroyed()) {
        mainWindow.webContents.send('index-progress', progress);
      }
    });

    indexer.on('file-scanning', (filePath) => {
      if (mainWindow && !mainWindow.isDestroyed()) {
        mainWindow.webContents.send('index-file-scanning', filePath);
      }
    });
  }
  return indexer;
}

async function performIndexing(directory, extensions) {
  const idx = initIndexer();
  try {
    const result = await idx.indexDirectory(directory, extensions);
    return { success: true, data: result };
  } catch (error) {
    return { success: false, error: error.message };
  }
}

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    minWidth: 800,
    minHeight: 600,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false
    }
  });

  mainWindow.loadFile(path.join(__dirname, 'src', 'index.html'));

  if (process.env.NODE_ENV === 'development') {
    mainWindow.webContents.openDevTools();
  }
}

app.whenReady().then(() => {
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

ipcMain.handle('select-directory', async () => {
  const result = await dialog.showOpenDialog(mainWindow, {
    properties: ['openDirectory']
  });
  return result.canceled ? null : result.filePaths[0];
});

ipcMain.handle('index-directory', async (event, directory, extensions) => {
  return performIndexing(directory, extensions);
});

ipcMain.handle('reindex-directory', async (event, directory, extensions) => {
  const idx = initIndexer();
  idx.clearCache();
  return performIndexing(directory, extensions);
});

ipcMain.handle('search', async (event, query, options = {}) => {
  if (!indexer) {
    return { success: false, error: '索引未初始化' };
  }

  try {
    const results = indexer.search(query, options);
    return { success: true, data: results };
  } catch (error) {
    return { success: false, error: error.message };
  }
});

ipcMain.handle('load-cache', async (event, directory) => {
  const idx = initIndexer();
  const hasCache = idx.hasCache(directory);
  if (hasCache) {
    try {
      const index = idx.loadCache(directory);
      return { success: true, hasCache: true, data: index };
    } catch (error) {
      return { success: true, hasCache: false };
    }
  }
  return { success: true, hasCache: false };
});

ipcMain.handle('open-file', async (event, filePath) => {
  try {
    await shell.openPath(filePath);
    return { success: true };
  } catch (error) {
    return { success: false, error: error.message };
  }
});

ipcMain.handle('get-index-stats', async () => {
  if (!indexer) {
    return { success: false, error: '索引未初始化' };
  }
  return { success: true, data: indexer.getStats() };
});
