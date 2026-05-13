let currentVersionId = null;
let currentAPIs = [];
let allVersions = [];

function showSection(sectionId) {
    document.querySelectorAll('.section').forEach(s => s.classList.add('hidden'));
    document.getElementById(sectionId).classList.remove('hidden');
    
    if (sectionId === 'versions') {
        loadVersions();
    } else if (sectionId === 'docs') {
        loadVersionSelects();
    } else if (sectionId === 'compare') {
        loadCompareSelects();
    }
}

async function loadVersions() {
    try {
        const res = await fetch('/api/versions');
        allVersions = await res.json();
        renderVersionsList();
    } catch (e) {
        console.error('加载版本失败:', e);
    }
}

function renderVersionsList() {
    const container = document.getElementById('versionsList');
    
    if (allVersions.length === 0) {
        container.innerHTML = '<div class="empty-state">暂无版本记录</div>';
        return;
    }
    
    container.innerHTML = allVersions.map(v => `
        <div class="list-item">
            <div class="list-item-info">
                <h4>${escapeHtml(v.name)} ${v.note ? `<small>(${escapeHtml(v.note)})</small>` : ''}</h4>
                <p>创建时间: ${formatDate(v.created_at)} | 源目录: ${escapeHtml(v.source_dir)}</p>
            </div>
            <div class="list-item-actions">
                <button class="btn" onclick="viewVersion(${v.id})">查看文档</button>
            </div>
        </div>
    `).join('');
}

async function loadVersionSelects() {
    try {
        const res = await fetch('/api/versions');
        allVersions = await res.json();
        
        const select = document.getElementById('versionSelect');
        select.innerHTML = '<option value="">选择版本...</option>' + 
            allVersions.map(v => `<option value="${v.id}">${escapeHtml(v.name)}</option>`).join('');
    } catch (e) {
        console.error('加载版本失败:', e);
    }
}

async function loadCompareSelects() {
    try {
        const res = await fetch('/api/versions');
        allVersions = await res.json();
        
        const oldSelect = document.getElementById('compareOldVersion');
        const newSelect = document.getElementById('compareNewVersion');
        
        const options = allVersions.map(v => `<option value="${v.id}">${escapeHtml(v.name)}</option>`).join('');
        
        oldSelect.innerHTML = options;
        newSelect.innerHTML = options;
        
        if (allVersions.length >= 2) {
            oldSelect.value = allVersions[1].id;
            newSelect.value = allVersions[0].id;
        }
    } catch (e) {
        console.error('加载版本失败:', e);
    }
}

document.getElementById('createVersionForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const name = document.getElementById('versionName').value;
    const moduleName = document.getElementById('moduleName').value;
    const sourceDir = document.getElementById('sourceDir').value;
    const note = document.getElementById('versionNote').value;
    
    try {
        const res = await fetch('/api/versions', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, module_name: moduleName, source_dir: sourceDir, note })
        });
        
        const data = await res.json();
        
        if (res.ok) {
            alert(`创建成功! 版本 ID: ${data.version_id}, 解析到 ${data.api_count} 个接口`);
            document.getElementById('createVersionForm').reset();
            loadVersions();
        } else {
            alert('创建失败: ' + (data.error || '未知错误'));
        }
    } catch (e) {
        alert('创建失败: ' + e.message);
    }
});

function viewVersion(id) {
    showSection('docs');
    document.getElementById('versionSelect').value = id;
    loadVersionDocs();
}

async function loadVersionDocs() {
    const select = document.getElementById('versionSelect');
    currentVersionId = select.value ? parseInt(select.value) : null;
    
    if (!currentVersionId) {
        document.getElementById('moduleList').innerHTML = '';
        document.getElementById('apiContent').innerHTML = '<div class="empty-state">请选择一个版本查看文档</div>';
        return;
    }
    
    try {
        const res = await fetch(`/api/versions/${currentVersionId}/apis`);
        currentAPIs = await res.json();
        
        const modulesRes = await fetch(`/api/versions/${currentVersionId}/modules`);
        const modules = await modulesRes.json();
        
        renderModules(modules);
        renderAPIs(currentAPIs);
    } catch (e) {
        console.error('加载文档失败:', e);
    }
}

function renderModules(modules) {
    const container = document.getElementById('moduleList');
    
    let html = '<h4>模块列表</h4>';
    html += '<div class="module-item active" onclick="filterByModule(null)">全部接口</div>';
    
    modules.forEach(m => {
        html += `<div class="module-item" onclick="filterByModule('${escapeHtml(m)}')">${escapeHtml(m)}</div>`;
    });
    
    container.innerHTML = html;
}

function filterByModule(module) {
    document.querySelectorAll('.module-item').forEach(item => {
        item.classList.remove('active');
    });
    event.target.classList.add('active');
    
    let filtered = currentAPIs;
    if (module) {
        filtered = currentAPIs.filter(api => api.module === module);
    }
    
    renderAPIs(filtered);
}

function renderAPIs(apis) {
    const container = document.getElementById('apiContent');
    
    if (apis.length === 0) {
        container.innerHTML = '<div class="empty-state">暂无接口数据</div>';
        return;
    }
    
    container.innerHTML = apis.map(api => renderAPICard(api)).join('');
}

function renderAPICard(api) {
    const incomplete = !api.is_complete;
    
    let paramsHtml = '';
    if (api.params && api.params.length > 0) {
        paramsHtml = `
            <div class="api-section">
                <h5>参数</h5>
                <table class="api-table">
                    <tr><th>名称</th><th>类型</th><th>位置</th><th>必填</th><th>描述</th></tr>
                    ${api.params.map(p => `
                        <tr>
                            <td>${escapeHtml(p.name)}</td>
                            <td>${escapeHtml(p.type)}</td>
                            <td>${escapeHtml(p.in)}</td>
                            <td>${p.required ? '是' : '否'}</td>
                            <td>${escapeHtml(p.description)}</td>
                        </tr>
                    `).join('')}
                </table>
            </div>
        `;
    }
    
    let returnsHtml = '';
    if (api.returns && api.returns.length > 0) {
        returnsHtml = `
            <div class="api-section">
                <h5>返回值</h5>
                <table class="api-table">
                    <tr><th>状态码</th><th>描述</th></tr>
                    ${api.returns.map(r => `
                        <tr>
                            <td>${escapeHtml(r.code)}</td>
                            <td>${escapeHtml(r.description)}</td>
                        </tr>
                    `).join('')}
                </table>
            </div>
        `;
    }
    
    let exampleHtml = '';
    if (api.example && (api.example.request || api.example.response)) {
        exampleHtml = '<div class="api-section"><h5>示例</h5>';
        if (api.example.request) {
            exampleHtml += `<p><strong>请求:</strong></p><pre class="code-block">${escapeHtml(api.example.request)}</pre>`;
        }
        if (api.example.response) {
            exampleHtml += `<p><strong>响应:</strong></p><pre class="code-block">${escapeHtml(api.example.response)}</pre>`;
        }
        exampleHtml += '</div>';
    } else {
        exampleHtml = '<div class="api-section"><h5>示例</h5><div class="no-example">暂无示例</div></div>';
    }
    
    return `
        <div class="api-item ${incomplete ? 'incomplete' : ''}">
            <div class="api-header">
                <span class="method-badge ${api.method}">${api.method}</span>
                <span class="api-path">${escapeHtml(api.path)}</span>
                ${incomplete ? '<span class="api-status incomplete">文档不完整</span>' : ''}
            </div>
            <div class="api-body">
                <div class="api-description">
                    <strong>模块:</strong> ${escapeHtml(api.module) || '默认'}<br>
                    <strong>Handler:</strong> ${escapeHtml(api.handler_name)}<br>
                    <strong>描述:</strong> ${escapeHtml(api.description) || '暂无描述'}
                    ${incomplete ? `<br><span style="color: #b45309;"><strong>缺失字段:</strong> ${api.missing_fields.join(', ')}</span>` : ''}
                </div>
                ${paramsHtml}
                ${returnsHtml}
                ${exampleHtml}
                <div class="api-actions">
                    <button class="btn" onclick="openTestModal('${escapeHtml(api.method)}', '${escapeHtml(api.path)}', '${escapeHtml(api.example ? api.example.request : '')}')">在线测试</button>
                </div>
            </div>
        </div>
    `;
}

async function searchAPIs() {
    const keyword = document.getElementById('searchInput').value;
    
    if (!currentVersionId) return;
    
    try {
        const res = await fetch(`/api/versions/${currentVersionId}/search?q=${encodeURIComponent(keyword)}`);
        const apis = await res.json();
        
        document.querySelectorAll('.module-item').forEach(item => {
            item.classList.remove('active');
            if (item.textContent === '全部接口') {
                item.classList.add('active');
            }
        });
        
        renderAPIs(apis);
    } catch (e) {
        console.error('搜索失败:', e);
    }
}

function exportMarkdown() {
    if (!currentVersionId) {
        alert('请先选择版本');
        return;
    }
    window.open(`/api/versions/${currentVersionId}/export/markdown`, '_blank');
}

function exportOpenAPI() {
    if (!currentVersionId) {
        alert('请先选择版本');
        return;
    }
    window.open(`/api/versions/${currentVersionId}/export/openapi`, '_blank');
}

async function compareVersions() {
    const oldId = document.getElementById('compareOldVersion').value;
    const newId = document.getElementById('compareNewVersion').value;
    
    if (!oldId || !newId) {
        alert('请选择两个版本进行对比');
        return;
    }
    
    try {
        const res = await fetch(`/api/compare?old=${oldId}&new=${newId}`);
        const diff = await res.json();
        
        renderDiffResult(diff);
    } catch (e) {
        console.error('对比失败:', e);
    }
}

function renderDiffResult(diff) {
    const container = document.getElementById('compareResult');
    
    let html = '';
    
    if (diff.added && diff.added.length > 0) {
        html += `
            <div class="diff-section added">
                <h3>✅ 新增 (${diff.added.length})</h3>
                <div class="diff-list">
                    ${diff.added.map(item => `
                        <div class="diff-item added">
                            <strong>[${escapeHtml(item.module)}]</strong> ${escapeHtml(item.method)} ${escapeHtml(item.path)}
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    }
    
    if (diff.removed && diff.removed.length > 0) {
        html += `
            <div class="diff-section removed">
                <h3>❌ 删除 (${diff.removed.length})</h3>
                <div class="diff-list">
                    ${diff.removed.map(item => `
                        <div class="diff-item removed">
                            <strong>[${escapeHtml(item.module)}]</strong> ${escapeHtml(item.method)} ${escapeHtml(item.path)}
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    }
    
    if (diff.modified && diff.modified.length > 0) {
        html += `
            <div class="diff-section modified">
                <h3>🔄 修改 (${diff.modified.length})</h3>
                <div class="diff-list">
                    ${diff.modified.map(item => `
                        <div class="diff-item modified">
                            <strong>[${escapeHtml(item.module)}]</strong> ${escapeHtml(item.method)} ${escapeHtml(item.path)}
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    }
    
    if (!html) {
        html = '<div class="empty-state">两个版本完全一致，没有差异</div>';
    }
    
    container.innerHTML = html;
}

function openTestModal(method, path, exampleReq) {
    document.getElementById('testModal').classList.remove('hidden');
    document.getElementById('testMethod').value = method === 'ALL' ? 'GET' : method;
    document.getElementById('testUrl').value = 'http://localhost:9200' + path;
    document.getElementById('testHeaders').value = '{"Content-Type": "application/json"}';
    document.getElementById('testBody').value = exampleReq || '';
    document.getElementById('testResult').classList.add('hidden');
}

function closeTestModal() {
    document.getElementById('testModal').classList.add('hidden');
}

async function executeTest() {
    const method = document.getElementById('testMethod').value;
    const url = document.getElementById('testUrl').value;
    const headersStr = document.getElementById('testHeaders').value;
    const body = document.getElementById('testBody').value;
    
    let headers = {};
    if (headersStr.trim()) {
        try {
            headers = JSON.parse(headersStr);
        } catch (e) {
            alert('请求头格式错误，请输入有效的 JSON');
            return;
        }
    }
    
    try {
        const res = await fetch('/api/test', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ method, url, headers, body })
        });
        
        const data = await res.json();
        
        if (res.ok) {
            renderTestResult(data);
        } else {
            document.getElementById('testResult').innerHTML = `
                <div class="result-status error">请求失败</div>
                <pre class="code-block">${escapeHtml(data.error || JSON.stringify(data, null, 2))}</pre>
            `;
            document.getElementById('testResult').classList.remove('hidden');
        }
    } catch (e) {
        document.getElementById('testResult').innerHTML = `
            <div class="result-status error">错误</div>
            <pre class="code-block">${escapeHtml(e.message)}</pre>
        `;
        document.getElementById('testResult').classList.remove('hidden');
    }
}

function renderTestResult(data) {
    const isSuccess = data.status_code >= 200 && data.status_code < 400;
    
    let headersHtml = '';
    if (data.headers) {
        headersHtml = Object.entries(data.headers).map(([k, v]) => 
            `${escapeHtml(k)}: ${escapeHtml(v)}`
        ).join('\n');
    }
    
    document.getElementById('testResult').innerHTML = `
        <h4>测试结果</h4>
        <div class="result-status ${isSuccess ? 'success' : 'error'}">
            HTTP ${data.status_code}
        </div>
        <span class="result-duration">耗时: ${data.duration}</span>
        
        ${headersHtml ? `
            <div class="api-section">
                <h5>响应头</h5>
                <pre class="code-block">${escapeHtml(headersHtml)}</pre>
            </div>
        ` : ''}
        
        <div class="api-section">
            <h5>响应体</h5>
            <pre class="code-block">${escapeHtml(data.body)}</pre>
        </div>
    `;
    document.getElementById('testResult').classList.remove('hidden');
}

function escapeHtml(str) {
    if (!str) return '';
    return String(str)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');
}

function formatDate(dateStr) {
    if (!dateStr) return '';
    return new Date(dateStr).toLocaleString('zh-CN');
}

document.addEventListener('DOMContentLoaded', () => {
    loadVersions();
});
