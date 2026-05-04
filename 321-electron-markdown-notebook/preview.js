class PreviewManager {
  constructor() {
    this.element = document.getElementById('preview');
    this.debounceTimer = null;
    this.DEBOUNCE_DELAY = 500;
    this.isLargeFile = false;
    this.currentContent = '';
  }

  async render(content, force = false) {
    if (this.isLargeFile && !force) {
      this.scheduleDebouncedRender(content);
      return;
    }
    
    this.currentContent = content;
    await this.doRender(content);
  }

  scheduleDebouncedRender(content) {
    if (this.debounceTimer) {
      clearTimeout(this.debounceTimer);
    }
    
    this.currentContent = content;
    
    this.debounceTimer = setTimeout(async () => {
      await this.doRender(this.currentContent);
    }, this.DEBOUNCE_DELAY);
  }

  async doRender(content) {
    try {
      const html = await window.electronAPI.renderMarkdown(content);
      this.element.innerHTML = html;
      this.processTaskLists();
    } catch (err) {
      console.error('渲染预览失败:', err);
      this.element.innerHTML = `<p class="error">渲染失败: ${err.message}</p>`;
    }
  }

  processTaskLists() {
    const checkboxes = this.element.querySelectorAll('input[type="checkbox"]');
    checkboxes.forEach(checkbox => {
      checkbox.disabled = true;
    });
  }

  setLargeFileMode(isLarge) {
    this.isLargeFile = isLarge;
  }

  clear() {
    if (this.debounceTimer) {
      clearTimeout(this.debounceTimer);
      this.debounceTimer = null;
    }
    this.element.innerHTML = '';
    this.currentContent = '';
  }

  showEmptyState() {
    this.element.innerHTML = `
      <div class="empty-state" style="height: 100%;">
        <div class="empty-state-icon">📄</div>
        <div class="empty-state-text">选择或创建一个笔记开始</div>
      </div>
    `;
  }
}

window.PreviewManager = PreviewManager;
