const fs = require('fs');
const path = require('path');
const os = require('os');

const DATA_DIR = path.join(os.homedir(), '.password-vault');
const CONFIG_FILE = path.join(DATA_DIR, 'config.json');
const DATA_FILE = path.join(DATA_DIR, 'vault.dat');
const LOCK_FILE = path.join(DATA_DIR, 'lock.json');

function ensureDataDir() {
  if (!fs.existsSync(DATA_DIR)) {
    fs.mkdirSync(DATA_DIR, { recursive: true });
  }
}

function hasMasterPassword() {
  return fs.existsSync(CONFIG_FILE);
}

function getConfig() {
  if (!fs.existsSync(CONFIG_FILE)) {
    return null;
  }
  const content = fs.readFileSync(CONFIG_FILE, 'utf8');
  return JSON.parse(content);
}

function saveConfig(config) {
  ensureDataDir();
  fs.writeFileSync(CONFIG_FILE, JSON.stringify(config, null, 2), 'utf8');
}

function saveEncryptedData(encryptedData) {
  ensureDataDir();
  fs.writeFileSync(DATA_FILE, JSON.stringify(encryptedData, null, 2), 'utf8');
}

function loadEncryptedData() {
  if (!fs.existsSync(DATA_FILE)) {
    return null;
  }
  const content = fs.readFileSync(DATA_FILE, 'utf8');
  return JSON.parse(content);
}

function getLockInfo() {
  if (!fs.existsSync(LOCK_FILE)) {
    return {
      failedAttempts: 0,
      lockedUntil: null
    };
  }
  const content = fs.readFileSync(LOCK_FILE, 'utf8');
  return JSON.parse(content);
}

function saveLockInfo(lockInfo) {
  ensureDataDir();
  fs.writeFileSync(LOCK_FILE, JSON.stringify(lockInfo, null, 2), 'utf8');
}

function clearLockInfo() {
  if (fs.existsSync(LOCK_FILE)) {
    fs.unlinkSync(LOCK_FILE);
  }
}

function getDefaultData() {
  return {
    categories: [
      { id: 'social', name: '社交', icon: '👥' },
      { id: 'email', name: '邮箱', icon: '📧' },
      { id: 'finance', name: '金融', icon: '💰' },
      { id: 'work', name: '工作', icon: '💼' },
      { id: 'other', name: '其他', icon: '📁' }
    ],
    entries: []
  };
}

module.exports = {
  ensureDataDir,
  hasMasterPassword,
  getConfig,
  saveConfig,
  saveEncryptedData,
  loadEncryptedData,
  getLockInfo,
  saveLockInfo,
  clearLockInfo,
  getDefaultData,
  DATA_DIR,
  CONFIG_FILE,
  DATA_FILE,
  LOCK_FILE
};
