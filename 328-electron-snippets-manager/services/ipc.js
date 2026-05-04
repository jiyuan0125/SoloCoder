const fs = require('fs');

function setupIPC(ipcMain, dialog, snippetService, categoryService, store) {
  
  ipcMain.handle('categories:getAll', async () => {
    try {
      const categories = categoryService.getAll();
      return { success: true, data: categories };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('categories:add', async (event, name) => {
    try {
      const category = categoryService.add(name);
      return { success: true, data: category };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('categories:update', async (event, id, name) => {
    try {
      const category = categoryService.update(id, name);
      return { success: true, data: category };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('categories:delete', async (event, id) => {
    try {
      categoryService.delete(id);
      return { success: true };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('snippets:getAll', async () => {
    try {
      const snippets = snippetService.getAll();
      return { success: true, data: snippets };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('snippets:getByCategory', async (event, categoryId) => {
    try {
      const snippets = snippetService.getByCategory(categoryId);
      return { success: true, data: snippets };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('snippets:getById', async (event, id) => {
    try {
      const snippet = snippetService.getById(id);
      return { success: true, data: snippet };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('snippets:create', async (event, data) => {
    try {
      const snippet = snippetService.create(data);
      return { success: true, data: snippet };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('snippets:update', async (event, id, data) => {
    try {
      const snippet = snippetService.update(id, data);
      return { success: true, data: snippet };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('snippets:delete', async (event, id) => {
    try {
      snippetService.delete(id);
      return { success: true };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('snippets:toggleFavorite', async (event, id) => {
    try {
      const snippet = snippetService.toggleFavorite(id);
      return { success: true, data: snippet };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('snippets:search', async (event, keyword) => {
    try {
      const snippets = snippetService.search(keyword);
      return { success: true, data: snippets };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('snippets:getFavorites', async () => {
    try {
      const snippets = snippetService.getFavorites();
      return { success: true, data: snippets };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('dialog:showSaveDialog', async (event, options) => {
    try {
      const result = await dialog.showSaveDialog({
        title: options.title || '保存文件',
        defaultPath: options.defaultPath || 'snippets-export.json',
        filters: [
          { name: 'JSON 文件', extensions: ['json'] },
          { name: '所有文件', extensions: ['*'] }
        ]
      });
      return { success: true, data: result };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('dialog:showOpenDialog', async (event, options) => {
    try {
      const result = await dialog.showOpenDialog({
        title: options.title || '选择文件',
        properties: ['openFile'],
        filters: [
          { name: 'JSON 文件', extensions: ['json'] },
          { name: '所有文件', extensions: ['*'] }
        ]
      });
      return { success: true, data: result };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('export:all', async (event, filePath) => {
    try {
      const data = store.exportAll();
      const jsonData = JSON.stringify(data, null, 2);
      fs.writeFileSync(filePath, jsonData, 'utf-8');
      return { success: true };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('export:category', async (event, categoryId, filePath) => {
    try {
      const category = categoryService.getById(categoryId);
      if (!category) {
        return { success: false, error: '分类不存在' };
      }
      
      const snippets = snippetService.getByCategory(categoryId);
      const exportData = {
        categories: [category],
        snippets: snippets
      };
      
      const jsonData = JSON.stringify(exportData, null, 2);
      fs.writeFileSync(filePath, jsonData, 'utf-8');
      return { success: true };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('import:fromFile', async (event, filePath, options) => {
    try {
      const fileContent = fs.readFileSync(filePath, 'utf-8');
      const importData = JSON.parse(fileContent);
      
      const result = store.importData(importData, options);
      
      if (result) {
        return { success: true };
      } else {
        return { success: false, error: '导入失败' };
      }
    } catch (error) {
      return { success: false, error: error.message };
    }
  });
}

module.exports = { setupIPC };
