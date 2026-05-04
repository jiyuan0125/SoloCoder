const MAX_FAILED_ATTEMPTS = 5;
const LOCK_DURATION_SECONDS = 30;

let currentKey = null;
let clipboardTimeout = null;

function init(ipcMain, crypto, storage, clipboard) {
  ipcMain.handle('check-has-master-password', async () => {
    return storage.hasMasterPassword();
  });

  ipcMain.handle('get-lock-status', async () => {
    const lockInfo = storage.getLockInfo();
    const now = Date.now();
    
    if (lockInfo.lockedUntil && now < lockInfo.lockedUntil) {
      const remaining = Math.ceil((lockInfo.lockedUntil - now) / 1000);
      return {
        isLocked: true,
        remainingSeconds: remaining,
        failedAttempts: lockInfo.failedAttempts
      };
    }
    
    return {
      isLocked: false,
      remainingSeconds: 0,
      failedAttempts: lockInfo.failedAttempts
    };
  });

  ipcMain.handle('setup-master-password', async (event, password) => {
    if (storage.hasMasterPassword()) {
      return { success: false, error: '主密码已设置' };
    }
    
    try {
      const salt = crypto.generateSalt();
      const key = await crypto.deriveKey(password, salt);
      const passwordHash = key.toString('hex');
      
      const config = {
        salt: salt.toString('hex'),
        passwordHash: passwordHash,
        iterations: crypto.PBKDF2_ITERATIONS
      };
      
      storage.saveConfig(config);
      
      const defaultData = storage.getDefaultData();
      const encryptedData = await crypto.encrypt(defaultData, key);
      storage.saveEncryptedData(encryptedData);
      
      currentKey = key;
      
      return { success: true };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('verify-master-password', async (event, password) => {
    const lockInfo = storage.getLockInfo();
    const now = Date.now();
    
    if (lockInfo.lockedUntil && now < lockInfo.lockedUntil) {
      const remaining = Math.ceil((lockInfo.lockedUntil - now) / 1000);
      return {
        success: false,
        isLocked: true,
        remainingSeconds: remaining,
        failedAttempts: lockInfo.failedAttempts
      };
    }
    
    const config = storage.getConfig();
    if (!config) {
      return { success: false, error: '未设置主密码' };
    }
    
    try {
      const salt = Buffer.from(config.salt, 'hex');
      const key = await crypto.deriveKey(password, salt);
      const passwordHash = key.toString('hex');
      
      if (passwordHash === config.passwordHash) {
        const encryptedData = storage.loadEncryptedData();
        if (encryptedData) {
          try {
            await crypto.decrypt(encryptedData, key);
          } catch (e) {
            lockInfo.failedAttempts = (lockInfo.failedAttempts || 0) + 1;
            
            if (lockInfo.failedAttempts >= MAX_FAILED_ATTEMPTS) {
              lockInfo.lockedUntil = now + (LOCK_DURATION_SECONDS * 1000);
              const remaining = LOCK_DURATION_SECONDS;
              storage.saveLockInfo(lockInfo);
              
              return {
                success: false,
                isLocked: true,
                remainingSeconds: remaining,
                failedAttempts: lockInfo.failedAttempts
              };
            }
            
            storage.saveLockInfo(lockInfo);
            
            return {
              success: false,
              isLocked: false,
              failedAttempts: lockInfo.failedAttempts,
              maxAttempts: MAX_FAILED_ATTEMPTS
            };
          }
        }
        
        storage.clearLockInfo();
        currentKey = key;
        
        return { success: true };
      } else {
        lockInfo.failedAttempts = (lockInfo.failedAttempts || 0) + 1;
        
        if (lockInfo.failedAttempts >= MAX_FAILED_ATTEMPTS) {
          lockInfo.lockedUntil = now + (LOCK_DURATION_SECONDS * 1000);
          const remaining = LOCK_DURATION_SECONDS;
          storage.saveLockInfo(lockInfo);
          
          return {
            success: false,
            isLocked: true,
            remainingSeconds: remaining,
            failedAttempts: lockInfo.failedAttempts
          };
        }
        
        storage.saveLockInfo(lockInfo);
        
        return {
          success: false,
          isLocked: false,
          failedAttempts: lockInfo.failedAttempts,
          maxAttempts: MAX_FAILED_ATTEMPTS
        };
      }
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('load-data', async () => {
    if (!currentKey) {
      return { success: false, error: '未解锁' };
    }
    
    try {
      const encryptedData = storage.loadEncryptedData();
      if (!encryptedData) {
        return { success: true, data: storage.getDefaultData() };
      }
      
      const data = await crypto.decrypt(encryptedData, currentKey);
      return { success: true, data };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('save-data', async (event, data) => {
    if (!currentKey) {
      return { success: false, error: '未解锁' };
    }
    
    try {
      const encryptedData = await crypto.encrypt(data, currentKey);
      storage.saveEncryptedData(encryptedData);
      return { success: true };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('generate-password', async (event, options) => {
    try {
      const password = crypto.generatePassword(options);
      return { success: true, password };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('copy-to-clipboard', async (event, text) => {
    try {
      clipboard.writeText(text);
      
      if (clipboardTimeout) {
        clearTimeout(clipboardTimeout);
      }
      
      clipboardTimeout = setTimeout(() => {
        clipboard.clear();
        clipboardTimeout = null;
      }, 3000);
      
      return { success: true };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('lock-vault', async () => {
    currentKey = null;
    if (clipboardTimeout) {
      clearTimeout(clipboardTimeout);
      clipboard.clear();
      clipboardTimeout = null;
    }
    return { success: true };
  });
}

module.exports = { init };
