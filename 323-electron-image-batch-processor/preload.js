const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  selectFiles: () => ipcRenderer.invoke('select-files'),
  selectFolder: () => ipcRenderer.invoke('select-folder'),
  getThumbnail: (filePath) => ipcRenderer.invoke('get-thumbnail', filePath),
  validateRenameTemplate: (template) => ipcRenderer.invoke('validate-rename-template', template),
  processImages: (options) => ipcRenderer.invoke('process-images', options),
  getComparisonImages: (originalPath, processedPath) => ipcRenderer.invoke('get-comparison-images', originalPath, processedPath),
  openOutputFolder: (folderPath) => ipcRenderer.invoke('open-output-folder', folderPath),
  onProcessingProgress: (callback) => {
    ipcRenderer.on('processing-progress', (event, data) => callback(data));
  },
  removeAllListeners: () => {
    ipcRenderer.removeAllListeners('processing-progress');
  }
});
