class APIDebugger {
    constructor() {
        this.currentRequest = {
            method: 'GET',
            url: '',
            headers: {},
            body: '',
            bodyType: 'none',
            files: []
        };
        
        this.history = [];
        this.environments = {
            active: 'default',
            environments: {
                default: {
                    name: '默认环境',
                    variables: {}
                }
            }
        };
        this.settings = {
            timeout: 30000
        };
        
        this.selectedEnvId = 'default';
        this.editingEnvId = null;
        this.searchQuery = '';
        
        this.init();
    }

    async init() {
        await this.loadData();
        this.bindEvents();
        this.renderEnvironmentSelector();
        this.renderHistory();
        this.addDefaultHeaderRow();
    }

    async loadData() {
        try {
            this.history = await window.electronAPI.getHistory();
            this.environments = await window.electronAPI.getEnvironments();
            this.settings = await window.electronAPI.getSettings();
            
            if (this.environments.active) {
                this.selectedEnvId = this.environments.active;
            }
        } catch (error) {
            console.error('Error loading data:', error);
        }
    }

    bindEvents() {
        document.getElementById('method-select').addEventListener('change', (e) => {
            this.currentRequest.method = e.target.value;
        });

        document.getElementById('url-input').addEventListener('input', (e) => {
            this.currentRequest.url = e.target.value;
        });

        document.getElementById('url-input').addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                this.sendRequest();
            }
        });

        document.getElementById('send-btn').addEventListener('click', () => this.sendRequest());
        document.getElementById('save-btn').addEventListener('click', () => this.saveRequestToHistory());

        document.querySelectorAll('.request-tabs .tab').forEach(tab => {
            tab.addEventListener('click', (e) => {
                this.switchTab(e.target.dataset.tab);
            });
        });

        document.querySelectorAll('.response-tabs .tab').forEach(tab => {
            tab.addEventListener('click', (e) => {
                this.switchResponseTab(e.target.dataset.responseTab);
            });
        });

        document.getElementById('add-header-btn').addEventListener('click', () => this.addHeaderRow());

        document.querySelectorAll('input[name="body-type"]').forEach(radio => {
            radio.addEventListener('change', (e) => {
                this.currentRequest.bodyType = e.target.value;
                this.renderBodyEditor();
            });
        });

        document.getElementById('environment-selector').addEventListener('change', (e) => {
            this.selectedEnvId = e.target.value;
            this.environments.active = e.target.value;
            this.saveEnvironments();
        });

        document.getElementById('manage-envs-btn').addEventListener('click', () => {
            this.openEnvironmentsModal();
        });

        document.getElementById('close-envs-modal').addEventListener('click', () => {
            this.closeEnvironmentsModal();
        });

        document.getElementById('cancel-envs-btn').addEventListener('click', () => {
            this.closeEnvironmentsModal();
        });

        document.getElementById('save-envs-btn').addEventListener('click', () => {
            this.saveCurrentEnvEditing();
            this.saveEnvironments();
            this.renderEnvironmentSelector();
            this.closeEnvironmentsModal();
        });

        document.getElementById('add-env-btn').addEventListener('click', () => {
            this.addNewEnvironment();
        });

        document.getElementById('delete-env-btn').addEventListener('click', () => {
            this.deleteCurrentEnvironment();
        });

        document.getElementById('add-env-var-btn').addEventListener('click', () => {
            this.addEnvVarRow();
        });

        document.getElementById('settings-btn').addEventListener('click', () => {
            this.openSettingsModal();
        });

        document.getElementById('close-settings-modal').addEventListener('click', () => {
            this.closeSettingsModal();
        });

        document.getElementById('cancel-settings-btn').addEventListener('click', () => {
            this.closeSettingsModal();
        });

        document.getElementById('save-settings-btn').addEventListener('click', () => {
            this.saveSettings();
            this.closeSettingsModal();
        });

        document.getElementById('clear-history-btn').addEventListener('click', () => {
            if (confirm('确定要清除所有历史记录吗？此操作不可撤销。')) {
                this.history = [];
                this.saveHistory();
                this.renderHistory();
                this.closeSettingsModal();
            }
        });

        document.getElementById('history-search').addEventListener('input', (e) => {
            this.searchQuery = e.target.value.toLowerCase();
            this.renderHistory();
        });

        document.getElementById('clear-search-btn').addEventListener('click', () => {
            document.getElementById('history-search').value = '';
            this.searchQuery = '';
            this.renderHistory();
        });

        document.querySelectorAll('.modal').forEach(modal => {
            modal.addEventListener('click', (e) => {
                if (e.target === modal) {
                    modal.classList.remove('active');
                }
            });
        });
    }

    switchTab(tabName) {
        document.querySelectorAll('.request-tabs .tab').forEach(tab => {
            tab.classList.remove('active');
            if (tab.dataset.tab === tabName) {
                tab.classList.add('active');
            }
        });

        document.querySelectorAll('.tab-panel').forEach(panel => {
            panel.classList.remove('active');
        });
        document.getElementById(tabName + '-panel').classList.add('active');
    }

    switchResponseTab(tabName) {
        document.querySelectorAll('.response-tabs .tab').forEach(tab => {
            tab.classList.remove('active');
            if (tab.dataset.responseTab === tabName) {
                tab.classList.add('active');
            }
        });

        document.querySelectorAll('.response-tab-panel').forEach(panel => {
            panel.classList.remove('active');
        });
        document.getElementById(tabName + '-panel').classList.add('active');
    }

    addDefaultHeaderRow() {
        this.addHeaderRow('Content-Type', 'application/json');
        this.addHeaderRow();
    }

    addHeaderRow(key = '', value = '') {
        const tbody = document.getElementById('headers-table-body');
        const row = document.createElement('tr');
        
        row.innerHTML = `
            <td><input type="text" class="header-key" value="${this.escapeHtml(key)}" placeholder="Header 名称"></td>
            <td><input type="text" class="header-value" value="${this.escapeHtml(value)}" placeholder="Header 值"></td>
            <td><button class="remove-btn">-</button></td>
        `;

        row.querySelector('.remove-btn').addEventListener('click', () => {
            row.remove();
        });

        tbody.appendChild(row);
        return row;
    }

    getCurrentHeaders() {
        const headers = {};
        const rows = document.querySelectorAll('#headers-table-body tr');
        
        rows.forEach(row => {
            const keyInput = row.querySelector('.header-key');
            const valueInput = row.querySelector('.header-value');
            
            if (keyInput && valueInput && keyInput.value.trim()) {
                headers[keyInput.value.trim()] = valueInput.value;
            }
        });

        return headers;
    }

    setHeaders(headers) {
        const tbody = document.getElementById('headers-table-body');
        tbody.innerHTML = '';

        if (Object.keys(headers).length === 0) {
            this.addDefaultHeaderRow();
            return;
        }

        Object.keys(headers).forEach(key => {
            this.addHeaderRow(key, headers[key]);
        });
        this.addHeaderRow();
    }

    renderBodyEditor() {
        const bodyContent = document.getElementById('body-content');
        const bodyType = this.currentRequest.bodyType;

        bodyContent.innerHTML = '';

        if (bodyType === 'none') {
            bodyContent.innerHTML = '<p class="body-hint">选择 Body 类型以继续</p>';
            return;
        }

        if (bodyType === 'json') {
            const container = document.createElement('div');
            container.innerHTML = `
                <textarea class="body-textarea" id="body-json-input" placeholder='{"key": "value"}'>${this.escapeHtml(this.currentRequest.body || '')}</textarea>
                <div class="json-error" id="json-error" style="display: none;"></div>
            `;
            bodyContent.appendChild(container);

            const textarea = container.querySelector('.body-textarea');
            textarea.addEventListener('input', (e) => {
                this.currentRequest.body = e.target.value;
                this.validateJSON(e.target.value);
            });

            this.validateJSON(this.currentRequest.body);
        } else if (bodyType === 'form-urlencoded') {
            const container = document.createElement('div');
            container.className = 'form-data-section';
            container.innerHTML = `
                <table class="key-value-table">
                    <thead>
                        <tr>
                            <th>Key</th>
                            <th>Value</th>
                            <th style="width: 60px;">操作</th>
                        </tr>
                    </thead>
                    <tbody id="form-body-table-body">
                    </tbody>
                </table>
                <button id="add-form-field-btn" class="btn btn-secondary">+ 添加字段</button>
            `;
            bodyContent.appendChild(container);

            document.getElementById('add-form-field-btn').addEventListener('click', () => {
                this.addFormFieldRow();
            });

            this.setFormFields(this.currentRequest.body ? JSON.parse(this.currentRequest.body) : {});
        } else if (bodyType === 'form-data') {
            const container = document.createElement('div');
            container.className = 'form-data-section';
            container.innerHTML = `
                <table class="key-value-table">
                    <thead>
                        <tr>
                            <th>Key</th>
                            <th>Value</th>
                            <th style="width: 60px;">操作</th>
                        </tr>
                    </thead>
                    <tbody id="form-data-fields-body">
                    </tbody>
                </table>
                <button id="add-form-data-field-btn" class="btn btn-secondary">+ 添加字段</button>
                
                <h4 style="margin: 20px 0 10px; color: #858585;">文件上传</h4>
                <div id="file-uploads-container">
                </div>
                <button id="add-file-btn" class="btn btn-secondary add-file-btn">+ 添加文件</button>
            `;
            bodyContent.appendChild(container);

            document.getElementById('add-form-data-field-btn').addEventListener('click', () => {
                this.addFormDataFieldRow();
            });

            document.getElementById('add-file-btn').addEventListener('click', () => {
                this.addFileUploadRow();
            });

            this.setFormDataFields(this.currentRequest.body ? JSON.parse(this.currentRequest.body) : {});
            this.setFileUploads(this.currentRequest.files || []);
        } else {
            const container = document.createElement('div');
            container.innerHTML = `
                <textarea class="body-textarea" id="body-raw-input" placeholder="输入请求体内容...">${this.escapeHtml(this.currentRequest.body || '')}</textarea>
            `;
            bodyContent.appendChild(container);

            const textarea = container.querySelector('.body-textarea');
            textarea.addEventListener('input', (e) => {
                this.currentRequest.body = e.target.value;
            });
        }
    }

    validateJSON(jsonString) {
        const errorDiv = document.getElementById('json-error');
        const textarea = document.getElementById('body-json-input');
        
        if (!errorDiv || !textarea) return;

        if (!jsonString.trim()) {
            errorDiv.style.display = 'none';
            textarea.classList.remove('error');
            return;
        }

        try {
            JSON.parse(jsonString);
            errorDiv.style.display = 'none';
            textarea.classList.remove('error');
        } catch (e) {
            errorDiv.textContent = 'JSON 格式错误: ' + e.message;
            errorDiv.style.display = 'block';
            textarea.classList.add('error');
        }
    }

    addFormFieldRow(key = '', value = '') {
        const tbody = document.getElementById('form-body-table-body');
        if (!tbody) return;
        
        const row = document.createElement('tr');
        row.innerHTML = `
            <td><input type="text" class="form-field-key" value="${this.escapeHtml(key)}" placeholder="字段名"></td>
            <td><input type="text" class="form-field-value" value="${this.escapeHtml(value)}" placeholder="值"></td>
            <td><button class="remove-btn">-</button></td>
        `;

        row.querySelector('.remove-btn').addEventListener('click', () => {
            row.remove();
        });

        tbody.appendChild(row);
    }

    setFormFields(fields) {
        const tbody = document.getElementById('form-body-table-body');
        if (!tbody) return;
        tbody.innerHTML = '';

        if (Object.keys(fields).length === 0) {
            this.addFormFieldRow();
            this.addFormFieldRow();
            return;
        }

        Object.keys(fields).forEach(key => {
            this.addFormFieldRow(key, fields[key]);
        });
        this.addFormFieldRow();
    }

    getFormFields() {
        const fields = {};
        const rows = document.querySelectorAll('#form-body-table-body tr');
        
        rows.forEach(row => {
            const keyInput = row.querySelector('.form-field-key');
            const valueInput = row.querySelector('.form-field-value');
            
            if (keyInput && valueInput && keyInput.value.trim()) {
                fields[keyInput.value.trim()] = valueInput.value;
            }
        });

        return fields;
    }

    addFormDataFieldRow(key = '', value = '') {
        const tbody = document.getElementById('form-data-fields-body');
        if (!tbody) return;
        
        const row = document.createElement('tr');
        row.innerHTML = `
            <td><input type="text" class="form-data-field-key" value="${this.escapeHtml(key)}" placeholder="字段名"></td>
            <td><input type="text" class="form-data-field-value" value="${this.escapeHtml(value)}" placeholder="值"></td>
            <td><button class="remove-btn">-</button></td>
        `;

        row.querySelector('.remove-btn').addEventListener('click', () => {
            row.remove();
        });

        tbody.appendChild(row);
    }

    setFormDataFields(fields) {
        const tbody = document.getElementById('form-data-fields-body');
        if (!tbody) return;
        tbody.innerHTML = '';

        if (Object.keys(fields).length === 0) {
            this.addFormDataFieldRow();
            this.addFormDataFieldRow();
            return;
        }

        Object.keys(fields).forEach(key => {
            this.addFormDataFieldRow(key, fields[key]);
        });
        this.addFormDataFieldRow();
    }

    getFormDataFields() {
        const fields = {};
        const rows = document.querySelectorAll('#form-data-fields-body tr');
        
        rows.forEach(row => {
            const keyInput = row.querySelector('.form-data-field-key');
            const valueInput = row.querySelector('.form-data-field-value');
            
            if (keyInput && valueInput && keyInput.value.trim()) {
                fields[keyInput.value.trim()] = valueInput.value;
            }
        });

        return fields;
    }

    addFileUploadRow(fileInfo = null) {
        const container = document.getElementById('file-uploads-container');
        if (!container) return;

        const item = document.createElement('div');
        item.className = 'file-upload-item';
        item.innerHTML = `
            <input type="text" class="file-field-name" value="${fileInfo ? this.escapeHtml(fileInfo.fieldName) : 'file'}" placeholder="字段名">
            <span class="file-name">${fileInfo ? this.escapeHtml(fileInfo.filename) : '未选择文件'}</span>
            <button class="select-file-btn">选择文件</button>
            <button class="remove-file-btn">-</button>
        `;

        if (fileInfo) {
            item.dataset.filePath = fileInfo.path;
            item.dataset.filename = fileInfo.filename;
        }

        item.querySelector('.select-file-btn').addEventListener('click', async () => {
            const files = await window.electronAPI.selectFiles();
            if (files && files.length > 0) {
                const filePath = files[0];
                const fileInfo = await window.electronAPI.getFileInfo(filePath);
                
                if (fileInfo.success) {
                    item.dataset.filePath = filePath;
                    item.dataset.filename = fileInfo.filename;
                    item.querySelector('.file-name').textContent = fileInfo.filename;
                }
            }
        });

        item.querySelector('.remove-file-btn').addEventListener('click', () => {
            item.remove();
        });

        container.appendChild(item);
    }

    setFileUploads(files) {
        const container = document.getElementById('file-uploads-container');
        if (!container) return;
        container.innerHTML = '';

        if (files.length === 0) {
            this.addFileUploadRow();
            return;
        }

        files.forEach(fileInfo => {
            this.addFileUploadRow(fileInfo);
        });
    }

    getFileUploads() {
        const files = [];
        const items = document.querySelectorAll('.file-upload-item');
        
        items.forEach(item => {
            const fieldName = item.querySelector('.file-field-name').value.trim();
            const filePath = item.dataset.filePath;
            const filename = item.dataset.filename;
            
            if (fieldName && filePath) {
                files.push({
                    fieldName: fieldName,
                    path: filePath,
                    filename: filename
                });
            }
        });

        return files;
    }

    getActiveEnvironmentVariables() {
        const env = this.environments.environments[this.selectedEnvId];
        return env ? env.variables : {};
    }

    replaceEnvironmentVariables(text) {
        if (!text) return text;
        
        const variables = this.getActiveEnvironmentVariables();
        
        return text.replace(/\{\{(\w+)\}\}/g, (match, key) => {
            if (variables.hasOwnProperty(key)) {
                return variables[key];
            }
            return match;
        });
    }

    async sendRequest() {
        const method = this.currentRequest.method;
        let url = document.getElementById('url-input').value.trim();

        if (!url) {
            alert('请输入请求 URL');
            return;
        }

        url = this.replaceEnvironmentVariables(url);

        let requestBody = null;
        let files = [];

        if (this.currentRequest.bodyType === 'json') {
            requestBody = this.currentRequest.body;
        } else if (this.currentRequest.bodyType === 'form-urlencoded') {
            const fields = this.getFormFields();
            this.currentRequest.body = JSON.stringify(fields);
            const params = new URLSearchParams();
            Object.keys(fields).forEach(key => {
                params.append(key, fields[key]);
            });
            requestBody = params.toString();
        } else if (this.currentRequest.bodyType === 'form-data') {
            const fields = this.getFormDataFields();
            this.currentRequest.body = JSON.stringify(fields);
            requestBody = JSON.stringify(fields);
            files = this.getFileUploads();
            this.currentRequest.files = files;
        } else if (this.currentRequest.bodyType === 'raw') {
            requestBody = this.currentRequest.body;
        }

        const headers = {};
        const rawHeaders = this.getCurrentHeaders();
        Object.keys(rawHeaders).forEach(key => {
            headers[key] = this.replaceEnvironmentVariables(rawHeaders[key]);
        });

        const sendBtn = document.getElementById('send-btn');
        const originalText = sendBtn.textContent;
        sendBtn.textContent = '发送中...';
        sendBtn.disabled = true;

        try {
            const result = await window.electronAPI.sendRequest({
                method: method,
                url: url,
                headers: headers,
                body: requestBody,
                bodyType: this.currentRequest.bodyType,
                files: files,
                timeout: this.settings.timeout
            });

            this.displayResponse(result);

            if (result.success) {
                this.addToHistory({
                    method: method,
                    url: document.getElementById('url-input').value,
                    headers: this.getCurrentHeaders(),
                    body: this.currentRequest.body,
                    bodyType: this.currentRequest.bodyType,
                    files: files,
                    statusCode: result.statusCode
                });
            }

        } catch (error) {
            this.displayResponse({
                success: false,
                error: error.message
            });
        } finally {
            sendBtn.textContent = originalText;
            sendBtn.disabled = false;
        }
    }

    displayResponse(result) {
        const responseStatus = document.getElementById('response-status');
        const responseTime = document.getElementById('response-time');
        const responseSize = document.getElementById('response-size');
        const responseBody = document.getElementById('response-body');
        const responseHeaders = document.getElementById('response-headers');
        const truncatedNotice = document.getElementById('response-truncated');

        if (result.success) {
            responseStatus.textContent = `${result.statusCode} ${result.statusMessage}`;
            responseStatus.className = 'response-status ' + this.getStatusClass(result.statusCode);
            
            responseTime.textContent = `${result.responseTime}ms`;
            responseSize.textContent = this.formatBytes(result.bodySize);

            if (result.isTruncated) {
                truncatedNotice.style.display = 'block';
            } else {
                truncatedNotice.style.display = 'none';
            }

            let formattedBody = result.body;
            let isJSON = false;
            
            try {
                const parsed = JSON.parse(result.body);
                formattedBody = JSON.stringify(parsed, null, 2);
                isJSON = true;
            } catch (e) {
                isJSON = false;
            }

            if (isJSON) {
                responseBody.innerHTML = `<code>${this.syntaxHighlightJSON(formattedBody)}</code>`;
            } else {
                responseBody.innerHTML = `<code>${this.escapeHtml(formattedBody)}</code>`;
            }

            const headersHtml = Object.keys(result.headers).map(key => {
                return `<div class="response-headers-item">
                    <span class="response-headers-key">${this.escapeHtml(key)}:</span>
                    <span class="response-headers-value">${this.escapeHtml(String(result.headers[key]))}</span>
                </div>`;
            }).join('');
            
            responseHeaders.innerHTML = headersHtml || '<p>-</p>';
        } else {
            responseStatus.textContent = result.error || '请求失败';
            responseStatus.className = 'response-status error';
            responseTime.textContent = '-';
            responseSize.textContent = '-';
            truncatedNotice.style.display = 'none';
            
            responseBody.innerHTML = `<code>${this.escapeHtml(result.error || '请求失败')}</code>`;
            responseHeaders.innerHTML = '<p>-</p>';
        }
    }

    getStatusClass(statusCode) {
        if (statusCode >= 200 && statusCode < 300) return 'success';
        if (statusCode >= 300 && statusCode < 400) return 'redirect';
        if (statusCode >= 400 && statusCode < 500) return 'error';
        if (statusCode >= 500) return 'error';
        return 'info';
    }

    formatBytes(bytes) {
        if (bytes < 1024) return bytes + ' B';
        if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
        return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    }

    syntaxHighlightJSON(json) {
        json = this.escapeHtml(json);
        return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
            let cls = 'json-number';
            if (/^"/.test(match)) {
                if (/:$/.test(match)) {
                    cls = 'json-key';
                } else {
                    cls = 'json-string';
                }
            } else if (/true|false/.test(match)) {
                cls = 'json-boolean';
            } else if (/null/.test(match)) {
                cls = 'json-null';
            }
            return '<span class="' + cls + '">' + match + '</span>';
        });
    }

    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    addToHistory(requestData) {
        const historyItem = {
            id: Date.now().toString(),
            method: requestData.method,
            url: requestData.url,
            headers: requestData.headers,
            body: requestData.body,
            bodyType: requestData.bodyType,
            files: requestData.files || [],
            statusCode: requestData.statusCode,
            timestamp: new Date().toISOString(),
            favorited: false
        };

        this.history.unshift(historyItem);
        
        if (this.history.length > 100) {
            this.history = this.history.slice(0, 100);
        }

        this.saveHistory();
        this.renderHistory();
    }

    saveRequestToHistory() {
        const method = this.currentRequest.method;
        const url = document.getElementById('url-input').value.trim();

        if (!url) {
            alert('请输入请求 URL');
            return;
        }

        let body = '';
        let files = [];

        if (this.currentRequest.bodyType === 'form-urlencoded') {
            body = JSON.stringify(this.getFormFields());
        } else if (this.currentRequest.bodyType === 'form-data') {
            body = JSON.stringify(this.getFormDataFields());
            files = this.getFileUploads();
        } else {
            body = this.currentRequest.body;
        }

        this.addToHistory({
            method: method,
            url: url,
            headers: this.getCurrentHeaders(),
            body: body,
            bodyType: this.currentRequest.bodyType,
            files: files,
            statusCode: null
        });

        alert('请求已保存到历史记录');
    }

    renderHistory() {
        const historyList = document.getElementById('history-list');
        historyList.innerHTML = '';

        let filteredHistory = this.history;
        
        if (this.searchQuery) {
            filteredHistory = this.history.filter(item => {
                return item.url.toLowerCase().includes(this.searchQuery) ||
                       item.method.toLowerCase().includes(this.searchQuery);
            });
        }

        filteredHistory.sort((a, b) => {
            if (a.favorited && !b.favorited) return -1;
            if (!a.favorited && b.favorited) return 1;
            return new Date(b.timestamp) - new Date(a.timestamp);
        });

        if (filteredHistory.length === 0) {
            historyList.innerHTML = '<div class="empty-history">暂无历史记录</div>';
            return;
        }

        filteredHistory.forEach(item => {
            const historyItem = document.createElement('div');
            historyItem.className = 'history-item';
            historyItem.dataset.id = item.id;

            const time = new Date(item.timestamp);
            const timeStr = time.toLocaleString('zh-CN', {
                month: '2-digit',
                day: '2-digit',
                hour: '2-digit',
                minute: '2-digit'
            });

            historyItem.innerHTML = `
                <div class="history-item-actions">
                    <button class="history-item-action-btn favorite-btn ${item.favorited ? 'favorited' : ''}" title="收藏">★</button>
                    <button class="history-item-action-btn delete-btn" title="删除">×</button>
                </div>
                <span class="history-item-method method-${item.method}">${item.method}</span>
                <span class="history-item-url">${this.escapeHtml(item.url)}</span>
                <div class="history-item-time">${timeStr}</div>
            `;

            historyItem.addEventListener('click', (e) => {
                if (e.target.closest('.history-item-actions')) return;
                this.loadRequestFromHistory(item);
            });

            historyItem.querySelector('.favorite-btn').addEventListener('click', (e) => {
                e.stopPropagation();
                this.toggleFavorite(item.id);
            });

            historyItem.querySelector('.delete-btn').addEventListener('click', (e) => {
                e.stopPropagation();
                this.deleteHistoryItem(item.id);
            });

            historyList.appendChild(historyItem);
        });
    }

    loadRequestFromHistory(item) {
        document.getElementById('method-select').value = item.method;
        document.getElementById('url-input').value = item.url;
        
        this.currentRequest.method = item.method;
        this.currentRequest.url = item.url;
        this.currentRequest.body = item.body || '';
        this.currentRequest.bodyType = item.bodyType || 'none';
        this.currentRequest.files = item.files || [];

        this.setHeaders(item.headers || {});

        document.querySelectorAll('input[name="body-type"]').forEach(radio => {
            radio.checked = (radio.value === this.currentRequest.bodyType);
        });

        this.renderBodyEditor();

        document.querySelectorAll('.history-item').forEach(el => {
            el.classList.remove('active');
            if (el.dataset.id === item.id) {
                el.classList.add('active');
            }
        });
    }

    toggleFavorite(id) {
        const item = this.history.find(h => h.id === id);
        if (item) {
            item.favorited = !item.favorited;
            this.saveHistory();
            this.renderHistory();
        }
    }

    deleteHistoryItem(id) {
        if (confirm('确定要删除这条历史记录吗？')) {
            this.history = this.history.filter(h => h.id !== id);
            this.saveHistory();
            this.renderHistory();
        }
    }

    async saveHistory() {
        await window.electronAPI.saveHistory(this.history);
    }

    renderEnvironmentSelector() {
        const selector = document.getElementById('environment-selector');
        selector.innerHTML = '';

        Object.keys(this.environments.environments).forEach(envId => {
            const env = this.environments.environments[envId];
            const option = document.createElement('option');
            option.value = envId;
            option.textContent = env.name;
            option.selected = (envId === this.selectedEnvId);
            selector.appendChild(option);
        });
    }

    openEnvironmentsModal() {
        const modal = document.getElementById('environments-modal');
        modal.classList.add('active');

        this.renderEnvironmentList();
        this.selectEnvironmentForEdit(this.selectedEnvId);
    }

    closeEnvironmentsModal() {
        const modal = document.getElementById('environments-modal');
        modal.classList.remove('active');
    }

    renderEnvironmentList() {
        const envList = document.getElementById('env-list');
        envList.innerHTML = '';

        Object.keys(this.environments.environments).forEach(envId => {
            const env = this.environments.environments[envId];
            const item = document.createElement('div');
            item.className = 'env-item' + (envId === this.editingEnvId ? ' active' : '');
            item.dataset.envId = envId;
            item.textContent = env.name;

            item.addEventListener('click', () => {
                this.saveCurrentEnvEditing();
                this.selectEnvironmentForEdit(envId);
            });

            envList.appendChild(item);
        });
    }

    selectEnvironmentForEdit(envId) {
        this.editingEnvId = envId;
        const env = this.environments.environments[envId];

        document.querySelectorAll('.env-item').forEach(item => {
            item.classList.remove('active');
            if (item.dataset.envId === envId) {
                item.classList.add('active');
            }
        });

        document.getElementById('env-name-input').value = env ? env.name : '';

        const varsBody = document.getElementById('env-vars-body');
        varsBody.innerHTML = '';

        if (env && env.variables) {
            Object.keys(env.variables).forEach(key => {
                this.addEnvVarRow(key, env.variables[key]);
            });
        }

        this.addEnvVarRow();
    }

    addEnvVarRow(key = '', value = '') {
        const tbody = document.getElementById('env-vars-body');
        const row = document.createElement('tr');
        
        row.innerHTML = `
            <td><input type="text" class="env-var-key" value="${this.escapeHtml(key)}" placeholder="变量名"></td>
            <td><input type="text" class="env-var-value" value="${this.escapeHtml(value)}" placeholder="值"></td>
            <td><button class="remove-btn">-</button></td>
        `;

        row.querySelector('.remove-btn').addEventListener('click', () => {
            row.remove();
        });

        tbody.appendChild(row);
    }

    saveCurrentEnvEditing() {
        if (!this.editingEnvId) return;

        const name = document.getElementById('env-name-input').value.trim();
        if (!name) return;

        const variables = {};
        const rows = document.querySelectorAll('#env-vars-body tr');
        
        rows.forEach(row => {
            const keyInput = row.querySelector('.env-var-key');
            const valueInput = row.querySelector('.env-var-value');
            
            if (keyInput && valueInput && keyInput.value.trim()) {
                variables[keyInput.value.trim()] = valueInput.value;
            }
        });

        this.environments.environments[this.editingEnvId] = {
            name: name,
            variables: variables
        };
    }

    addNewEnvironment() {
        const envId = 'env_' + Date.now();
        this.environments.environments[envId] = {
            name: '新环境',
            variables: {}
        };

        this.renderEnvironmentList();
        this.selectEnvironmentForEdit(envId);
    }

    deleteCurrentEnvironment() {
        if (!this.editingEnvId || this.editingEnvId === 'default') {
            alert('默认环境不能删除');
            return;
        }

        if (confirm('确定要删除这个环境吗？')) {
            delete this.environments.environments[this.editingEnvId];
            
            if (this.selectedEnvId === this.editingEnvId) {
                this.selectedEnvId = 'default';
                this.environments.active = 'default';
            }

            this.renderEnvironmentList();
            this.renderEnvironmentSelector();
            this.selectEnvironmentForEdit('default');
        }
    }

    async saveEnvironments() {
        await window.electronAPI.saveEnvironments(this.environments);
    }

    openSettingsModal() {
        const modal = document.getElementById('settings-modal');
        modal.classList.add('active');

        document.getElementById('timeout-input').value = this.settings.timeout / 1000;
    }

    closeSettingsModal() {
        const modal = document.getElementById('settings-modal');
        modal.classList.remove('active');
    }

    async saveSettings() {
        const timeoutSeconds = parseInt(document.getElementById('timeout-input').value) || 30;
        this.settings.timeout = timeoutSeconds * 1000;
        await window.electronAPI.saveSettings(this.settings);
    }
}

document.addEventListener('DOMContentLoaded', () => {
    new APIDebugger();
});
