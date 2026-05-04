const { app, BrowserWindow, ipcMain, Notification, dialog } = require('electron');
const path = require('path');
const dataStore = require('./data-store');
const notificationScheduler = require('./notification-scheduler');

let mainWindow;

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'preload.js')
    },
    icon: path.join(__dirname, 'assets/icon.png')
  });

  mainWindow.loadFile('index.html');
  
  // 打开开发者工具（调试时使用）
  // mainWindow.webContents.openDevTools();
}

app.whenReady().then(() => {
  createWindow();
  notificationScheduler.start();
  
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

// IPC 处理程序
ipcMain.handle('get-habits', () => {
  return dataStore.getHabits();
});

ipcMain.handle('add-habit', (event, habit) => {
  return dataStore.addHabit(habit);
});

ipcMain.handle('delete-habit', (event, habitId) => {
  const confirmResult = dialog.showMessageBoxSync(mainWindow, {
    type: 'warning',
    buttons: ['取消', '确认删除'],
    defaultId: 0,
    title: '确认删除',
    message: '确定要删除这个习惯吗？',
    detail: '该习惯的历史打卡数据将被永久清除且不可恢复。'
  });
  
  if (confirmResult === 1) {
    return dataStore.deleteHabit(habitId);
  }
  return { success: false, cancelled: true };
});

ipcMain.handle('get-checkin-data', (event, year, month) => {
  return dataStore.getCheckinData(year, month);
});

ipcMain.handle('save-checkin', (event, checkin) => {
  return dataStore.saveCheckin(checkin);
});

ipcMain.handle('get-streaks', (event, habitId) => {
  return dataStore.getStreaks(habitId);
});

ipcMain.handle('get-monthly-stats', (event, year, month) => {
  return dataStore.getMonthlyStats(year, month);
});

ipcMain.handle('export-csv', (event, year, month) => {
  return dataStore.exportToCSV(year, month);
});

ipcMain.handle('show-notification', (event, title, body) => {
  new Notification({
    title: title,
    body: body,
    icon: path.join(__dirname, 'assets/icon.png')
  }).show();
});
