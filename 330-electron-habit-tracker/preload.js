const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  getHabits: () => ipcRenderer.invoke('get-habits'),
  addHabit: (habit) => ipcRenderer.invoke('add-habit', habit),
  deleteHabit: (habitId) => ipcRenderer.invoke('delete-habit', habitId),
  getCheckinData: (year, month) => ipcRenderer.invoke('get-checkin-data', year, month),
  saveCheckin: (checkin) => ipcRenderer.invoke('save-checkin', checkin),
  getStreaks: (habitId) => ipcRenderer.invoke('get-streaks', habitId),
  getMonthlyStats: (year, month) => ipcRenderer.invoke('get-monthly-stats', year, month),
  exportCSV: (year, month) => ipcRenderer.invoke('export-csv', year, month),
  showNotification: (title, body) => ipcRenderer.invoke('show-notification', title, body)
});
