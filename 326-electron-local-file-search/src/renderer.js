class FileSearchApp {
  constructor() {
    this.selectedDirectory = null;
    this.indexReady = false;
    this.searchHistory = this.loadSearchHistory();
    this.currentSearchQuery = '';
    this.debounceTimer = null;
    
    this.initElements();
    this.initEventListeners();
    this.initIPCListeners();
    this.loadSavedState();
  }

  initElements() {
    this.searchInput = document.getElementById('searchInput');
    this.searchHistoryBtn = document.getElementById('searchHistoryBtn');
    this.historyDropdown = document.getElementById('historyDropdown');
    this.historyList = document.getElementById('historyList');
    this.clearHistoryBtn = document.getElementById('clearHistoryBtn');
    
    this.directoryInput = document.getElementById('directoryInput');
    this.selectDirBtn = document.getElementById('selectDirBtn');
    
    this.fileTypeSelect = document.getElementById('fileTypeSelect');
    this.customExtensions = document.getElementById('customExtensions');
    
    this.indexBtn = document.getElementById('indexBtn');
    this.reindexBtn = document.getElementById('reindexBtn');
    
    this.progressSection = document.getElementById('progressSection');
    this.progressText = document.getElementById('progressText');
    this.progressPercent = document.getElementById('progressPercent');
    this.progressFill = document.getElementById('progressFill');
    this.scanningFile = document.getElementById('scanningFile');
    
    this.statsBar = document.getElementById('statsBar');
    this.statFiles = document.getElementById('statFiles');
    this.statTerms = document.getElementById('statTerms');
    this.statDirectory = document.getElementById('statDirectory');
    
    this.resultsList = document.getElementById('resultsList');
    this.resultCount = document.getElementById('resultCount');
    this.emptyState = document.getElementById('emptyState');
  }

  initEventListeners() {
    this.selectDirBtn.addEventListener('click', () => this.selectDirectory());
    this.indexBtn.addEventListener('click', () => this.startIndexing());
    this.reindexBtn.addEventListener('click', () => this.reindex());
    
    this.fileTypeSelect.addEventListener('change', () => this.handleFileTypeChange());
    this.customExtensions.addEventListener('input', () => this.handleCustomExtensionChange());
    
    this.searchInput.addEventListener('input', (e) => this.handleSearchInput(e));
    this.searchInput.addEventListener('keydown', (e) => this.handleSearchKeyDown(e));
    
    this.searchHistoryBtn.addEventListener('click', () => this.toggleHistoryDropdown());
    this.clearHistoryBtn.addEventListener('click', () => this.clearSearchHistory());
    
    document.addEventListener('click', (e) => {
      if (!this.searchHistoryBtn.contains(e.target) && 
          !this.historyDropdown.contains(e.target)) {
        this.hideHistoryDropdown();
      }
    });
    
    this.resultsList.addEventListener('dblclick', (e) => {
      const resultItem = e.target.closest('.result-item');
      if (resultItem && resultItem.dataset.filePath) {
        this.openFile(resultItem.dataset.filePath);
      }
    });
  }

  initIPCListeners() {
    window.electronAPI.onIndexProgress((progress) => {
      this.updateProgress(progress);
    });
    
    window.electronAPI.onFileScanning((filePath) => {
      this.updateScanningFile(filePath);
    });
  }

  loadSavedState() {
    const savedDirectory = localStorage.getItem('lastDirectory');
    if (savedDirectory) {
      this.selectedDirectory = savedDirectory;
      this.directoryInput.value = savedDirectory;
      this.tryLoadCache();
    }
  }

  saveState() {
    if (this.selectedDirectory) {
      localStorage.setItem('lastDirectory', this.selectedDirectory);
    }
  }

  loadSearchHistory() {
    const saved = localStorage.getItem('searchHistory');
    return saved ? JSON.parse(saved) : [];
  }

  saveSearchHistory() {
    localStorage.setItem('searchHistory', JSON.stringify(this.searchHistory));
  }

  addToSearchHistory(query) {
    if (!query || query.trim() === '') return;
    
    const trimmedQuery = query.trim();
    
    this.searchHistory = this.searchHistory.filter(q => q !== trimmedQuery);
    this.searchHistory.unshift(trimmedQuery);
    
    if (this.searchHistory.length > 20) {
      this.searchHistory = this.searchHistory.slice(0, 20);
    }
    
    this.saveSearchHistory();
    this.updateHistoryUI();
  }

  updateHistoryUI() {
    this.historyList.innerHTML = '';
    
    if (this.searchHistory.length === 0) {
      const li = document.createElement('li');
      li.textContent = '暂无搜索历史';
      li.style.color = 'var(--text-muted)';
      li.style.cursor = 'default';
      this.historyList.appendChild(li);
      return;
    }
    
    this.searchHistory.forEach((query) => {
      const li = document.createElement('li');
      li.textContent = query;
      li.addEventListener('click', () => {
        this.searchInput.value = query;
        this.hideHistoryDropdown();
        this.search(query);
      });
      this.historyList.appendChild(li);
    });
  }

  toggleHistoryDropdown() {
    this.updateHistoryUI();
    this.historyDropdown.classList.toggle('visible');
    this.searchHistoryBtn.classList.toggle('active', this.historyDropdown.classList.contains('visible'));
  }

  hideHistoryDropdown() {
    this.historyDropdown.classList.remove('visible');
    this.searchHistoryBtn.classList.remove('active');
  }

  clearSearchHistory() {
    this.searchHistory = [];
    this.saveSearchHistory();
    this.updateHistoryUI();
  }

  async selectDirectory() {
    const directory = await window.electronAPI.selectDirectory();
    if (directory) {
      this.selectedDirectory = directory;
      this.directoryInput.value = directory;
      this.saveState();
      this.tryLoadCache();
    }
  }

  async tryLoadCache() {
    if (!this.selectedDirectory) return;
    
    const result = await window.electronAPI.loadCache(this.selectedDirectory);
    if (result.success && result.hasCache) {
      this.indexReady = true;
      this.showReindexButton();
      this.updateStats();
    }
  }

  getSelectedExtensions() {
    const value = this.fileTypeSelect.value;
    
    if (value === 'all') {
      return ['.txt', '.md', '.js', '.py', '.json', '.html', '.css', '.xml', '.log', '.csv'];
    }
    
    if (value === 'custom') {
      const custom = this.customExtensions.value;
      if (!custom) return [];
      
      return custom.split(',')
        .map(ext => ext.trim())
        .filter(ext => ext.length > 0)
        .map(ext => ext.startsWith('.') ? ext : `.${ext}`);
    }
    
    return value.split(',').map(ext => ext.trim());
  }

  handleFileTypeChange() {
    const value = this.fileTypeSelect.value;
    if (value === 'custom') {
      this.customExtensions.style.display = 'inline-block';
      this.customExtensions.focus();
    } else {
      this.customExtensions.style.display = 'none';
    }
  }

  handleCustomExtensionChange() {
    if (this.indexReady && this.selectedDirectory) {
      if (this.searchInput.value.trim()) {
        this.debouncedSearch();
      }
    }
  }

  async startIndexing() {
    if (!this.selectedDirectory) {
      alert('请先选择一个目录');
      return;
    }
    
    const extensions = this.getSelectedExtensions();
    if (extensions.length === 0) {
      alert('请至少选择一种文件类型');
      return;
    }
    
    this.indexBtn.disabled = true;
    this.showProgress();
    
    try {
      const result = await window.electronAPI.indexDirectory(this.selectedDirectory, extensions);
      
      if (result.success) {
        this.indexReady = true;
        this.showReindexButton();
        this.updateStats();
        
        if (this.searchInput.value.trim()) {
          this.search(this.searchInput.value);
        }
      } else {
        alert(`索引失败: ${result.error}`);
      }
    } catch (error) {
      alert(`索引过程出错: ${error.message}`);
    } finally {
      this.hideProgress();
      this.indexBtn.disabled = false;
    }
  }

  async reindex() {
    if (!this.selectedDirectory) return;
    
    const extensions = this.getSelectedExtensions();
    if (extensions.length === 0) {
      alert('请至少选择一种文件类型');
      return;
    }
    
    this.reindexBtn.disabled = true;
    this.showProgress();
    
    try {
      const result = await window.electronAPI.reindexDirectory(this.selectedDirectory, extensions);
      
      if (result.success) {
        this.indexReady = true;
        this.updateStats();
        
        if (this.searchInput.value.trim()) {
          this.search(this.searchInput.value);
        }
      } else {
        alert(`重新索引失败: ${result.error}`);
      }
    } catch (error) {
      alert(`重新索引过程出错: ${error.message}`);
    } finally {
      this.hideProgress();
      this.reindexBtn.disabled = false;
    }
  }

  showProgress() {
    this.progressSection.style.display = 'block';
    this.updateProgress({ current: 0, total: 0, percent: 0 });
  }

  hideProgress() {
    this.progressSection.style.display = 'none';
  }

  updateProgress(progress) {
    const { current, total, percent } = progress;
    this.progressText.textContent = `正在建立索引... (${current}/${total})`;
    this.progressPercent.textContent = `${percent}%`;
    this.progressFill.style.width = `${percent}%`;
  }

  updateScanningFile(filePath) {
    this.scanningFile.textContent = filePath;
    this.scanningFile.title = filePath;
  }

  showReindexButton() {
    this.indexBtn.style.display = 'none';
    this.reindexBtn.style.display = 'inline-block';
  }

  async updateStats() {
    const result = await window.electronAPI.getIndexStats();
    if (result.success && result.data) {
      const { totalFiles, totalTerms, directory } = result.data;
      this.statFiles.textContent = totalFiles || 0;
      this.statTerms.textContent = totalTerms || 0;
      this.statDirectory.textContent = directory || '-';
      this.statsBar.style.display = 'flex';
    }
  }

  handleSearchInput(e) {
    this.debouncedSearch();
  }

  handleSearchKeyDown(e) {
    if (e.key === 'Enter') {
      this.search(this.searchInput.value);
    }
  }

  debouncedSearch() {
    if (this.debounceTimer) {
      clearTimeout(this.debounceTimer);
    }
    
    this.debounceTimer = setTimeout(() => {
      const query = this.searchInput.value;
      this.search(query, false);
    }, 300);
  }

  async search(query, addToHistory = true) {
    const trimmedQuery = query.trim();
    
    if (!trimmedQuery) {
      this.clearResults();
      return;
    }
    
    if (!this.indexReady) {
      this.showResults([], '请先建立索引');
      return;
    }
    
    if (addToHistory) {
      this.addToSearchHistory(trimmedQuery);
    }
    
    const extensions = this.getSelectedExtensions();
    const result = await window.electronAPI.search(trimmedQuery, {
      extensions: extensions.length > 0 ? extensions : null
    });
    
    if (result.success) {
      this.showResults(result.data, trimmedQuery);
    } else {
      this.showResults([], `搜索出错: ${result.error}`);
    }
  }

  clearResults() {
    this.resultsList.innerHTML = '';
    this.resultCount.textContent = '搜索结果: 0';
    this.emptyState.style.display = 'flex';
    this.resultsList.appendChild(this.emptyState);
  }

  showResults(results, query) {
    this.resultsList.innerHTML = '';
    this.resultCount.textContent = `搜索结果: ${results.length}`;
    
    if (results.length === 0) {
      this.emptyState.style.display = 'flex';
      this.emptyState.querySelector('p:first-of-type').textContent = 
        typeof query === 'string' ? `未找到包含 "${query}" 的结果` : query;
      this.resultsList.appendChild(this.emptyState);
      return;
    }
    
    const searchTerms = this.extractSearchTerms(query);
    
    results.forEach((result, index) => {
      const resultItem = this.createResultItem(result, searchTerms, index);
      this.resultsList.appendChild(resultItem);
    });
  }

  extractSearchTerms(query) {
    if (!query || typeof query !== 'string') return [];
    
    const tokens = query.trim().split(/\s+/);
    const terms = [];
    
    for (const token of tokens) {
      if (token === '|' || token === 'OR') continue;
      if (token.startsWith('-')) {
        terms.push(token.slice(1));
      } else {
        terms.push(token);
      }
    }
    
    return terms.filter(t => t.length > 0);
  }

  createResultItem(result, searchTerms, index) {
    const { filePath, matches } = result;
    
    const item = document.createElement('div');
    item.className = 'result-item';
    item.dataset.filePath = filePath;
    item.dataset.index = index;
    
    const header = document.createElement('div');
    header.className = 'result-header';
    header.innerHTML = `
      <span class="result-file-path">${this.escapeHtml(filePath)}</span>
      <span class="result-match-count">${matches.length} 个匹配</span>
    `;
    item.appendChild(header);
    
    const matchesContainer = document.createElement('div');
    matchesContainer.className = 'result-matches';
    
    matches.forEach((match) => {
      const matchEl = document.createElement('div');
      matchEl.className = 'result-match';
      matchEl.dataset.filePath = filePath;
      
      const highlightedContent = this.highlightText(match.content, searchTerms);
      
      matchEl.innerHTML = `
        <span class="match-line-number">${match.line}</span>
        <span class="match-content">${highlightedContent}</span>
      `;
      matchesContainer.appendChild(matchEl);
    });
    
    item.appendChild(matchesContainer);
    return item;
  }

  highlightText(text, terms) {
    if (!terms || terms.length === 0) {
      return this.escapeHtml(text);
    }
    
    const pattern = new RegExp(
      `(${terms.map(t => this.escapeRegex(t)).join('|')})`,
      'gi'
    );
    
    const escaped = this.escapeHtml(text);
    return escaped.replace(pattern, '<span class="match-highlight">$1</span>');
  }

  escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  escapeRegex(string) {
    return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  }

  async openFile(filePath) {
    try {
      await window.electronAPI.openFile(filePath);
    } catch (error) {
      alert(`无法打开文件: ${error.message}`);
    }
  }
}

document.addEventListener('DOMContentLoaded', () => {
  window.app = new FileSearchApp();
});
