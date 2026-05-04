const fs = require('fs');
const path = require('path');

const DEFAULT_CATEGORIES = [
  { id: 'cat_1', name: 'JavaScript', isDefault: true },
  { id: 'cat_2', name: 'Python', isDefault: true },
  { id: 'cat_3', name: 'Go', isDefault: true },
  { id: 'cat_4', name: 'Java', isDefault: true },
  { id: 'cat_5', name: 'SQL', isDefault: true },
  { id: 'cat_6', name: 'HTML/CSS', isDefault: true }
];

class StoreService {
  constructor(appDataPath) {
    this.dataDir = path.join(appDataPath, 'snippets-manager');
    this.dataFile = path.join(this.dataDir, 'data.json');
    this.ensureDataDirectory();
  }

  ensureDataDirectory() {
    if (!fs.existsSync(this.dataDir)) {
      fs.mkdirSync(this.dataDir, { recursive: true });
    }
    
    if (!fs.existsSync(this.dataFile)) {
      const initialData = {
        categories: [...DEFAULT_CATEGORIES],
        snippets: []
      };
      this.saveData(initialData);
    }
  }

  readData() {
    try {
      const rawData = fs.readFileSync(this.dataFile, 'utf-8');
      return JSON.parse(rawData);
    } catch (error) {
      console.error('读取数据失败:', error);
      return { categories: [...DEFAULT_CATEGORIES], snippets: [] };
    }
  }

  saveData(data) {
    try {
      const jsonData = JSON.stringify(data, null, 2);
      fs.writeFileSync(this.dataFile, jsonData, 'utf-8');
      return true;
    } catch (error) {
      console.error('保存数据失败:', error);
      return false;
    }
  }

  getSnippets() {
    const data = this.readData();
    return data.snippets || [];
  }

  saveSnippets(snippets) {
    const data = this.readData();
    data.snippets = snippets;
    return this.saveData(data);
  }

  getCategories() {
    const data = this.readData();
    return data.categories || [];
  }

  saveCategories(categories) {
    const data = this.readData();
    data.categories = categories;
    return this.saveData(data);
  }

  exportAll() {
    return this.readData();
  }

  importData(data, options = { skipDuplicates: true, overwriteDuplicates: false }) {
    const currentData = this.readData();
    const { categories: importCategories, snippets: importSnippets } = data;
    
    if (importCategories) {
      const existingIds = new Set(currentData.categories.map(c => c.id));
      importCategories.forEach(cat => {
        if (!existingIds.has(cat.id) && !currentData.categories.find(c => c.name === cat.name)) {
          currentData.categories.push(cat);
        }
      });
    }
    
    if (importSnippets) {
      const existingTitles = new Map();
      currentData.snippets.forEach((s, idx) => {
        existingTitles.set(s.title, idx);
      });
      
      importSnippets.forEach(snippet => {
        if (existingTitles.has(snippet.title)) {
          if (options.overwriteDuplicates) {
            const idx = existingTitles.get(snippet.title);
            currentData.snippets[idx] = snippet;
          }
        } else {
          currentData.snippets.push(snippet);
        }
      });
    }
    
    return this.saveData(currentData);
  }
}

module.exports = StoreService;
