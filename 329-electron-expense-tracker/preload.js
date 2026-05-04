const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  // 记录操作
  addRecord: (record) => ipcRenderer.invoke('add-record', record),
  updateRecord: (id, record) => ipcRenderer.invoke('update-record', id, record),
  deleteRecord: (id) => ipcRenderer.invoke('delete-record', id),
  getRecordsByDate: (date) => ipcRenderer.invoke('get-records-by-date', date),
  getRecordsByMonth: (year, month) => ipcRenderer.invoke('get-records-by-month', year, month),
  searchRecords: (filters) => ipcRenderer.invoke('search-records', filters),

  // 分类操作
  getCategories: () => ipcRenderer.invoke('get-categories'),
  addCategory: (category) => ipcRenderer.invoke('add-category', category),
  deleteCategory: (name) => ipcRenderer.invoke('delete-category', name),

  // 预算操作
  getBudgets: (year, month) => ipcRenderer.invoke('get-budgets', year, month),
  setBudget: (year, month, category, amount) => ipcRenderer.invoke('set-budget', year, month, category, amount),

  // 统计计算
  getMonthlyStats: (year, month) => ipcRenderer.invoke('get-monthly-stats', year, month),
  getDailyExpenses: (year, month) => ipcRenderer.invoke('get-daily-expenses', year, month),
  getCategoryStats: (year, month) => ipcRenderer.invoke('get-category-stats', year, month),
  getBudgetStatus: (year, month) => ipcRenderer.invoke('get-budget-status', year, month),

  // 导出功能
  exportToCSV: (year, month) => ipcRenderer.invoke('export-to-csv', year, month)
});
