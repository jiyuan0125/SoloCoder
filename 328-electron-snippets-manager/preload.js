const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  
  categories: {
    getAll: () => ipcRenderer.invoke('categories:getAll'),
    add: (name) => ipcRenderer.invoke('categories:add', name),
    update: (id, name) => ipcRenderer.invoke('categories:update', id, name),
    delete: (id) => ipcRenderer.invoke('categories:delete', id)
  },
  
  snippets: {
    getAll: () => ipcRenderer.invoke('snippets:getAll'),
    getByCategory: (categoryId) => ipcRenderer.invoke('snippets:getByCategory', categoryId),
    getById: (id) => ipcRenderer.invoke('snippets:getById', id),
    create: (data) => ipcRenderer.invoke('snippets:create', data),
    update: (id, data) => ipcRenderer.invoke('snippets:update', id, data),
    delete: (id) => ipcRenderer.invoke('snippets:delete', id),
    toggleFavorite: (id) => ipcRenderer.invoke('snippets:toggleFavorite', id),
    search: (keyword) => ipcRenderer.invoke('snippets:search', keyword),
    getFavorites: () => ipcRenderer.invoke('snippets:getFavorites')
  },
  
  dialog: {
    showSaveDialog: (options) => ipcRenderer.invoke('dialog:showSaveDialog', options),
    showOpenDialog: (options) => ipcRenderer.invoke('dialog:showOpenDialog', options)
  },
  
  export: {
    all: (filePath) => ipcRenderer.invoke('export:all', filePath),
    category: (categoryId, filePath) => ipcRenderer.invoke('export:category', categoryId, filePath)
  },
  
  import: {
    fromFile: (filePath, options) => ipcRenderer.invoke('import:fromFile', filePath, options)
  },
  
  clipboard: {
    writeText: (text) => {
      return navigator.clipboard.writeText(text);
    }
  }
});
