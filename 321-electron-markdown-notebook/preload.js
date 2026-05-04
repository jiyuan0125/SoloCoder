const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  getDirectoryTree: () => ipcRenderer.invoke('get-directory-tree'),
  readFile: (filePath) => ipcRenderer.invoke('read-file', filePath),
  saveFile: (filePath, content) => ipcRenderer.invoke('save-file', filePath, content),
  createFile: (parentPath, name) => ipcRenderer.invoke('create-file', parentPath, name),
  createDirectory: (parentPath, name) => ipcRenderer.invoke('create-directory', parentPath, name),
  renameItem: (oldPath, newName) => ipcRenderer.invoke('rename-item', oldPath, newName),
  deleteItem: (itemPath, itemType) => ipcRenderer.invoke('delete-item', itemPath, itemType),
  checkExists: (filePath) => ipcRenderer.invoke('check-exists', filePath),
  renderMarkdown: (markdown) => ipcRenderer.invoke('render-markdown', markdown)
});
