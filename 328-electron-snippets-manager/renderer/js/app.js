const App = {
  categories: [],
  snippets: [],
  currentCategory: 'all',
  currentSnippet: null,
  isEditing: false,
  searchKeyword: '',
  sortBy: 'favorite',
  editingCategoryId: null,

  async init() {
    await this.loadCategories();
    await this.loadSnippets();
    this.bindEvents();
    this.updateCounts();
    this.renderCategoryList();
    this.renderSnippets();
  },

  async loadCategories() {
    const result = await window.electronAPI.categories.getAll();
    if (result.success) {
      this.categories = result.data;
    }
  },

  async loadSnippets() {
    const result = await window.electronAPI.snippets.getAll();
    if (result.success) {
      this.snippets = result.data;
    }
  },

  getCategoryById(id) {
    return this.categories.find(c => c.id === id) || { name: '未知分类' };
  },

  getFilteredSnippets() {
    let filtered = [...this.snippets];

    if (this.currentCategory === 'all') {
      filtered = [...this.snippets];
    } else if (this.currentCategory === 'favorites') {
      filtered = this.snippets.filter(s => s.isFavorite);
    } else {
      filtered = this.snippets.filter(s => s.categoryId === this.currentCategory);
    }

    if (this.searchKeyword && this.searchKeyword.trim()) {
      const kw = this.searchKeyword.toLowerCase().trim();
      filtered = filtered.filter(snippet => {
        const titleMatch = snippet.title.toLowerCase().includes(kw);
        const contentMatch = snippet.content.toLowerCase().includes(kw);
        const notesMatch = snippet.notes.toLowerCase().includes(kw);
        const tagsMatch = snippet.tags.some(tag => tag.toLowerCase().includes(kw));
        return titleMatch || contentMatch || notesMatch || tagsMatch;
      });
    }

    if (this.sortBy === 'favorite') {
      filtered.sort((a, b) => {
        if (a.isFavorite !== b.isFavorite) {
          return a.isFavorite ? -1 : 1;
        }
        return new Date(b.updatedAt) - new Date(a.updatedAt);
      });
    } else if (this.sortBy === 'title') {
      filtered.sort((a, b) => a.title.localeCompare(b.title));
    } else if (this.sortBy === 'date') {
      filtered.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));
    }

    return filtered;
  },

  updateCounts() {
    const allCount = this.snippets.length;
    const favCount = this.snippets.filter(s => s.isFavorite).length;
    
    const allCountEl = document.getElementById('allCount');
    const favCountEl = document.getElementById('favoritesCount');
    
    if (allCountEl) allCountEl.textContent = allCount;
    if (favCountEl) favCountEl.textContent = favCount;

    this.categories.forEach(cat => {
      const count = this.snippets.filter(s => s.categoryId === cat.id).length;
      cat.count = count;
    });
  },

  renderCategoryList() {
    const categoryList = document.getElementById('categoryList');
    const divider = categoryList.querySelector('.category-divider');
    
    const existingCategories = categoryList.querySelectorAll('.category-item[data-id][data-id]:not([data-id="all"]):not([data-id="favorites"])');
    existingCategories.forEach(el => el.remove());

    const allItem = categoryList.querySelector('[data-id="all"]');
    const favItem = categoryList.querySelector('[data-id="favorites"]');

    if (this.currentCategory === 'all') {
      allItem.classList.add('active');
      favItem.classList.remove('active');
    } else if (this.currentCategory === 'favorites') {
      allItem.classList.remove('active');
      favItem.classList.add('active');
    } else {
      allItem.classList.remove('active');
      favItem.classList.remove('active');
    }

    this.categories.forEach(cat => {
      const categoryItem = document.createElement('div');
      categoryItem.className = `category-item ${this.currentCategory === cat.id ? 'active' : ''}`;
      categoryItem.dataset.id = cat.id;
      categoryItem.innerHTML = `
        <span class="category-icon">📂</span>
        <span class="category-name">${this.escapeHtml(cat.name)}</span>
        <span class="category-count">${cat.count || 0}</span>
      `;
      
      categoryItem.addEventListener('click', (e) => {
        if (e.target.classList.contains('category-edit') || e.target.classList.contains('category-delete')) {
          return;
        }
        this.selectCategory(cat.id);
      });

      categoryItem.addEventListener('contextmenu', (e) => {
        if (cat.isDefault) return;
        e.preventDefault();
        this.showCategoryContextMenu(e, cat);
      });

      categoryList.insertBefore(categoryItem, divider.nextSibling);
    });
  },

  selectCategory(categoryId) {
    this.currentCategory = categoryId;
    this.currentSnippet = null;
    this.isEditing = false;
    this.renderCategoryList();
    this.renderSnippets();
    this.hideEditor();
  },

  renderSnippets() {
    const snippetsList = document.getElementById('snippetsList');
    const filtered = this.getFilteredSnippets();

    if (filtered.length === 0) {
      snippetsList.innerHTML = `
        <div class="empty-state">
          <p>暂无代码片段</p>
          <p class="empty-hint">点击"新建片段"开始添加</p>
        </div>
      `;
      return;
    }

    snippetsList.innerHTML = '';

    filtered.forEach(snippet => {
      const category = this.getCategoryById(snippet.categoryId);
      const snippetItem = document.createElement('div');
      snippetItem.className = `snippet-item ${this.currentSnippet?.id === snippet.id ? 'active' : ''}`;
      snippetItem.dataset.id = snippet.id;

      const previewContent = snippet.content.substring(0, 80).replace(/\n/g, ' ');
      
      let tagsHtml = '';
      if (snippet.tags && snippet.tags.length > 0) {
        tagsHtml = snippet.tags.slice(0, 3).map(tag => 
          `<span class="snippet-tag">${this.escapeHtml(tag)}</span>`
        ).join('');
        if (snippet.tags.length > 3) {
          tagsHtml += `<span class="snippet-tag">+${snippet.tags.length - 3}</span>`;
        }
      }

      snippetItem.innerHTML = `
        <div class="snippet-item-header">
          <span class="snippet-favorite ${snippet.isFavorite ? 'is-favorite' : ''}" data-id="${snippet.id}">${snippet.isFavorite ? '⭐' : '☆'}</span>
          <span class="snippet-title">${this.escapeHtml(snippet.title)}</span>
        </div>
        <div class="snippet-item-meta">
          <span class="meta-tag language">${this.escapeHtml(snippet.language)}</span>
          <span class="meta-tag">${this.escapeHtml(category.name)}</span>
        </div>
        ${snippetItem.previewContent ? `<div class="snippet-item-preview">${this.escapeHtml(previewContent)}...</div>` : ''}
        ${tagsHtml ? `<div class="snippet-tags">${tagsHtml}</div>` : ''}
      `;

      snippetItem.addEventListener('click', (e) => {
        if (e.target.classList.contains('snippet-favorite')) {
          return;
        }
        this.selectSnippet(snippet);
      });

      const favBtn = snippetItem.querySelector('.snippet-favorite');
      favBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        this.toggleFavorite(snippet.id);
      });

      snippetsList.appendChild(snippetItem);
    });
  },

  selectSnippet(snippet) {
    this.currentSnippet = snippet;
    this.isEditing = false;
    this.renderSnippets();
    this.showSnippetView(snippet);
  },

  showSnippetView(snippet) {
    const category = this.getCategoryById(snippet.categoryId);
    
    document.getElementById('editorEmpty').style.display = 'none';
    document.getElementById('editorForm').style.display = 'none';
    document.getElementById('editorView').style.display = 'block';
    
    document.getElementById('saveBtn').style.display = 'none';
    document.getElementById('cancelBtn').style.display = 'none';
    document.getElementById('copyBtn').style.display = 'inline-block';
    document.getElementById('editorTitle').textContent = '片段详情';

    const favEl = document.getElementById('viewFavorite');
    favEl.innerHTML = snippet.isFavorite ? '⭐' : '☆';
    favEl.className = `view-favorite ${snippet.isFavorite ? 'is-favorite' : ''}`;
    favEl.onclick = () => this.toggleFavorite(snippet.id);

    document.getElementById('viewTitle').textContent = snippet.title;
    document.getElementById('viewCategory').textContent = category.name;
    document.getElementById('viewLanguage').textContent = snippet.language;

    const tagsContainer = document.getElementById('viewTags');
    if (snippet.tags && snippet.tags.length > 0) {
      tagsContainer.innerHTML = snippet.tags.map(tag => 
        `<span class="tag-badge">${this.escapeHtml(tag)}</span>`
      ).join('');
      tagsContainer.style.display = 'flex';
    } else {
      tagsContainer.style.display = 'none';
    }

    const codeEl = document.getElementById('viewCode');
    codeEl.textContent = snippet.content || '';
    
    if (typeof hljs !== 'undefined') {
      try {
        hljs.highlightElement(codeEl);
      } catch (e) {
        console.log('Highlight error:', e);
      }
    }

    const notesSection = document.getElementById('viewNotesSection');
    const notesEl = document.getElementById('viewNotes');
    if (snippet.notes && snippet.notes.trim()) {
      notesEl.textContent = snippet.notes;
      notesSection.style.display = 'block';
    } else {
      notesSection.style.display = 'none';
    }

    document.getElementById('viewCreated').textContent = this.formatDate(snippet.createdAt);
    document.getElementById('viewUpdated').textContent = this.formatDate(snippet.updatedAt);
  },

  hideEditor() {
    document.getElementById('editorEmpty').style.display = 'flex';
    document.getElementById('editorForm').style.display = 'none';
    document.getElementById('editorView').style.display = 'none';
    document.getElementById('editorTitle').textContent = '片段详情';
    document.getElementById('saveBtn').style.display = 'none';
    document.getElementById('cancelBtn').style.display = 'none';
    document.getElementById('copyBtn').style.display = 'none';
  },

  async toggleFavorite(id) {
    const result = await window.electronAPI.snippets.toggleFavorite(id);
    if (result.success) {
      const idx = this.snippets.findIndex(s => s.id === id);
      if (idx !== -1) {
        this.snippets[idx].isFavorite = result.data.isFavorite;
      }
      
      if (this.currentSnippet && this.currentSnippet.id === id) {
        this.currentSnippet.isFavorite = result.data.isFavorite;
        this.showSnippetView(this.currentSnippet);
      }
      
      this.updateCounts();
      this.renderSnippets();
      this.renderCategoryList();
    } else {
      this.showToast(result.error, 'error');
    }
  },

  async createNewSnippet() {
    this.currentSnippet = null;
    this.isEditing = true;

    document.getElementById('editorEmpty').style.display = 'none';
    document.getElementById('editorView').style.display = 'none';
    document.getElementById('editorForm').style.display = 'block';
    
    document.getElementById('editorTitle').textContent = '新建片段';
    document.getElementById('saveBtn').style.display = 'inline-block';
    document.getElementById('cancelBtn').style.display = 'inline-block';
    document.getElementById('copyBtn').style.display = 'none';
    document.getElementById('formMeta').style.display = 'none';

    document.getElementById('formTitle').value = '';
    document.getElementById('formCode').value = '';
    document.getElementById('formNotes').value = '';
    document.getElementById('formTags').value = '';
    document.getElementById('tagsPreview').innerHTML = '';

    this.populateCategorySelect();
    
    if (this.currentCategory !== 'all' && this.currentCategory !== 'favorites') {
      document.getElementById('formCategory').value = this.currentCategory;
    }
    
    if (this.categories.length > 0) {
      const firstCatId = document.getElementById('formCategory').value;
      if (firstCatId) {
        const lang = this.guessLanguageByCategory(firstCatId);
        document.getElementById('formLanguage').value = lang;
      }
    }

    document.getElementById('formTitle').focus();
  },

  populateCategorySelect() {
    const select = document.getElementById('formCategory');
    select.innerHTML = this.categories.map(cat => 
      `<option value="${cat.id}">${this.escapeHtml(cat.name)}</option>`
    ).join('');
  },

  guessLanguageByCategory(categoryId) {
    const category = this.getCategoryById(categoryId);
    const name = category.name.toLowerCase();
    
    const map = {
      'javascript': 'javascript',
      'js': 'javascript',
      'python': 'python',
      'go': 'go',
      'java': 'java',
      'sql': 'sql',
      'html': 'html',
      'css': 'css',
      'typescript': 'typescript',
      'rust': 'rust',
      'php': 'php',
      'c#': 'csharp',
      'c++': 'cpp',
      'ruby': 'ruby',
      'swift': 'swift',
      'kotlin': 'kotlin',
      'shell': 'shell',
      'yaml': 'yaml',
      'xml': 'xml',
      'markdown': 'markdown',
      'json': 'json'
    };

    for (const [key, value] of Object.entries(map)) {
      if (name.includes(key)) {
        return value;
      }
    }

    return 'text';
  },

  async editSnippet() {
    if (!this.currentSnippet) return;
    
    this.isEditing = true;
    
    document.getElementById('editorView').style.display = 'none';
    document.getElementById('editorForm').style.display = 'block';
    
    document.getElementById('editorTitle').textContent = '编辑片段';
    document.getElementById('saveBtn').style.display = 'inline-block';
    document.getElementById('cancelBtn').style.display = 'inline-block';
    document.getElementById('copyBtn').style.display = 'none';

    this.populateCategorySelect();

    document.getElementById('formTitle').value = this.currentSnippet.title;
    document.getElementById('formCategory').value = this.currentSnippet.categoryId;
    document.getElementById('formLanguage').value = this.currentSnippet.language;
    document.getElementById('formCode').value = this.currentSnippet.content;
    document.getElementById('formNotes').value = this.currentSnippet.notes;
    document.getElementById('formTags').value = this.currentSnippet.tags.join(' ');

    this.updateTagsPreview();

    document.getElementById('formMeta').style.display = 'block';
    document.getElementById('metaCreated').textContent = this.formatDate(this.currentSnippet.createdAt);
    document.getElementById('metaUpdated').textContent = this.formatDate(this.currentSnippet.updatedAt);
  },

  cancelEdit() {
    if (this.currentSnippet) {
      this.isEditing = false;
      this.showSnippetView(this.currentSnippet);
    } else {
      this.isEditing = false;
      this.hideEditor();
    }
  },

  async saveSnippet() {
    const title = document.getElementById('formTitle').value.trim();
    const categoryId = document.getElementById('formCategory').value;
    const language = document.getElementById('formLanguage').value;
    const content = document.getElementById('formCode').value;
    const notes = document.getElementById('formNotes').value;
    const tagsInput = document.getElementById('formTags').value;
    const tags = this.parseTags(tagsInput);

    if (!title) {
      this.showToast('标题不能为空', 'error');
      return;
    }

    if (!categoryId) {
      this.showToast('请选择分类', 'error');
      return;
    }

    const data = {
      title,
      categoryId,
      language,
      content,
      notes,
      tags
    };

    let result;
    if (this.currentSnippet) {
      result = await window.electronAPI.snippets.update(this.currentSnippet.id, data);
    } else {
      result = await window.electronAPI.snippets.create(data);
    }

    if (result.success) {
      if (this.currentSnippet) {
        const idx = this.snippets.findIndex(s => s.id === this.currentSnippet.id);
        if (idx !== -1) {
          this.snippets[idx] = result.data;
        }
        this.currentSnippet = result.data;
        this.isEditing = false;
        this.showSnippetView(result.data);
      } else {
        this.snippets.push(result.data);
        this.currentSnippet = result.data;
        this.isEditing = false;
        this.showSnippetView(result.data);
      }

      this.updateCounts();
      this.renderCategoryList();
      this.renderSnippets();
      this.showToast('保存成功', 'success');
    } else {
      this.showToast(result.error || '保存失败', 'error');
    }
  },

  async deleteSnippet() {
    if (!this.currentSnippet) return;

    const confirmed = await this.showConfirm('确认删除', `确定要删除代码片段 "${this.currentSnippet.title}" 吗？此操作不可恢复。`);
    
    if (!confirmed) return;

    const result = await window.electronAPI.snippets.delete(this.currentSnippet.id);
    
    if (result.success) {
      const idx = this.snippets.findIndex(s => s.id === this.currentSnippet.id);
      if (idx !== -1) {
        this.snippets.splice(idx, 1);
      }
      
      this.currentSnippet = null;
      this.isEditing = false;
      this.hideEditor();
      this.updateCounts();
      this.renderCategoryList();
      this.renderSnippets();
      this.showToast('删除成功', 'success');
    } else {
      this.showToast(result.error || '删除失败', 'error');
    }
  },

  async copyCode() {
    if (!this.currentSnippet) return;

    try {
      await window.electronAPI.clipboard.writeText(this.currentSnippet.content);
      this.showToast('已复制', 'success');
    } catch (e) {
      try {
        await navigator.clipboard.writeText(this.currentSnippet.content);
        this.showToast('已复制', 'success');
      } catch (err) {
        this.showToast('复制失败', 'error');
      }
    }
  },

  async addCategory() {
    const name = await this.showPrompt('添加分类', '请输入分类名称：');
    
    if (!name || !name.trim()) return;

    const result = await window.electronAPI.categories.add(name.trim());
    
    if (result.success) {
      this.categories.push(result.data);
      this.renderCategoryList();
      this.showToast('分类添加成功', 'success');
    } else {
      this.showToast(result.error || '添加失败', 'error');
    }
  },

  async renameCategory(category) {
    const name = await this.showPrompt('重命名分类', '请输入新名称：', category.name);
    
    if (!name || !name.trim() || name.trim() === category.name) return;

    const result = await window.electronAPI.categories.update(category.id, name.trim());
    
    if (result.success) {
      const idx = this.categories.findIndex(c => c.id === category.id);
      if (idx !== -1) {
        this.categories[idx].name = result.data.name;
      }
      this.renderCategoryList();
      this.showToast('分类已重命名', 'success');
    } else {
      this.showToast(result.error || '重命名失败', 'error');
    }
  },

  async deleteCategory(category) {
    const confirmed = await this.showConfirm('删除分类', `确定要删除分类 "${category.name}" 吗？只有空分类才能删除。`);
    
    if (!confirmed) return;

    const result = await window.electronAPI.categories.delete(category.id);
    
    if (result.success) {
      const idx = this.categories.findIndex(c => c.id === category.id);
      if (idx !== -1) {
        this.categories.splice(idx, 1);
      }
      if (this.currentCategory === category.id) {
        this.selectCategory('all');
      }
      this.renderCategoryList();
      this.showToast('分类已删除', 'success');
    } else {
      this.showToast(result.error || '删除失败', 'error');
    }
  },

  showCategoryContextMenu(e, category) {
    const existingMenu = document.querySelector('.context-menu');
    if (existingMenu) existingMenu.remove();

    const menu = document.createElement('div');
    menu.className = 'context-menu';
    menu.innerHTML = `
      <div class="context-menu-item" data-action="rename">重命名</div>
      <div class="context-menu-divider"></div>
      <div class="context-menu-item danger" data-action="delete">删除</div>
    `;

    menu.style.left = e.pageX + 'px';
    menu.style.top = e.pageY + 'px';
    document.body.appendChild(menu);

    menu.addEventListener('click', (ev) => {
      const action = ev.target.dataset.action;
      if (action === 'rename') {
        this.renameCategory(category);
      } else if (action === 'delete') {
        this.deleteCategory(category);
      }
      menu.remove();
    });

    const closeMenu = (ev) => {
      if (!menu.contains(ev.target)) {
        menu.remove();
        document.removeEventListener('click', closeMenu);
      }
    };
    setTimeout(() => document.addEventListener('click', closeMenu), 0);
  },

  async exportAll() {
    const result = await window.electronAPI.dialog.showSaveDialog({
      title: '导出全部代码片段',
      defaultPath: 'snippets-all.json'
    });

    if (result.data.canceled) return;

    const filePath = result.data.filePath;
    const exportResult = await window.electronAPI.export.all(filePath);

    if (exportResult.success) {
      this.showToast('导出成功', 'success');
    } else {
      this.showToast(exportResult.error || '导出失败', 'error');
    }
  },

  async exportCurrentCategory() {
    if (this.currentCategory === 'all') {
      this.exportAll();
      return;
    }
    
    if (this.currentCategory === 'favorites') {
      this.showToast('收藏夹不能单独导出，请使用"导出全部"', 'warning');
      return;
    }

    const category = this.getCategoryById(this.currentCategory);
    
    const result = await window.electronAPI.dialog.showSaveDialog({
      title: `导出分类: ${category.name}`,
      defaultPath: `snippets-${category.name.toLowerCase()}.json`
    });

    if (result.data.canceled) return;

    const filePath = result.data.filePath;
    const exportResult = await window.electronAPI.export.category(this.currentCategory, filePath);

    if (exportResult.success) {
      this.showToast('导出成功', 'success');
    } else {
      this.showToast(exportResult.error || '导出失败', 'error');
    }
  },

  async showExportOptions() {
    if (this.currentCategory === 'all' || this.currentCategory === 'favorites') {
      this.exportAll();
      return;
    }

    const choice = await this.showOptions(
      '选择导出范围',
      '请选择要导出的内容：',
      [
        { label: `导出当前分类`, value: 'category' },
        { label: '导出全部', value: 'all' }
      ]
    );

    if (choice === 'category') {
      this.exportCurrentCategory();
    } else if (choice === 'all') {
      this.exportAll();
    }
  },

  async importSnippets() {
    const result = await window.electronAPI.dialog.showOpenDialog({
      title: '选择要导入的 JSON 文件'
    });

    if (result.data.canceled || result.data.filePaths.length === 0) return;

    const filePath = result.data.filePaths[0];
    
    const choice = await this.showOptions(
      '处理重复项',
      '如果发现标题重复的片段，应该如何处理？',
      [
        { label: '跳过重复项', value: 'skip' },
        { label: '覆盖现有项', value: 'overwrite' }
      ]
    );

    if (!choice) return;

    const options = {
      skipDuplicates: choice === 'skip',
      overwriteDuplicates: choice === 'overwrite'
    };

    const importResult = await window.electronAPI.import.fromFile(filePath, options);

    if (importResult.success) {
      await this.loadCategories();
      await this.loadSnippets();
      this.updateCounts();
      this.renderCategoryList();
      this.renderSnippets();
      this.showToast('导入成功', 'success');
    } else {
      this.showToast(importResult.error || '导入失败', 'error');
    }
  },

  search(keyword) {
    this.searchKeyword = keyword;
    this.renderSnippets();
  },

  sortBy(option) {
    this.sortBy = option;
    this.renderSnippets();
  },

  parseTags(input) {
    if (!input || !input.trim()) return [];
    return input
      .split(/[,\s]+/)
      .map(t => t.trim())
      .filter(t => t.length > 0);
  },

  updateTagsPreview() {
    const input = document.getElementById('formTags').value;
    const tags = this.parseTags(input);
    const preview = document.getElementById('tagsPreview');
    
    if (tags.length === 0) {
      preview.innerHTML = '';
      return;
    }

    preview.innerHTML = tags.map((tag, idx) => 
      `<span class="tag-badge">${this.escapeHtml(tag)}<span class="tag-remove" data-idx="${idx}">&times;</span></span>`
    ).join('');

    preview.querySelectorAll('.tag-remove').forEach(btn => {
      btn.addEventListener('click', () => {
        const idx = parseInt(btn.dataset.idx);
        const currentTags = this.parseTags(document.getElementById('formTags').value);
        currentTags.splice(idx, 1);
        document.getElementById('formTags').value = currentTags.join(' ');
        this.updateTagsPreview();
      });
    });
  },

  formatDate(isoString) {
    const date = new Date(isoString);
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    });
  },

  escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  },

  showToast(message, type = 'success') {
    const container = document.getElementById('toastContainer');
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    
    let icon = '✓';
    if (type === 'error') icon = '✕';
    if (type === 'warning') icon = '⚠';
    
    toast.innerHTML = `<span>${icon}</span><span>${this.escapeHtml(message)}</span>`;
    container.appendChild(toast);

    setTimeout(() => {
      toast.style.animation = 'slideOut 0.3s ease forwards';
      setTimeout(() => toast.remove(), 300);
    }, 2000);
  },

  async showConfirm(title, message) {
    return new Promise((resolve) => {
      const modal = document.getElementById('modal');
      document.getElementById('modalTitle').textContent = title;
      document.getElementById('modalBody').innerHTML = `<p>${this.escapeHtml(message)}</p>`;
      document.getElementById('modalFooter').innerHTML = `
        <button class="btn btn-secondary" id="modalNo">取消</button>
        <button class="btn btn-primary" id="modalYes">确定</button>
      `;

      modal.style.display = 'flex';

      const closeModal = (result) => {
        modal.style.display = 'none';
        resolve(result);
      };

      document.getElementById('modalYes').onclick = () => closeModal(true);
      document.getElementById('modalNo').onclick = () => closeModal(false);
      document.getElementById('modalClose').onclick = () => closeModal(false);
    });
  },

  async showPrompt(title, message, defaultValue = '') {
    return new Promise((resolve) => {
      const modal = document.getElementById('modal');
      document.getElementById('modalTitle').textContent = title;
      document.getElementById('modalBody').innerHTML = `
        <p>${this.escapeHtml(message)}</p>
        <div class="form-group">
          <input type="text" id="promptInput" class="form-input" value="${this.escapeHtml(defaultValue)}">
        </div>
      `;
      document.getElementById('modalFooter').innerHTML = `
        <button class="btn btn-secondary" id="modalCancel">取消</button>
        <button class="btn btn-primary" id="modalOk">确定</button>
      `;

      modal.style.display = 'flex';
      const input = document.getElementById('promptInput');
      input.focus();
      input.select();

      const closeModal = (result) => {
        modal.style.display = 'none';
        resolve(result);
      };

      document.getElementById('modalOk').onclick = () => closeModal(input.value);
      document.getElementById('modalCancel').onclick = () => closeModal(null);
      document.getElementById('modalClose').onclick = () => closeModal(null);

      input.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
          closeModal(input.value);
        } else if (e.key === 'Escape') {
          closeModal(null);
        }
      });
    });
  },

  async showOptions(title, message, options) {
    return new Promise((resolve) => {
      const modal = document.getElementById('modal');
      document.getElementById('modalTitle').textContent = title;
      
      let optionsHtml = `<p>${this.escapeHtml(message)}</p><div class="option-group">`;
      options.forEach((opt, idx) => {
        optionsHtml += `
          <label>
            <input type="radio" name="modalOption" value="${this.escapeHtml(opt.value)}" ${idx === 0 ? 'checked' : ''}>
            ${this.escapeHtml(opt.label)}
          </label>
        `;
      });
      optionsHtml += '</div>';
      
      document.getElementById('modalBody').innerHTML = optionsHtml;
      document.getElementById('modalFooter').innerHTML = `
        <button class="btn btn-secondary" id="modalCancel">取消</button>
        <button class="btn btn-primary" id="modalOk">确定</button>
      `;

      modal.style.display = 'flex';

      const closeModal = (result) => {
        modal.style.display = 'none';
        resolve(result);
      };

      document.getElementById('modalOk').onclick = () => {
        const selected = document.querySelector('input[name="modalOption"]:checked');
        closeModal(selected ? selected.value : null);
      };
      document.getElementById('modalCancel').onclick = () => closeModal(null);
      document.getElementById('modalClose').onclick = () => closeModal(null);
    });
  },

  bindEvents() {
    document.getElementById('searchInput').addEventListener('input', (e) => {
      const value = e.target.value;
      const clearBtn = document.getElementById('searchClearBtn');
      clearBtn.style.display = value ? 'inline-block' : 'none';
      this.search(value);
    });

    document.getElementById('searchClearBtn').addEventListener('click', () => {
      document.getElementById('searchInput').value = '';
      document.getElementById('searchClearBtn').style.display = 'none';
      this.search('');
    });

    document.getElementById('sortSelect').addEventListener('change', (e) => {
      this.sortBy(e.target.value);
    });

    document.getElementById('addSnippetBtn').addEventListener('click', () => {
      this.createNewSnippet();
    });

    document.getElementById('copyBtn').addEventListener('click', () => {
      this.copyCode();
    });

    document.getElementById('saveBtn').addEventListener('click', () => {
      this.saveSnippet();
    });

    document.getElementById('cancelBtn').addEventListener('click', () => {
      this.cancelEdit();
    });

    document.getElementById('editBtn').addEventListener('click', () => {
      this.editSnippet();
    });

    document.getElementById('deleteBtn').addEventListener('click', () => {
      this.deleteSnippet();
    });

    document.getElementById('addCategoryBtn').addEventListener('click', () => {
      this.addCategory();
    });

    document.getElementById('importBtn').addEventListener('click', () => {
      this.importSnippets();
    });

    document.getElementById('exportBtn').addEventListener('click', () => {
      this.showExportOptions();
    });

    document.querySelectorAll('[data-id="all"], [data-id="favorites"]').forEach(el => {
      el.addEventListener('click', (e) => {
        if (e.target === el || el.contains(e.target)) {
          this.selectCategory(el.dataset.id);
        }
      });
    });

    document.getElementById('formTags').addEventListener('input', () => {
      this.updateTagsPreview();
    });

    document.getElementById('formCategory').addEventListener('change', (e) => {
      const lang = this.guessLanguageByCategory(e.target.value);
      document.getElementById('formLanguage').value = lang;
    });

    document.addEventListener('keydown', (e) => {
      if (e.ctrlKey || e.metaKey) {
        if (e.key === 'n') {
          e.preventDefault();
          this.createNewSnippet();
        } else if (e.key === 's') {
          e.preventDefault();
          if (this.isEditing) {
            this.saveSnippet();
          }
        } else if (e.key === 'f') {
          e.preventDefault();
          document.getElementById('searchInput').focus();
        }
      } else if (e.key === 'Escape') {
        if (this.isEditing) {
          this.cancelEdit();
        }
      }
    });
  }
};

document.addEventListener('DOMContentLoaded', () => {
  App.init();
});
