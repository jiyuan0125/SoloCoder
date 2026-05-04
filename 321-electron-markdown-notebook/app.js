class App {
  constructor() {
    this.treeManager = null;
    this.editorManager = null;
    this.previewManager = null;
    this.currentLayout = 'split';
    this.editorContainer = document.getElementById('editor-container');
    this.fileNameEl = document.getElementById('current-file-name');
    this.init();
  }

  async init() {
    this.treeManager = new TreeManager();
    this.editorManager = new EditorManager();
    this.previewManager = new PreviewManager();

    this.setupEventListeners();
    this.setupLayoutButtons();
    this.setupKeyboardShortcuts();

    await this.treeManager.load();
    this.previewManager.showEmptyState();
  }

  setupEventListeners() {
    document.addEventListener('file-selected', async (e) => {
      const fileItem = e.detail;
      await this.openFile(fileItem);
    });

    document.addEventListener('file-loaded', (e) => {
      const { file, content, isLargeFile } = e.detail;
      this.updateFileName(file.name);
      this.previewManager.setLargeFileMode(isLargeFile);
      this.previewManager.render(content, true);
    });

    document.addEventListener('content-changed', (e) => {
      const { content, isLargeFile } = e.detail;
      this.previewManager.setLargeFileMode(isLargeFile);
      this.previewManager.render(content);
    });

    document.addEventListener('file-deleted', (e) => {
      const deletedFile = e.detail;
      const currentFile = this.editorManager.getCurrentFile();
      
      if (currentFile && currentFile.path === deletedFile.path) {
        this.editorManager.clear();
        this.previewManager.clear();
        this.previewManager.showEmptyState();
        this.updateFileName('未打开文件');
      }
    });
  }

  async openFile(fileItem) {
    if (!fileItem || fileItem.type !== 'file') {
      return;
    }

    const result = await this.editorManager.loadFile(fileItem);
    if (result.success) {
      this.editorManager.focus();
    }
  }

  updateFileName(name) {
    this.fileNameEl.textContent = name;
    document.title = name ? `${name} - Markdown 笔记` : 'Markdown 笔记';
  }

  setupLayoutButtons() {
    const layoutBtns = document.querySelectorAll('.layout-btn');
    
    layoutBtns.forEach(btn => {
      btn.addEventListener('click', () => {
        const layout = btn.dataset.layout;
        this.setLayout(layout);
      });
    });
  }

  setLayout(layout) {
    this.currentLayout = layout;
    
    const layoutBtns = document.querySelectorAll('.layout-btn');
    layoutBtns.forEach(btn => {
      btn.classList.toggle('active', btn.dataset.layout === layout);
    });

    this.editorContainer.className = 'editor-container';
    this.editorContainer.classList.add(`layout-${layout}`);
  }

  setupKeyboardShortcuts() {
    document.addEventListener('keydown', (e) => {
      if (e.ctrlKey && e.key === 'n') {
        e.preventDefault();
        this.treeManager.showNewFileDialog();
      }
    });
  }
}

document.addEventListener('DOMContentLoaded', () => {
  new App();
});
