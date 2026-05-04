const { app, BrowserWindow, ipcMain, dialog } = require('electron');
const path = require('path');
const database = require('./src/backend/database');
const calculator = require('./src/backend/calculator');
const exporter = require('./src/backend/exporter');

let mainWindow;

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1400,
    height: 900,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false
    }
  });

  mainWindow.loadFile('src/frontend/index.html');

  // 开发环境下打开开发者工具
  // mainWindow.webContents.openDevTools();
}

app.whenReady().then(() => {
  database.initialize();
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

// IPC 处理 - 记录操作
ipcMain.handle('add-record', async (event, record) => {
  return database.addRecord(record);
});

ipcMain.handle('update-record', async (event, id, record) => {
  return database.updateRecord(id, record);
});

ipcMain.handle('delete-record', async (event, id) => {
  return database.deleteRecord(id);
});

ipcMain.handle('get-records-by-date', async (event, date) => {
  return database.getRecordsByDate(date);
});

ipcMain.handle('get-records-by-month', async (event, year, month) => {
  return database.getRecordsByMonth(year, month);
});

ipcMain.handle('search-records', async (event, filters) => {
  return database.searchRecords(filters);
});

// IPC 处理 - 分类操作
ipcMain.handle('get-categories', async (event) => {
  return database.getCategories();
});

ipcMain.handle('add-category', async (event, category) => {
  return database.addCategory(category);
});

ipcMain.handle('delete-category', async (event, name) => {
  return database.deleteCategory(name);
});

// IPC 处理 - 预算操作
ipcMain.handle('get-budgets', async (event, year, month) => {
  return database.getBudgets(year, month);
});

ipcMain.handle('set-budget', async (event, year, month, category, amount) => {
  return database.setBudget(year, month, category, amount);
});

// IPC 处理 - 统计计算
ipcMain.handle('get-monthly-stats', async (event, year, month) => {
  return calculator.getMonthlyStats(year, month);
});

ipcMain.handle('get-daily-expenses', async (event, year, month) => {
  return calculator.getDailyExpenses(year, month);
});

ipcMain.handle('get-category-stats', async (event, year, month) => {
  return calculator.getCategoryStats(year, month);
});

// IPC 处理 - 预算状态
ipcMain.handle('get-budget-status', async (event, year, month) => {
  return calculator.getBudgetStatus(year, month);
});

// IPC 处理 - 导出功能
ipcMain.handle('export-to-csv', async (event, year, month) => {
  const records = database.getRecordsByMonth(year, month);
  const defaultPath = `expenses_${year}_${month.toString().padStart(2, '0')}.csv`;
  
  const { filePath } = await dialog.showSaveDialog(mainWindow, {
    title: '导出 CSV 文件',
    defaultPath: defaultPath,
    filters: [{ name: 'CSV 文件', extensions: ['csv'] }]
  });

  if (filePath) {
    return exporter.exportToCSV(records, filePath);
  }
  
  return { success: false, message: '取消导出' };
});
