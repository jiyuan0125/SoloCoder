class ImageBatchProcessor {
  constructor() {
    this.images = [];
    this.processedFiles = [];
    this.currentComparisonIndex = 0;
    this.outputDirectory = null;
    this.lastResults = null;
    
    this.init();
  }
  
  init() {
    this.bindElements();
    this.bindEvents();
  }
  
  bindElements() {
    // 导入区域
    this.dropZone = document.getElementById('dropZone');
    this.selectFilesBtn = document.getElementById('selectFilesBtn');
    this.selectFolderBtn = document.getElementById('selectFolderBtn');
    
    // 操作面板
    this.operationsPanel = document.getElementById('operationsPanel');
    this.previewSection = document.getElementById('previewSection');
    
    // 尺寸调整
    this.enableResize = document.getElementById('enableResize');
    this.resizeContent = document.getElementById('resizeContent');
    this.resizeModeRadios = document.querySelectorAll('input[name="resizeMode"]');
    this.resizePixels = document.getElementById('resizePixels');
    this.resizePercent = document.getElementById('resizePercent');
    this.resizeWidth = document.getElementById('resizeWidth');
    this.resizeHeight = document.getElementById('resizeHeight');
    this.lockAspect = document.getElementById('lockAspect');
    this.resizePercentValue = document.getElementById('resizePercentValue');
    
    // 格式转换
    this.enableFormat = document.getElementById('enableFormat');
    this.formatContent = document.getElementById('formatContent');
    this.targetFormat = document.getElementById('targetFormat');
    
    // 批量重命名
    this.enableRename = document.getElementById('enableRename');
    this.renameContent = document.getElementById('renameContent');
    this.renameTemplate = document.getElementById('renameTemplate');
    this.previewValue = document.getElementById('previewValue');
    this.templateError = document.getElementById('templateError');
    
    // 执行按钮
    this.executeBtn = document.getElementById('executeBtn');
    this.clearBtn = document.getElementById('clearBtn');
    
    // 进度显示
    this.progressSection = document.getElementById('progressSection');
    this.progressFill = document.getElementById('progressFill');
    this.progressText = document.getElementById('progressText');
    this.progressPercent = document.getElementById('progressPercent');
    
    // 结果显示
    this.resultsSection = document.getElementById('resultsSection');
    this.successCount = document.getElementById('successCount');
    this.skippedCount = document.getElementById('skippedCount');
    this.failedCount = document.getElementById('failedCount');
    this.openFolderBtn = document.getElementById('openFolderBtn');
    this.showComparisonBtn = document.getElementById('showComparisonBtn');
    this.successList = document.getElementById('successList');
    this.skippedList = document.getElementById('skippedList');
    this.failedList = document.getElementById('failedList');
    
    // 对比预览
    this.comparisonSection = document.getElementById('comparisonSection');
    this.prevComparison = document.getElementById('prevComparison');
    this.nextComparison = document.getElementById('nextComparison');
    this.comparisonIndex = document.getElementById('comparisonIndex');
    this.comparisonOriginal = document.getElementById('comparisonOriginal');
    this.comparisonProcessed = document.getElementById('comparisonProcessed');
    this.closeComparisonBtn = document.getElementById('closeComparisonBtn');
    
    // 图片预览网格
    this.imageCount = document.getElementById('imageCount');
    this.previewGrid = document.getElementById('previewGrid');
  }
  
  bindEvents() {
    // 拖拽导入
    this.dropZone.addEventListener('dragover', (e) => {
      e.preventDefault();
      this.dropZone.classList.add('dragover');
    });
    
    this.dropZone.addEventListener('dragleave', () => {
      this.dropZone.classList.remove('dragover');
    });
    
    this.dropZone.addEventListener('drop', (e) => {
      e.preventDefault();
      this.dropZone.classList.remove('dragover');
      this.handleDroppedFiles(e.dataTransfer.files);
    });
    
    // 按钮选择
    this.selectFilesBtn.addEventListener('click', () => this.selectFiles());
    this.selectFolderBtn.addEventListener('click', () => this.selectFolder());
    
    // 操作面板开关
    this.enableResize.addEventListener('change', () => {
      this.resizeContent.style.display = this.enableResize.checked ? 'block' : 'none';
    });
    
    this.enableFormat.addEventListener('change', () => {
      this.formatContent.style.display = this.enableFormat.checked ? 'block' : 'none';
    });
    
    this.enableRename.addEventListener('change', () => {
      this.renameContent.style.display = this.enableRename.checked ? 'block' : 'none';
      if (this.enableRename.checked && this.renameTemplate.value) {
        this.validateRenameTemplate();
      }
    });
    
    // 尺寸调整模式切换
    this.resizeModeRadios.forEach(radio => {
      radio.addEventListener('change', () => {
        if (radio.value === 'pixels') {
          this.resizePixels.style.display = 'flex';
          this.resizePercent.style.display = 'none';
        } else {
          this.resizePixels.style.display = 'none';
          this.resizePercent.style.display = 'block';
        }
      });
    });
    
    // 宽高比锁定
    this.lockAspect.addEventListener('change', () => {
      if (this.lockAspect.checked) {
        this.syncAspectRatio();
      }
    });
    
    this.resizeWidth.addEventListener('input', () => {
      if (this.lockAspect.checked) {
        this.syncHeight();
      }
    });
    
    this.resizeHeight.addEventListener('input', () => {
      if (this.lockAspect.checked) {
        this.syncWidth();
      }
    });
    
    // 重命名模板验证
    this.renameTemplate.addEventListener('input', () => {
      this.validateRenameTemplate();
    });
    
    // 执行和清空
    this.executeBtn.addEventListener('click', () => this.executeProcessing());
    this.clearBtn.addEventListener('click', () => this.clearAll());
    
    // 结果操作
    this.openFolderBtn.addEventListener('click', () => this.openOutputFolder());
    this.showComparisonBtn.addEventListener('click', () => this.showComparison());
    this.closeComparisonBtn.addEventListener('click', () => this.hideComparison());
    
    // 对比导航
    this.prevComparison.addEventListener('click', () => this.prevComparisonImage());
    this.nextComparison.addEventListener('click', () => this.nextComparisonImage());
  }
  
  async handleDroppedFiles(files) {
    const filePaths = [];
    
    for (const file of files) {
      if (file.path) {
        filePaths.push(file.path);
      }
    }
    
    if (filePaths.length > 0) {
      await this.loadImages(filePaths);
    }
  }
  
  async selectFiles() {
    const result = await window.electronAPI.selectFiles();
    if (result.files.length > 0) {
      await this.loadImages(result.files);
    }
  }
  
  async selectFolder() {
    const result = await window.electronAPI.selectFolder();
    if (result.files.length > 0) {
      await this.loadImages(result.files);
    }
  }
  
  async loadImages(filePaths) {
    this.images = [...this.images, ...filePaths];
    this.images = [...new Set(this.images)]; // 去重
    
    await this.renderPreviewGrid();
    this.showOperationsPanel();
  }
  
  async renderPreviewGrid() {
    this.previewGrid.innerHTML = '';
    this.imageCount.textContent = this.images.length;
    
    for (const filePath of this.images) {
      const result = await window.electronAPI.getThumbnail(filePath);
      
      const item = document.createElement('div');
      item.className = 'preview-item fade-in';
      
      if (result.success) {
        item.innerHTML = `
          <img class="thumbnail" src="${result.data}" alt="${result.name}">
          <div class="filename">${result.name}</div>
        `;
      } else {
        item.innerHTML = `
          <div class="thumbnail" style="display: flex; align-items: center; justify-content: center; color: var(--danger-color);">
            无法读取
          </div>
          <div class="filename" style="color: var(--danger-color);">${result.name || '未知'}</div>
        `;
      }
      
      this.previewGrid.appendChild(item);
    }
  }
  
  showOperationsPanel() {
    this.operationsPanel.style.display = 'block';
    this.previewSection.style.display = 'block';
  }
  
  syncAspectRatio() {
    // 这个功能需要知道原始图片的尺寸，
    // 这里简化处理，实际应用中可能需要从后端获取
  }
  
  syncHeight() {
    // 简化实现，实际需要根据原始图片比例计算
  }
  
  syncWidth() {
    // 简化实现，实际需要根据原始图片比例计算
  }
  
  async validateRenameTemplate() {
    const template = this.renameTemplate.value.trim();
    
    if (!template) {
      this.previewValue.textContent = '请输入模板';
      this.templateError.style.display = 'none';
      return;
    }
    
    const result = await window.electronAPI.validateRenameTemplate(template);
    
    if (result.valid) {
      const preview = this.previewRename(template);
      this.previewValue.textContent = preview;
      this.templateError.style.display = 'none';
    } else {
      this.previewValue.textContent = '';
      this.templateError.textContent = `错误：无效的变量 ${result.invalidVariables.join(', ')}`;
      this.templateError.style.display = 'block';
    }
  }
  
  previewRename(template) {
    const date = new Date().toISOString().slice(0, 10).replace(/-/g, '');
    return template
      .replace('{name}', 'image')
      .replace('{index}', '001')
      .replace('{date}', date)
      .replace('{ext}', '.jpg');
  }
  
  async executeProcessing() {
    if (this.images.length === 0) {
      alert('请先导入图片！');
      return;
    }
    
    const operations = {};
    
    // 检查是否启用了任何操作
    if (!this.enableResize.checked && !this.enableFormat.checked && !this.enableRename.checked) {
      alert('请至少选择一个操作！');
      return;
    }
    
    // 验证重命名模板
    if (this.enableRename.checked) {
      const template = this.renameTemplate.value.trim();
      if (!template) {
        alert('请输入重命名模板！');
        return;
      }
      
      const validation = await window.electronAPI.validateRenameTemplate(template);
      if (!validation.valid) {
        alert(`重命名模板错误：无效的变量 ${validation.invalidVariables.join(', ')}`);
        return;
      }
    }
    
    // 收集操作参数
    if (this.enableResize.checked) {
      const resizeMode = document.querySelector('input[name="resizeMode"]:checked').value;
      operations.resize = {
        mode: resizeMode,
        lockAspect: this.lockAspect.checked
      };
      
      if (resizeMode === 'pixels') {
        operations.resize.width = this.resizeWidth.value ? parseInt(this.resizeWidth.value) : null;
        operations.resize.height = this.resizeHeight.value ? parseInt(this.resizeHeight.value) : null;
      } else {
        operations.resize.percent = parseInt(this.resizePercentValue.value) || 100;
      }
    }
    
    if (this.enableFormat.checked) {
      operations.format = this.targetFormat.value;
    }
    
    if (this.enableRename.checked) {
      operations.rename = {
        template: this.renameTemplate.value.trim()
      };
    }
    
    // 显示进度
    this.showProgress();
    
    // 注册进度监听器
    window.electronAPI.onProcessingProgress((data) => {
      this.updateProgress(data);
    });
    
    try {
      const result = await window.electronAPI.processImages({
        images: this.images,
        operations,
        outputDir: null // 使用默认的 processed 目录
      });
      
      this.processedFiles = result.processedFiles;
      this.lastResults = result.results;
      
      // 保存输出目录
      if (result.results.success.length > 0) {
        const processedPath = result.results.success[0].processed;
        // 使用字符串操作获取目录名，避免在渲染进程中使用 require
        const lastSeparator = processedPath.lastIndexOf('/') !== -1 ? 
                             processedPath.lastIndexOf('/') : 
                             processedPath.lastIndexOf('\\');
        this.outputDirectory = processedPath.substring(0, lastSeparator);
      }
      
      // 显示结果
      this.hideProgress();
      this.showResults(result.results);
      
    } catch (error) {
      console.error('处理失败:', error);
      alert('处理过程中发生错误：' + error.message);
      this.hideProgress();
    }
    
    // 移除监听器
    window.electronAPI.removeAllListeners();
  }
  
  showProgress() {
    this.progressSection.style.display = 'block';
    this.progressFill.style.width = '0%';
    this.progressPercent.textContent = '0%';
    this.progressText.textContent = '准备处理...';
  }
  
  updateProgress(data) {
    this.progressFill.style.width = `${data.percentage}%`;
    this.progressPercent.textContent = `${data.percentage}%`;
    this.progressText.textContent = `处理中... ${data.current} / ${data.total}`;
  }
  
  hideProgress() {
    this.progressSection.style.display = 'none';
  }
  
  showResults(results) {
    this.resultsSection.style.display = 'block';
    
    // 更新计数
    this.successCount.textContent = results.success.length;
    this.skippedCount.textContent = results.skipped.length;
    this.failedCount.textContent = results.failed.length;
    
    // 更新成功列表
    const successUl = this.successList.querySelector('ul');
    successUl.innerHTML = '';
    results.success.forEach(item => {
      const li = document.createElement('li');
      li.textContent = item.name;
      successUl.appendChild(li);
    });
    this.successList.querySelector('h4').textContent = `成功处理 (${results.success.length})`;
    
    // 更新跳过列表
    if (results.skipped.length > 0) {
      this.skippedList.style.display = 'block';
      const skippedUl = this.skippedList.querySelector('ul');
      skippedUl.innerHTML = '';
      results.skipped.forEach(item => {
        const li = document.createElement('li');
        li.textContent = `${item.name}: ${item.error}`;
        skippedUl.appendChild(li);
      });
      this.skippedList.querySelector('h4').textContent = `跳过 (${results.skipped.length})`;
    } else {
      this.skippedList.style.display = 'none';
    }
    
    // 更新失败列表
    if (results.failed.length > 0) {
      this.failedList.style.display = 'block';
      const failedUl = this.failedList.querySelector('ul');
      failedUl.innerHTML = '';
      results.failed.forEach(item => {
        const li = document.createElement('li');
        li.textContent = `${item.name}: ${item.error}`;
        failedUl.appendChild(li);
      });
      this.failedList.querySelector('h4').textContent = `失败 (${results.failed.length})`;
    } else {
      this.failedList.style.display = 'none';
    }
    
    // 滚动到结果区域
    this.resultsSection.scrollIntoView({ behavior: 'smooth' });
  }
  
  clearAll() {
    if (confirm('确定要清空所有导入的图片吗？')) {
      this.images = [];
      this.processedFiles = [];
      this.outputDirectory = null;
      this.lastResults = null;
      
      this.previewGrid.innerHTML = '';
      this.imageCount.textContent = '0';
      
      this.operationsPanel.style.display = 'none';
      this.previewSection.style.display = 'none';
      this.resultsSection.style.display = 'none';
      this.comparisonSection.style.display = 'none';
      
      // 重置表单
      this.enableResize.checked = false;
      this.enableFormat.checked = false;
      this.enableRename.checked = false;
      
      this.resizeContent.style.display = 'none';
      this.formatContent.style.display = 'none';
      this.renameContent.style.display = 'none';
      
      this.resizeWidth.value = '';
      this.resizeHeight.value = '';
      this.resizePercentValue.value = '100';
      this.renameTemplate.value = '';
      this.templateError.style.display = 'none';
      this.previewValue.textContent = 'image_001.jpg';
    }
  }
  
  async openOutputFolder() {
    if (this.outputDirectory) {
      await window.electronAPI.openOutputFolder(this.outputDirectory);
    } else {
      alert('没有可用的输出目录');
    }
  }
  
  async showComparison() {
    if (this.processedFiles.length === 0) {
      alert('没有可对比的图片');
      return;
    }
    
    this.currentComparisonIndex = 0;
    this.comparisonSection.style.display = 'block';
    this.updateComparisonNav();
    await this.loadComparisonImage(0);
    
    this.comparisonSection.scrollIntoView({ behavior: 'smooth' });
  }
  
  hideComparison() {
    this.comparisonSection.style.display = 'none';
  }
  
  async loadComparisonImage(index) {
    if (index < 0 || index >= this.processedFiles.length) return;
    
    const file = this.processedFiles[index];
    
    try {
      const result = await window.electronAPI.getComparisonImages(file.original, file.processed);
      
      if (result.success) {
        this.comparisonOriginal.src = result.original;
        this.comparisonProcessed.src = result.processed;
      } else {
        console.error('加载对比图片失败:', result.error);
      }
    } catch (error) {
      console.error('加载对比图片出错:', error);
    }
  }
  
  updateComparisonNav() {
    const total = this.processedFiles.length;
    const current = this.currentComparisonIndex + 1;
    
    this.comparisonIndex.textContent = `${current} / ${total}`;
    this.prevComparison.disabled = this.currentComparisonIndex === 0;
    this.nextComparison.disabled = this.currentComparisonIndex >= total - 1;
  }
  
  async prevComparisonImage() {
    if (this.currentComparisonIndex > 0) {
      this.currentComparisonIndex--;
      this.updateComparisonNav();
      await this.loadComparisonImage(this.currentComparisonIndex);
    }
  }
  
  async nextComparisonImage() {
    if (this.currentComparisonIndex < this.processedFiles.length - 1) {
      this.currentComparisonIndex++;
      this.updateComparisonNav();
      await this.loadComparisonImage(this.currentComparisonIndex);
    }
  }
}

// 初始化应用
document.addEventListener('DOMContentLoaded', () => {
  new ImageBatchProcessor();
});
