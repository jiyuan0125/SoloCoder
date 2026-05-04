const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  selectDirectory: () => ipcRenderer.invoke('select-directory'),
  
  indexDirectory: (directory, extensions) => 
    ipcRenderer.invoke('index-directory', directory, extensions),
  
  reindexDirectory: (directory, extensions) => 
    ipcRenderer.invoke('reindex-directory', directory, extensions),
  
  search: (query, options) => 
    ipcRenderer.invoke('search', query, options),
  
  loadCache: (directory) => 
    ipcRenderer.invoke('load-cache', directory),
  
  openFile: (filePath) => 
    ipcRenderer.invoke('open-file', filePath),
  
  getIndexStats: () => 
    ipcRenderer.invoke('get-index-stats'),

  onIndexProgress: (callback) => {
    ipcRenderer.on('index-progress', (event, progress) => callback(progress));
  },

  onFileScanning: (callback) => {
    ipcRenderer.on('index-file-scanning', (event, filePath) => callback(filePath));
  },

  removeIndexProgressListener: () => {
    ipcRenderer.removeAllListeners('index-progress');
  },

  removeFileScanningListener: () => {
    ipcRenderer.removeAllListeners('index-file-scanning');
  }
});
