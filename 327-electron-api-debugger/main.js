const { app, BrowserWindow, ipcMain, dialog } = require('electron');
const path = require('path');
const fs = require('fs');
const https = require('https');
const http = require('http');
const FormData = require('form-data');
const { URL } = require('url');

const DATA_DIR = path.join(app.getPath('userData'), 'api-debugger-data');
const HISTORY_FILE = path.join(DATA_DIR, 'history.json');
const ENVIRONMENTS_FILE = path.join(DATA_DIR, 'environments.json');
const SETTINGS_FILE = path.join(DATA_DIR, 'settings.json');

let mainWindow;

function ensureDataDir() {
  if (!fs.existsSync(DATA_DIR)) {
    fs.mkdirSync(DATA_DIR, { recursive: true });
  }
}

function loadJSONFile(filePath, defaultValue) {
  if (fs.existsSync(filePath)) {
    try {
      const data = fs.readFileSync(filePath, 'utf-8');
      return JSON.parse(data);
    } catch (e) {
      return defaultValue;
    }
  }
  return defaultValue;
}

function saveJSONFile(filePath, data) {
  try {
    fs.writeFileSync(filePath, JSON.stringify(data, null, 2), 'utf-8');
    return true;
  } catch (e) {
    console.error('Error saving file:', e);
    return false;
  }
}

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1400,
    height: 900,
    minWidth: 1000,
    minHeight: 700,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false
    }
  });

  mainWindow.loadFile('index.html');
  
  if (process.argv.includes('--dev-tools')) {
    mainWindow.webContents.openDevTools();
  }
}

app.whenReady().then(() => {
  ensureDataDir();
  createWindow();

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

ipcMain.handle('send-request', async (event, requestData) => {
  return new Promise((resolve) => {
    try {
      const { method, url, headers, body, bodyType, files, timeout } = requestData;
      const parsedUrl = new URL(url);
      const isHttps = parsedUrl.protocol === 'https:';
      const client = isHttps ? https : http;
      const port = parsedUrl.port || (isHttps ? 443 : 80);

      let requestOptions = {
        method: method,
        hostname: parsedUrl.hostname,
        port: parseInt(port),
        path: parsedUrl.pathname + parsedUrl.search,
        headers: {},
        timeout: timeout || 30000
      };

      Object.keys(headers).forEach(key => {
        if (headers[key]) {
          requestOptions.headers[key] = headers[key];
        }
      });

      let requestBody = null;
      let form = null;

      if (method !== 'GET' && method !== 'HEAD') {
        if (bodyType === 'form-data' && files && files.length > 0) {
          form = new FormData();
          
          if (body) {
            try {
              const formFields = typeof body === 'string' ? JSON.parse(body) : body;
              Object.keys(formFields).forEach(key => {
                form.append(key, formFields[key]);
              });
            } catch (e) {
              console.error('Error parsing form fields:', e);
            }
          }

          files.forEach(fileInfo => {
            if (fs.existsSync(fileInfo.path)) {
              form.append(fileInfo.fieldName, fs.createReadStream(fileInfo.path), {
                filename: fileInfo.filename
              });
            }
          });

          requestOptions.headers = { ...requestOptions.headers, ...form.getHeaders() };
          requestBody = form;
        } else if (bodyType === 'json') {
          requestOptions.headers['Content-Type'] = 'application/json';
          requestBody = body;
        } else if (bodyType === 'form-urlencoded') {
          requestOptions.headers['Content-Type'] = 'application/x-www-form-urlencoded';
          requestBody = body;
        } else {
          requestOptions.headers['Content-Type'] = 'text/plain';
          requestBody = body;
        }
      }

      const requestStartTime = Date.now();
      
      const req = client.request(requestOptions, (res) => {
        let responseData = [];
        const responseSize = parseInt(res.headers['content-length']) || 0;
        
        res.on('data', (chunk) => {
          responseData.push(chunk);
        });

        res.on('end', () => {
          const responseTime = Date.now() - requestStartTime;
          const buffer = Buffer.concat(responseData);
          const responseHeaders = {};
          
          Object.keys(res.headers).forEach(key => {
            responseHeaders[key] = res.headers[key];
          });

          let responseBody = buffer.toString('utf-8');
          let isTruncated = false;
          
          if (buffer.length > 1024 * 1024) {
            responseBody = buffer.slice(0, 100 * 1024).toString('utf-8');
            isTruncated = true;
          }

          resolve({
            success: true,
            statusCode: res.statusCode,
            statusMessage: res.statusMessage,
            headers: responseHeaders,
            body: responseBody,
            bodySize: buffer.length,
            isTruncated: isTruncated,
            responseTime: responseTime
          });
        });
      });

      req.on('timeout', () => {
        req.destroy();
        resolve({
          success: false,
          error: '请求超时',
          statusCode: 0
        });
      });

      req.on('error', (error) => {
        resolve({
          success: false,
          error: error.message,
          statusCode: 0
        });
      });

      if (requestBody) {
        if (form) {
          form.pipe(req);
        } else {
          req.write(requestBody);
          req.end();
        }
      } else {
        req.end();
      }

    } catch (error) {
      resolve({
        success: false,
        error: error.message,
        statusCode: 0
      });
    }
  });
});

ipcMain.handle('get-history', () => {
  return loadJSONFile(HISTORY_FILE, []);
});

ipcMain.handle('save-history', (event, history) => {
  return saveJSONFile(HISTORY_FILE, history);
});

ipcMain.handle('get-environments', () => {
  return loadJSONFile(ENVIRONMENTS_FILE, {
    active: 'default',
    environments: {
      default: {
        name: '默认环境',
        variables: {}
      }
    }
  });
});

ipcMain.handle('save-environments', (event, environments) => {
  return saveJSONFile(ENVIRONMENTS_FILE, environments);
});

ipcMain.handle('get-settings', () => {
  return loadJSONFile(SETTINGS_FILE, {
    timeout: 30000
  });
});

ipcMain.handle('save-settings', (event, settings) => {
  return saveJSONFile(SETTINGS_FILE, settings);
});

ipcMain.handle('select-files', async () => {
  const result = await dialog.showOpenDialog(mainWindow, {
    properties: ['openFile', 'multiSelections'],
    title: '选择文件'
  });
  
  return result.canceled ? [] : result.filePaths;
});

ipcMain.handle('get-file-info', (event, filePath) => {
  try {
    const stats = fs.statSync(filePath);
    return {
      success: true,
      filename: path.basename(filePath),
      size: stats.size
    };
  } catch (e) {
    return {
      success: false,
      error: e.message
    };
  }
});
