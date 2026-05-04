class CategoryService {
  constructor(store) {
    this.store = store;
  }

  getAll() {
    return this.store.getCategories();
  }

  add(name) {
    if (!name || name.trim() === '') {
      throw new Error('分类名称不能为空');
    }
    
    const categories = this.store.getCategories();
    const exists = categories.some(cat => cat.name.toLowerCase() === name.toLowerCase().trim());
    
    if (exists) {
      throw new Error('分类已存在');
    }
    
    const newCategory = {
      id: `cat_${Date.now()}`,
      name: name.trim(),
      isDefault: false
    };
    
    categories.push(newCategory);
    this.store.saveCategories(categories);
    
    return newCategory;
  }

  update(id, name) {
    if (!name || name.trim() === '') {
      throw new Error('分类名称不能为空');
    }
    
    const categories = this.store.getCategories();
    const index = categories.findIndex(cat => cat.id === id);
    
    if (index === -1) {
      throw new Error('分类不存在');
    }
    
    const nameExists = categories.some(
      (cat, idx) => idx !== index && cat.name.toLowerCase() === name.toLowerCase().trim()
    );
    
    if (nameExists) {
      throw new Error('分类名称已存在');
    }
    
    categories[index].name = name.trim();
    this.store.saveCategories(categories);
    
    return categories[index];
  }

  delete(id) {
    const categories = this.store.getCategories();
    const index = categories.findIndex(cat => cat.id === id);
    
    if (index === -1) {
      throw new Error('分类不存在');
    }
    
    if (categories[index].isDefault) {
      throw new Error('默认分类不能删除');
    }
    
    const snippets = this.store.getSnippets();
    const hasSnippets = snippets.some(s => s.categoryId === id);
    
    if (hasSnippets) {
      throw new Error('该分类下还有代码片段，请先移动或删除这些片段');
    }
    
    categories.splice(index, 1);
    this.store.saveCategories(categories);
    
    return true;
  }

  getById(id) {
    const categories = this.store.getCategories();
    return categories.find(cat => cat.id === id) || null;
  }
}

module.exports = CategoryService;
