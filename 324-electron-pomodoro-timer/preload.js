const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  getInitialState: () => ipcRenderer.invoke('get-initial-state'),
  
  startTimer: () => ipcRenderer.invoke('start-timer'),
  pauseTimer: () => ipcRenderer.invoke('pause-timer'),
  skipSession: () => ipcRenderer.invoke('skip-session'),
  setMode: (mode) => ipcRenderer.invoke('set-mode', mode),
  
  addTask: (taskData) => ipcRenderer.invoke('add-task', taskData),
  deleteTask: (taskId) => ipcRenderer.invoke('delete-task', taskId),
  toggleTaskCompletion: (taskId) => ipcRenderer.invoke('toggle-task-completion', taskId),
  selectTask: (taskIndex) => ipcRenderer.invoke('select-task', taskIndex),
  reorderTasks: (newOrder) => ipcRenderer.invoke('reorder-tasks', newOrder),
  
  getStatistics: () => ipcRenderer.invoke('get-statistics'),
  
  onTimerTick: (callback) => ipcRenderer.on('timer-tick', (event, data) => callback(data)),
  onTimerStateChanged: (callback) => ipcRenderer.on('timer-state-changed', (event, data) => callback(data)),
  onTimerComplete: (callback) => ipcRenderer.on('timer-complete', (event, data) => callback(data)),
  onModeChanged: (callback) => ipcRenderer.on('mode-changed', (event, data) => callback(data)),
  onSessionSkipped: (callback) => ipcRenderer.on('session-skipped', (event, data) => callback(data)),
  onInitialStateLoaded: (callback) => ipcRenderer.on('initial-state-loaded', (event, data) => callback(data)),
  onBreakSkipped: (callback) => ipcRenderer.on('break-skipped', (event) => callback()),
  
  removeAllListeners: () => {
    ipcRenderer.removeAllListeners('timer-tick');
    ipcRenderer.removeAllListeners('timer-state-changed');
    ipcRenderer.removeAllListeners('timer-complete');
    ipcRenderer.removeAllListeners('mode-changed');
    ipcRenderer.removeAllListeners('session-skipped');
    ipcRenderer.removeAllListeners('initial-state-loaded');
    ipcRenderer.removeAllListeners('break-skipped');
  }
});
