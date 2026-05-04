class SnippetService {
  constructor(store, categoryService) {
    this.store = store;
    this.categoryService = categoryService;
  }

  getAll() {
    return this.store.getSnippets();
  }

  getByCategory(categoryId) {
    const snippets = this.store.getSnippets();
    return snippets.filter(s => s.categoryId === categoryId);
  }

  getById(id) {
    const snippets = this.store.getSnippets();
    return snippets.find(s => s.id === id) || null;
  }

  create(data) {
    if (!data.title || data.title.trim() === '') {
      throw new Error('标题不能为空');
    }
    
    if (!data.categoryId) {
      throw new Error('必须选择分类');
    }
    
    const category = this.categoryService.getById(data.categoryId);
    if (!category) {
      throw new Error('分类不存在');
    }
    
    const snippets = this.store.getSnippets();
    const titleExists = snippets.some(s => s.title === data.title.trim());
    
    if (titleExists) {
      throw new Error('标题已存在');
    }
    
    const newSnippet = {
      id: `snippet_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      title: data.title.trim(),
      content: data.content || '',
      language: data.language || 'text',
      notes: data.notes || '',
      tags: data.tags || [],
      categoryId: data.categoryId,
      isFavorite: false,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    };
    
    snippets.push(newSnippet);
    this.store.saveSnippets(snippets);
    
    return newSnippet;
  }

  update(id, data) {
    const snippets = this.store.getSnippets();
    const index = snippets.findIndex(s => s.id === id);
    
    if (index === -1) {
      throw new Error('代码片段不存在');
    }
    
    const existingSnippet = snippets[index];
    
    if (data.title && data.title.trim() !== existingSnippet.title) {
      const titleExists = snippets.some((s, idx) => idx !== index && s.title === data.title.trim());
      if (titleExists) {
        throw new Error('标题已存在');
      }
      existingSnippet.title = data.title.trim();
    }
    
    if (data.categoryId && data.categoryId !== existingSnippet.categoryId) {
      const category = this.categoryService.getById(data.categoryId);
      if (!category) {
        throw new Error('分类不存在');
      }
      existingSnippet.categoryId = data.categoryId;
    }
    
    if (data.content !== undefined) {
      existingSnippet.content = data.content;
    }
    
    if (data.language !== undefined) {
      existingSnippet.language = data.language || 'text';
    }
    
    if (data.notes !== undefined) {
      existingSnippet.notes = data.notes || '';
    }
    
    if (data.tags !== undefined) {
      existingSnippet.tags = data.tags || [];
    }
    
    existingSnippet.updatedAt = new Date().toISOString();
    
    this.store.saveSnippets(snippets);
    
    return existingSnippet;
  }

  delete(id) {
    const snippets = this.store.getSnippets();
    const index = snippets.findIndex(s => s.id === id);
    
    if (index === -1) {
      throw new Error('代码片段不存在');
    }
    
    snippets.splice(index, 1);
    this.store.saveSnippets(snippets);
    
    return true;
  }

  toggleFavorite(id) {
    const snippets = this.store.getSnippets();
    const index = snippets.findIndex(s => s.id === id);
    
    if (index === -1) {
      throw new Error('代码片段不存在');
    }
    
    snippets[index].isFavorite = !snippets[index].isFavorite;
    snippets[index].updatedAt = new Date().toISOString();
    
    this.store.saveSnippets(snippets);
    
    return snippets[index];
  }

  search(keyword) {
    if (!keyword || keyword.trim() === '') {
      return this.getAll();
    }
    
    const kw = keyword.toLowerCase().trim();
    const snippets = this.store.getSnippets();
    
    return snippets.filter(snippet => {
      const titleMatch = snippet.title.toLowerCase().includes(kw);
      const contentMatch = snippet.content.toLowerCase().includes(kw);
      const notesMatch = snippet.notes.toLowerCase().includes(kw);
      const tagsMatch = snippet.tags.some(tag => tag.toLowerCase().includes(kw));
      
      return titleMatch || contentMatch || notesMatch || tagsMatch;
    });
  }

  getFavorites() {
    const snippets = this.store.getSnippets();
    return snippets.filter(s => s.isFavorite);
  }

  getSortedForDisplay(snippets, sortBy = 'favorite') {
    const sorted = [...snippets];
    
    sorted.sort((a, b) => {
      if (sortBy === 'favorite') {
        if (a.isFavorite !== b.isFavorite) {
          return a.isFavorite ? -1 : 1;
        }
        return new Date(b.updatedAt) - new Date(a.updatedAt);
      }
      
      if (sortBy === 'title') {
        return a.title.localeCompare(b.title);
      }
      
      if (sortBy === 'date') {
        return new Date(b.createdAt) - new Date(a.createdAt);
      }
      
      return 0;
    });
    
    return sorted;
  }
}

module.exports = SnippetService;
