const { app, BrowserWindow, ipcMain, dialog } = require('electron');
const path = require('path');
const fs = require('fs');
const { marked } = require('marked');

let mainWindow;
let notesDirectory = path.join(app.getPath('documents'), 'MarkdownNotes');

function ensureNotesDirectory() {
  if (!fs.existsSync(notesDirectory)) {
    fs.mkdirSync(notesDirectory, { recursive: true });
  }
}

function createWindow() {
  ensureNotesDirectory();
  
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    minWidth: 800,
    minHeight: 600,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false
    }
  });

  mainWindow.loadFile('index.html');
}

app.whenReady().then(createWindow);

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    createWindow();
  }
});

function readDirectory(dir, relativePath = '') {
  const items = [];
  try {
    const files = fs.readdirSync(dir, { withFileTypes: true });
    for (const file of files) {
      const fullPath = path.join(dir, file.name);
      const itemRelativePath = path.join(relativePath, file.name);
      
      if (file.isDirectory()) {
        items.push({
          name: file.name,
          path: fullPath,
          relativePath: itemRelativePath,
          type: 'directory',
          children: readDirectory(fullPath, itemRelativePath)
        });
      } else if (file.isFile() && path.extname(file.name).toLowerCase() === '.md') {
        const stats = fs.statSync(fullPath);
        items.push({
          name: file.name,
          path: fullPath,
          relativePath: itemRelativePath,
          type: 'file',
          size: stats.size
        });
      }
    }
    items.sort((a, b) => {
      if (a.type !== b.type) {
        return a.type === 'directory' ? -1 : 1;
      }
      return a.name.localeCompare(b.name);
    });
  } catch (err) {
    console.error('读取目录失败:', err);
  }
  return items;
}

ipcMain.handle('get-directory-tree', async () => {
  ensureNotesDirectory();
  return {
    rootPath: notesDirectory,
    children: readDirectory(notesDirectory)
  };
});

ipcMain.handle('read-file', async (event, filePath) => {
  try {
    const stats = fs.statSync(filePath);
    const content = fs.readFileSync(filePath, 'utf-8');
    return {
      success: true,
      content: content,
      size: stats.size
    };
  } catch (err) {
    return {
      success: false,
      error: err.message
    };
  }
});

ipcMain.handle('save-file', async (event, filePath, content) => {
  try {
    const dir = path.dirname(filePath);
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
    }
    fs.writeFileSync(filePath, content, 'utf-8');
    return { success: true };
  } catch (err) {
    return {
      success: false,
      error: err.message
    };
  }
});

ipcMain.handle('create-file', async (event, parentPath, name) => {
  const filePath = path.join(parentPath, name);
  try {
    if (fs.existsSync(filePath)) {
      return {
        success: false,
        error: '文件已存在'
      };
    }
    fs.writeFileSync(filePath, '', 'utf-8');
    return { success: true, path: filePath };
  } catch (err) {
    return {
      success: false,
      error: err.message
    };
  }
});

ipcMain.handle('create-directory', async (event, parentPath, name) => {
  const dirPath = path.join(parentPath, name);
  try {
    if (fs.existsSync(dirPath)) {
      return {
        success: false,
        error: '文件夹已存在'
      };
    }
    fs.mkdirSync(dirPath, { recursive: true });
    return { success: true, path: dirPath };
  } catch (err) {
    return {
      success: false,
      error: err.message
    };
  }
});

ipcMain.handle('rename-item', async (event, oldPath, newName) => {
  const dir = path.dirname(oldPath);
  const newPath = path.join(dir, newName);
  try {
    if (fs.existsSync(newPath)) {
      return {
        success: false,
        error: '目标已存在',
        conflict: true
      };
    }
    fs.renameSync(oldPath, newPath);
    return { success: true, newPath: newPath };
  } catch (err) {
    return {
      success: false,
      error: err.message
    };
  }
});

ipcMain.handle('delete-item', async (event, itemPath, itemType) => {
  try {
    if (itemType === 'directory') {
      fs.rmSync(itemPath, { recursive: true, force: true });
    } else {
      fs.unlinkSync(itemPath);
    }
    return { success: true };
  } catch (err) {
    return {
      success: false,
      error: err.message
    };
  }
});

ipcMain.handle('check-exists', async (event, filePath) => {
  return fs.existsSync(filePath);
});

ipcMain.handle('render-markdown', async (event, markdown) => {
  try {
    return marked.parse(markdown);
  } catch (err) {
    return `<p class="error">渲染失败: ${err.message}</p>`;
  }
});
