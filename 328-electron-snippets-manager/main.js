const { app, BrowserWindow, ipcMain, dialog } = require('electron');
const path = require('path');
const StoreService = require('./services/store');
const SnippetService = require('./services/snippets');
const CategoryService = require('./services/categories');
const { setupIPC } = require('./services/ipc');

let mainWindow = null;
let store = null;
let snippetService = null;
let categoryService = null;

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1400,
    height: 900,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'preload.js')
    },
    title: '代码片段管理工具'
  });

  mainWindow.loadFile('renderer/index.html');

  mainWindow.webContents.openDevTools();

  mainWindow.on('closed', () => {
    mainWindow = null;
  });
}

app.whenReady().then(() => {
  const appDataPath = app.getPath('userData');
  store = new StoreService(appDataPath);
  
  categoryService = new CategoryService(store);
  snippetService = new SnippetService(store, categoryService);
  
  setupIPC(ipcMain, dialog, snippetService, categoryService, store);
  
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
