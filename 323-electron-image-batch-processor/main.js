const { app, BrowserWindow, ipcMain, dialog, shell } = require('electron');
const path = require('path');
const fs = require('fs');
const sharp = require('sharp');

let mainWindow;

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'preload.js')
    }
  });

  mainWindow.loadFile(path.join(__dirname, 'src', 'index.html'));

  mainWindow.on('closed', function () {
    mainWindow = null;
  });
}

app.whenReady().then(() => {
  createWindow();

  app.on('activate', function () {
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
  });
});

app.on('window-all-closed', function () {
  if (process.platform !== 'darwin') app.quit();
});

const supportedFormats = ['.jpg', '.jpeg', '.png', '.webp', '.bmp'];

function isImageFile(filePath) {
  const ext = path.extname(filePath).toLowerCase();
  return supportedFormats.includes(ext);
}

function getFilesFromPath(dirPath, files = []) {
  if (fs.statSync(dirPath).isDirectory()) {
    const items = fs.readdirSync(dirPath);
    items.forEach(item => {
      const fullPath = path.join(dirPath, item);
      if (fs.statSync(fullPath).isDirectory()) {
        getFilesFromPath(fullPath, files);
      } else if (isImageFile(fullPath)) {
        files.push(fullPath);
      }
    });
  } else if (isImageFile(dirPath)) {
    files.push(dirPath);
  }
  return files;
}

function generateUniqueOutputPath(outputDir, originalName, newExt) {
  let baseName = path.basename(originalName, path.extname(originalName));
  let outputName = `${baseName}${newExt}`;
  let outputPath = path.join(outputDir, outputName);
  let counter = 1;

  while (fs.existsSync(outputPath)) {
    outputName = `${baseName}_${counter}${newExt}`;
    outputPath = path.join(outputDir, outputName);
    counter++;
  }

  return outputPath;
}

function applyRenameTemplate(template, originalName, index, total, date) {
  const ext = path.extname(originalName);
  const baseName = path.basename(originalName, ext);
  
  let result = template
    .replace('{name}', baseName)
    .replace('{index}', String(index).padStart(String(total).length, '0'))
    .replace('{date}', date)
    .replace('{ext}', ext);
  
  if (!result.includes(ext)) {
    result += ext;
  }
  
  return result;
}

function validateRenameTemplate(template) {
  const validVariables = ['{name}', '{index}', '{date}', '{ext}'];
  const variableRegex = /\{[^{}]+\}/g;
  const matches = template.match(variableRegex) || [];
  
  const invalidVariables = matches.filter(v => !validVariables.includes(v));
  return {
    valid: invalidVariables.length === 0,
    invalidVariables
  };
}

ipcMain.handle('select-files', async () => {
  const result = await dialog.showOpenDialog(mainWindow, {
    properties: ['openFile', 'multiSelections'],
    filters: [
      { name: 'Images', extensions: ['jpg', 'jpeg', 'png', 'webp', 'bmp'] }
    ]
  });

  if (result.canceled) return { files: [] };

  const allFiles = [];
  result.filePaths.forEach(filePath => {
    getFilesFromPath(filePath, allFiles);
  });

  return { files: allFiles };
});

ipcMain.handle('select-folder', async () => {
  const result = await dialog.showOpenDialog(mainWindow, {
    properties: ['openDirectory']
  });

  if (result.canceled) return { files: [] };

  const allFiles = [];
  result.filePaths.forEach(filePath => {
    getFilesFromPath(filePath, allFiles);
  });

  return { files: allFiles };
});

ipcMain.handle('get-thumbnail', async (event, filePath) => {
  try {
    const thumbnail = await sharp(filePath)
      .resize(150, 150, { fit: 'inside', withoutEnlargement: true })
      .toBuffer();
    
    return {
      success: true,
      data: `data:image/png;base64,${thumbnail.toString('base64')}`,
      name: path.basename(filePath),
      path: filePath
    };
  } catch (error) {
    return {
      success: false,
      error: error.message,
      path: filePath
    };
  }
});

ipcMain.handle('validate-rename-template', (event, template) => {
  return validateRenameTemplate(template);
});

ipcMain.handle('process-images', async (event, options) => {
  const { images, operations, outputDir } = options;
  const date = new Date().toISOString().slice(0, 10).replace(/-/g, '');
  const results = {
    success: [],
    failed: [],
    skipped: [],
    total: images.length
  };
  
  const processedFiles = [];
  
  for (let i = 0; i < images.length; i++) {
    const imagePath = images[i];
    const fileName = path.basename(imagePath);
    const fileDir = path.dirname(imagePath);
    
    const targetOutputDir = outputDir || path.join(fileDir, 'processed');
    
    if (!fs.existsSync(targetOutputDir)) {
      fs.mkdirSync(targetOutputDir, { recursive: true });
    }
    
    let outputExt = path.extname(fileName).toLowerCase();
    if (operations.format && operations.format !== 'original') {
      outputExt = `.${operations.format}`;
    }
    
    let outputName = fileName;
    if (operations.rename && operations.rename.template) {
      const validation = validateRenameTemplate(operations.rename.template);
      if (validation.valid) {
        outputName = applyRenameTemplate(
          operations.rename.template, 
          fileName, 
          i + 1, 
          images.length, 
          date
        );
      }
    }
    
    const outputPath = generateUniqueOutputPath(targetOutputDir, outputName, outputExt);
    
    try {
      const image = sharp(imagePath);
      const metadata = await image.metadata();
      
      let width = metadata.width;
      let height = metadata.height;
      
      if (operations.resize) {
        const { mode, width: targetWidth, height: targetHeight, percent, lockAspect } = operations.resize;
        
        if (mode === 'percent' && percent) {
          const scale = percent / 100;
          width = Math.round(metadata.width * scale);
          height = Math.round(metadata.height * scale);
        } else if (mode === 'pixels') {
          if (lockAspect) {
            if (targetWidth) {
              const aspectRatio = metadata.height / metadata.width;
              width = targetWidth;
              height = Math.round(targetWidth * aspectRatio);
            } else if (targetHeight) {
              const aspectRatio = metadata.width / metadata.height;
              height = targetHeight;
              width = Math.round(targetHeight * aspectRatio);
            }
          } else {
            width = targetWidth || metadata.width;
            height = targetHeight || metadata.height;
          }
        }
        
        image.resize({
          width,
          height,
          fit: 'cover',
          withoutEnlargement: true
        });
      }
      
      const formatMap = {
        'jpg': 'jpeg',
        'jpeg': 'jpeg',
        'png': 'png',
        'webp': 'webp'
      };
      
      const outputFormat = formatMap[outputExt.slice(1)] || 'jpeg';
      image.toFormat(outputFormat);
      
      await image.toFile(outputPath);
      
      results.success.push({
        original: imagePath,
        processed: outputPath,
        name: path.basename(outputPath)
      });
      
      processedFiles.push({
        original: imagePath,
        processed: outputPath,
        backup: null
      });
      
    } catch (error) {
      if (error.message.includes('Input file is missing') || 
          error.message.includes('unsupported') ||
          error.message.includes('invalid')) {
        results.skipped.push({
          path: imagePath,
          name: fileName,
          error: error.message
        });
      } else {
        results.failed.push({
          path: imagePath,
          name: fileName,
          error: error.message
        });
      }
    }
    
    event.sender.send('processing-progress', {
      current: i + 1,
      total: images.length,
      percentage: Math.round(((i + 1) / images.length) * 100)
    });
  }
  
  return { results, processedFiles };
});

ipcMain.handle('get-comparison-images', async (event, originalPath, processedPath) => {
  try {
    const [originalThumb, processedThumb] = await Promise.all([
      sharp(originalPath)
        .resize(400, 400, { fit: 'inside', withoutEnlargement: true })
        .toBuffer(),
      sharp(processedPath)
        .resize(400, 400, { fit: 'inside', withoutEnlargement: true })
        .toBuffer()
    ]);
    
    return {
      success: true,
      original: `data:image/png;base64,${originalThumb.toString('base64')}`,
      processed: `data:image/png;base64,${processedThumb.toString('base64')}`
    };
  } catch (error) {
    return {
      success: false,
      error: error.message
    };
  }
});

ipcMain.handle('open-output-folder', async (event, folderPath) => {
  try {
    await shell.openPath(folderPath);
    return { success: true };
  } catch (error) {
    return { success: false, error: error.message };
  }
});
