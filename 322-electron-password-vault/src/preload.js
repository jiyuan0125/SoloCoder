const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  checkHasMasterPassword: () => ipcRenderer.invoke('check-has-master-password'),
  
  getLockStatus: () => ipcRenderer.invoke('get-lock-status'),
  
  setupMasterPassword: (password) => ipcRenderer.invoke('setup-master-password', password),
  
  verifyMasterPassword: (password) => ipcRenderer.invoke('verify-master-password', password),
  
  loadData: () => ipcRenderer.invoke('load-data'),
  
  saveData: (data) => ipcRenderer.invoke('save-data', data),
  
  generatePassword: (options) => ipcRenderer.invoke('generate-password', options),
  
  copyToClipboard: (text) => ipcRenderer.invoke('copy-to-clipboard', text),
  
  lockVault: () => ipcRenderer.invoke('lock-vault')
});
