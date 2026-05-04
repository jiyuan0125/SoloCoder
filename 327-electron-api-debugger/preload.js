const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  sendRequest: (requestData) => ipcRenderer.invoke('send-request', requestData),
  getHistory: () => ipcRenderer.invoke('get-history'),
  saveHistory: (history) => ipcRenderer.invoke('save-history', history),
  getEnvironments: () => ipcRenderer.invoke('get-environments'),
  saveEnvironments: (environments) => ipcRenderer.invoke('save-environments', environments),
  getSettings: () => ipcRenderer.invoke('get-settings'),
  saveSettings: (settings) => ipcRenderer.invoke('save-settings', settings),
  selectFiles: () => ipcRenderer.invoke('select-files'),
  getFileInfo: (filePath) => ipcRenderer.invoke('get-file-info', filePath)
});
