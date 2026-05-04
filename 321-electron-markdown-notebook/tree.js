class TreeManager {
  constructor() {
    this.rootPath = null;
    this.selectedItem = null;
    this.contextMenuItem = null;
    this.expandedDirs = new Set();
    this.element = document.getElementById('file-tree');
    this.contextMenu = document.getElementById('context-menu');
    this.initEventListeners();
  }

  initEventListeners() {
    document.getElementById('refresh-btn').addEventListener('click', () => this.refresh());
    document.getElementById('new-file-btn').addEventListener('click', () => this.showNewFileDialog());
    document.getElementById('new-folder-btn').addEventListener('click', () => this.showNewFolderDialog());

    document.addEventListener('click', (e) => {
      if (!this.contextMenu.contains(e.target)) {
        this.hideContextMenu();
      }
    });

    this.contextMenu.querySelectorAll('.context-menu-item').forEach(item => {
      item.addEventListener('click', (e) => {
        const action = e.currentTarget.dataset.action;
        this.handleContextMenuAction(action);
        this.hideContextMenu();
      });
    });
  }

  async load() {
    try {
      const treeData = await window.electronAPI.getDirectoryTree();
      this.rootPath = treeData.rootPath;
      this.render(treeData);
    } catch (err) {
      console.error('加载目录树失败:', err);
      this.showEmptyState();
    }
  }

  render(treeData) {
    this.element.innerHTML = '';
    
    if (!treeData.children || treeData.children.length === 0) {
      this.showEmptyState();
      return;
    }

    const rootNode = document.createElement('div');
    rootNode.className = 'tree-root';
    
    treeData.children.forEach(child => {
      const node = this.createTreeNode(child);
      if (node) {
        rootNode.appendChild(node);
      }
    });

    this.element.appendChild(rootNode);
  }

  createTreeNode(item, level = 0) {
    const node = document.createElement('div');
    node.className = 'tree-node';
    node.dataset.path = item.path;
    node.dataset.type = item.type;
    node.dataset.name = item.name;

    const content = document.createElement('div');
    content.className = 'tree-node-content';
    
    const expandIcon = document.createElement('span');
    expandIcon.className = 'tree-expand-icon';
    expandIcon.textContent = item.type === 'directory' ? 
      (this.expandedDirs.has(item.path) ? '▼' : '▶') : '';

    const icon = document.createElement('span');
    icon.className = 'tree-icon';
    icon.textContent = item.type === 'directory' ? '📁' : '📄';

    const name = document.createElement('span');
    name.className = 'tree-name';
    name.textContent = item.name;

    content.appendChild(expandIcon);
    content.appendChild(icon);
    content.appendChild(name);
    node.appendChild(content);

    content.addEventListener('click', (e) => {
      if (e.target.classList.contains('tree-expand-icon') || 
          item.type === 'directory') {
        this.toggleExpand(node, item);
      }
      if (item.type === 'file') {
        this.selectFile(item, node);
      }
    });

    content.addEventListener('contextmenu', (e) => {
      e.preventDefault();
      this.showContextMenu(e.clientX, e.clientY, item);
    });

    if (item.type === 'directory' && item.children && item.children.length > 0) {
      const children = document.createElement('div');
      children.className = 'tree-children';
      if (!this.expandedDirs.has(item.path)) {
        children.classList.add('collapsed');
      }
      
      item.children.forEach(child => {
        const childNode = this.createTreeNode(child, level + 1);
        if (childNode) {
          children.appendChild(childNode);
        }
      });
      
      node.appendChild(children);
    }

    return node;
  }

  toggleExpand(node, item) {
    const children = node.querySelector('.tree-children');
    const expandIcon = node.querySelector('.tree-expand-icon');
    
    if (children) {
      if (children.classList.contains('collapsed')) {
        children.classList.remove('collapsed');
        expandIcon.textContent = '▼';
        this.expandedDirs.add(item.path);
      } else {
        children.classList.add('collapsed');
        expandIcon.textContent = '▶';
        this.expandedDirs.delete(item.path);
      }
    }
  }

  selectFile(item, node) {
    this.element.querySelectorAll('.tree-node-content.active').forEach(el => {
      el.classList.remove('active');
    });
    
    node.querySelector('.tree-node-content').classList.add('active');
    this.selectedItem = item;
    
    const event = new CustomEvent('file-selected', { detail: item });
    document.dispatchEvent(event);
  }

  showContextMenu(x, y, item) {
    this.contextMenuItem = item;
    
    this.contextMenu.style.left = `${x}px`;
    this.contextMenu.style.top = `${y}px`;
    this.contextMenu.classList.remove('hidden');

    const renameItem = this.contextMenu.querySelector('[data-action="rename"]');
    const deleteItem = this.contextMenu.querySelector('[data-action="delete"]');
    
    if (item) {
      renameItem.style.display = 'block';
      deleteItem.style.display = 'block';
    } else {
      renameItem.style.display = 'none';
      deleteItem.style.display = 'none';
    }
  }

  hideContextMenu() {
    this.contextMenu.classList.add('hidden');
    this.contextMenuItem = null;
  }

  handleContextMenuAction(action) {
    const item = this.contextMenuItem;
    const parentPath = item ? 
      (item.type === 'directory' ? item.path : this.rootPath) : 
      this.rootPath;

    switch (action) {
      case 'new-file':
        this.showNewFileDialog(parentPath);
        break;
      case 'new-folder':
        this.showNewFolderDialog(parentPath);
        break;
      case 'rename':
        if (item) {
          this.showRenameDialog(item);
        }
        break;
      case 'delete':
        if (item) {
          this.showDeleteConfirm(item);
        }
        break;
    }
  }

  showNewFileDialog(parentPath = null) {
    const targetPath = parentPath || this.rootPath;
    const defaultName = this.getDefaultFileName();
    
    this.showModal(
      '新建笔记',
      `
        <p>输入新笔记的名称：</p>
        <input type="text" id="new-file-name" value="${defaultName}" placeholder="filename.md">
        <div id="new-file-error" class="error-message hidden"></div>
      `,
      [
        { text: '取消', class: 'secondary', action: () => this.hideModal() },
        { text: '创建', class: 'primary', action: async () => {
          const name = document.getElementById('new-file-name').value.trim();
          if (!name) {
            this.showModalError('new-file-error', '请输入文件名');
            return;
          }
          if (!name.toLowerCase().endsWith('.md')) {
            this.showModalError('new-file-error', '文件名必须以 .md 结尾');
            return;
          }
          const result = await window.electronAPI.createFile(targetPath, name);
          if (result.success) {
            this.hideModal();
            this.refresh();
            this.loadAndSelectFile(result.path);
          } else {
            this.showModalError('new-file-error', result.error || '创建失败');
          }
        }}
      ]
    );

    setTimeout(() => {
      const input = document.getElementById('new-file-name');
      if (input) {
        input.focus();
        input.select();
        input.addEventListener('keydown', (e) => {
          if (e.key === 'Enter') {
            document.querySelector('#modal-footer .modal-btn.primary').click();
          }
        });
      }
    }, 50);
  }

  showNewFolderDialog(parentPath = null) {
    const targetPath = parentPath || this.rootPath;
    
    this.showModal(
      '新建文件夹',
      `
        <p>输入新文件夹的名称：</p>
        <input type="text" id="new-folder-name" value="新建文件夹" placeholder="文件夹名">
        <div id="new-folder-error" class="error-message hidden"></div>
      `,
      [
        { text: '取消', class: 'secondary', action: () => this.hideModal() },
        { text: '创建', class: 'primary', action: async () => {
          const name = document.getElementById('new-folder-name').value.trim();
          if (!name) {
            this.showModalError('new-folder-error', '请输入文件夹名');
            return;
          }
          const result = await window.electronAPI.createDirectory(targetPath, name);
          if (result.success) {
            this.hideModal();
            this.refresh();
          } else {
            this.showModalError('new-folder-error', result.error || '创建失败');
          }
        }}
      ]
    );

    setTimeout(() => {
      const input = document.getElementById('new-folder-name');
      if (input) {
        input.focus();
        input.select();
        input.addEventListener('keydown', (e) => {
          if (e.key === 'Enter') {
            document.querySelector('#modal-footer .modal-btn.primary').click();
          }
        });
      }
    }, 50);
  }

  showRenameDialog(item) {
    this.showModal(
      '重命名',
      `
        <p>输入新名称：</p>
        <input type="text" id="rename-name" value="${item.name}" placeholder="新名称">
        <div id="rename-error" class="error-message hidden"></div>
      `,
      [
        { text: '取消', class: 'secondary', action: () => this.hideModal() },
        { text: '确定', class: 'primary', action: async () => {
          const newName = document.getElementById('rename-name').value.trim();
          if (!newName) {
            this.showModalError('rename-error', '请输入名称');
            return;
          }
          if (newName === item.name) {
            this.hideModal();
            return;
          }
          const result = await window.electronAPI.renameItem(item.path, newName);
          if (result.success) {
            this.hideModal();
            this.refresh();
          } else if (result.conflict) {
            this.showModalError('rename-error', '目标名称已存在，无法覆盖');
          } else {
            this.showModalError('rename-error', result.error || '重命名失败');
          }
        }}
      ]
    );

    setTimeout(() => {
      const input = document.getElementById('rename-name');
      if (input) {
        input.focus();
        const dotIndex = item.name.lastIndexOf('.');
        input.setSelectionRange(0, dotIndex > 0 ? dotIndex : item.name.length);
        input.addEventListener('keydown', (e) => {
          if (e.key === 'Enter') {
            document.querySelector('#modal-footer .modal-btn.primary').click();
          }
        });
      }
    }, 50);
  }

  showDeleteConfirm(item) {
    const typeText = item.type === 'directory' ? '文件夹' : '笔记';
    
    this.showModal(
      '确认删除',
      `
        <p>确定要删除${typeText} <strong>${item.name}</strong> 吗？</p>
        <p style="color: var(--text-secondary); font-size: 13px;">此操作无法撤销。</p>
      `,
      [
        { text: '取消', class: 'secondary', action: () => this.hideModal() },
        { text: '删除', class: 'danger', action: async () => {
          const result = await window.electronAPI.deleteItem(item.path, item.type);
          if (result.success) {
            this.hideModal();
            const event = new CustomEvent('file-deleted', { detail: item });
            document.dispatchEvent(event);
            this.refresh();
          } else {
            alert('删除失败: ' + result.error);
          }
        }}
      ]
    );
  }

  showModal(title, body, buttons) {
    const overlay = document.getElementById('modal-overlay');
    const titleEl = document.getElementById('modal-title');
    const bodyEl = document.getElementById('modal-body');
    const footerEl = document.getElementById('modal-footer');
    const closeBtn = document.getElementById('modal-close');

    titleEl.textContent = title;
    bodyEl.innerHTML = body;
    footerEl.innerHTML = '';

    buttons.forEach(btn => {
      const button = document.createElement('button');
      button.className = `modal-btn ${btn.class}`;
      button.textContent = btn.text;
      button.addEventListener('click', btn.action);
      footerEl.appendChild(button);
    });

    closeBtn.onclick = () => this.hideModal();
    overlay.classList.remove('hidden');
  }

  hideModal() {
    document.getElementById('modal-overlay').classList.add('hidden');
  }

  showModalError(id, message) {
    const errorEl = document.getElementById(id);
    if (errorEl) {
      errorEl.textContent = message;
      errorEl.classList.remove('hidden');
    }
  }

  getDefaultFileName() {
    const today = new Date();
    const year = today.getFullYear();
    const month = String(today.getMonth() + 1).padStart(2, '0');
    const day = String(today.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}.md`;
  }

  async loadAndSelectFile(filePath) {
    await this.refresh();
    
    setTimeout(() => {
      const node = this.element.querySelector(`.tree-node[data-path="${filePath}"]`);
      if (node) {
        const item = {
          path: filePath,
          name: filePath.split(/[\\/]/).pop(),
          type: 'file'
        };
        this.selectFile(item, node);
      }
    }, 100);
  }

  showEmptyState() {
    this.element.innerHTML = `
      <div class="empty-state">
        <div class="empty-state-icon">📝</div>
        <div class="empty-state-text">暂无笔记</div>
        <div class="empty-state-text" style="font-size: 12px; margin-top: 8px;">点击上方按钮创建新笔记</div>
      </div>
    `;
  }

  refresh() {
    return this.load();
  }

  getSelectedItem() {
    return this.selectedItem;
  }

  getRootPath() {
    return this.rootPath;
  }
}

window.TreeManager = TreeManager;
