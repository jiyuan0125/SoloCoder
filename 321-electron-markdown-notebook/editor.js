class EditorManager {
  constructor() {
    this.element = document.getElementById('editor');
    this.currentFile = null;
    this.fileSize = 0;
    this.originalContent = '';
    this.isDirty = false;
    this.autoSaveTimer = null;
    this.debounceTimer = null;
    this.LARGE_FILE_THRESHOLD = 1024 * 1024;
    this.DEBOUNCE_DELAY = 500;
    this.AUTO_SAVE_DELAY = 1000;
    this.saveStatusEl = document.getElementById('save-status');
    this.initEventListeners();
  }

  initEventListeners() {
    this.element.addEventListener('input', () => {
      this.handleInput();
    });

    this.element.addEventListener('keydown', (e) => {
      this.handleKeydown(e);
    });

    this.element.addEventListener('tab', (e) => {
      e.preventDefault();
      this.insertTab();
    });
  }

  handleInput() {
    this.isDirty = this.element.value !== this.originalContent;
    this.updateSaveStatus('dirty');
    
    this.scheduleAutoSave();
    
    const event = new CustomEvent('content-changed', { 
      detail: { content: this.element.value, isLargeFile: this.isLargeFile() }
    });
    document.dispatchEvent(event);
  }

  handleKeydown(e) {
    if (e.key === 'Tab') {
      e.preventDefault();
      this.insertTab();
    }
    
    if (e.ctrlKey && e.key === 's') {
      e.preventDefault();
      this.save();
    }
  }

  insertTab() {
    const start = this.element.selectionStart;
    const end = this.element.selectionEnd;
    const value = this.element.value;
    
    this.element.value = value.substring(0, start) + '  ' + value.substring(end);
    this.element.selectionStart = this.element.selectionEnd = start + 2;
    
    this.handleInput();
  }

  scheduleAutoSave() {
    if (this.autoSaveTimer) {
      clearTimeout(this.autoSaveTimer);
    }
    
    this.autoSaveTimer = setTimeout(() => {
      if (this.currentFile && this.isDirty) {
        this.save();
      }
    }, this.AUTO_SAVE_DELAY);
  }

  async save() {
    if (!this.currentFile || !this.isDirty) {
      return { success: true };
    }

    this.updateSaveStatus('saving');
    
    const result = await window.electronAPI.saveFile(
      this.currentFile.path,
      this.element.value
    );

    if (result.success) {
      this.originalContent = this.element.value;
      this.isDirty = false;
      this.updateSaveStatus('saved');
      
      setTimeout(() => {
        this.updateSaveStatus('');
      }, 2000);
    } else {
      this.updateSaveStatus('error');
      console.error('保存失败:', result.error);
    }

    return result;
  }

  async loadFile(fileItem) {
    if (this.isDirty && this.currentFile) {
      await this.save();
    }

    const result = await window.electronAPI.readFile(fileItem.path);
    
    if (result.success) {
      this.currentFile = fileItem;
      this.fileSize = result.size;
      this.originalContent = result.content;
      this.element.value = result.content;
      this.isDirty = false;
      this.updateSaveStatus('');
      
      const event = new CustomEvent('file-loaded', { 
        detail: { 
          file: fileItem, 
          content: result.content,
          isLargeFile: this.isLargeFile()
        }
      });
      document.dispatchEvent(event);
      
      return { success: true };
    } else {
      console.error('加载文件失败:', result.error);
      return { success: false, error: result.error };
    }
  }

  clear() {
    this.currentFile = null;
    this.fileSize = 0;
    this.originalContent = '';
    this.element.value = '';
    this.isDirty = false;
    this.updateSaveStatus('');
  }

  isLargeFile() {
    return this.fileSize > this.LARGE_FILE_THRESHOLD;
  }

  updateSaveStatus(status) {
    this.saveStatusEl.className = 'save-status';
    
    switch (status) {
      case 'saving':
        this.saveStatusEl.textContent = '保存中...';
        this.saveStatusEl.classList.add('saving');
        break;
      case 'saved':
        this.saveStatusEl.textContent = '已保存';
        this.saveStatusEl.classList.add('saved');
        break;
      case 'dirty':
        this.saveStatusEl.textContent = '未保存';
        break;
      case 'error':
        this.saveStatusEl.textContent = '保存失败';
        this.saveStatusEl.style.color = '#d93025';
        break;
      default:
        this.saveStatusEl.textContent = '';
        break;
    }
  }

  getContent() {
    return this.element.value;
  }

  getCurrentFile() {
    return this.currentFile;
  }

  hasUnsavedChanges() {
    return this.isDirty;
  }

  focus() {
    this.element.focus();
  }
}

window.EditorManager = EditorManager;
