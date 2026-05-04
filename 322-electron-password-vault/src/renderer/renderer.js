const { electronAPI } = window;

let appData = {
  categories: [],
  entries: []
};

let currentCategoryId = 'all';
let currentEntryId = null;
let isEditing = false;
let originalEntry = null;
let selectedIcon = '📁';
let lockTimer = null;

// DOM Elements
const authScreen = document.getElementById('auth-screen');
const mainScreen = document.getElementById('main-screen');
const setupForm = document.getElementById('setup-form');
const loginForm = document.getElementById('login-form');
const authTitle = document.getElementById('auth-title');
const authSubtitle = document.getElementById('auth-subtitle');
const setupPassword = document.getElementById('setup-password');
const setupConfirm = document.getElementById('setup-confirm');
const setupBtn = document.getElementById('setup-btn');
const setupError = document.getElementById('setup-error');
const loginPassword = document.getElementById('login-password');
const loginBtn = document.getElementById('login-btn');
const loginError = document.getElementById('login-error');
const lockMessage = document.getElementById('lock-message');

const categoryList = document.getElementById('category-list');
const addCategoryBtn = document.getElementById('add-category-btn');
const entriesList = document.getElementById('entries-list');
const currentCategoryName = document.getElementById('current-category-name');
const addEntryBtn = document.getElementById('add-entry-btn');

const emptyDetail = document.getElementById('empty-detail');
const entryForm = document.getElementById('entry-form');
const detailTitle = document.getElementById('detail-title');
const saveEntryBtn = document.getElementById('save-entry-btn');
const deleteEntryBtn = document.getElementById('delete-entry-btn');
const cancelEditBtn = document.getElementById('cancel-edit-btn');

const entryWebsite = document.getElementById('entry-website');
const entryUsername = document.getElementById('entry-username');
const entryPassword = document.getElementById('entry-password');
const entryUrl = document.getElementById('entry-url');
const entryNotes = document.getElementById('entry-notes');
const entryCategory = document.getElementById('entry-category');

const searchInput = document.getElementById('search-input');
const clearSearch = document.getElementById('clear-search');
const lockBtn = document.getElementById('lock-btn');

const categoryModal = document.getElementById('category-modal');
const newCategoryName = document.getElementById('new-category-name');
const saveCategoryBtn = document.getElementById('save-category-btn');
const iconPicker = document.getElementById('icon-picker');

const passwordModal = document.getElementById('password-modal');
const generatedPassword = document.getElementById('generated-password');
const passwordLength = document.getElementById('password-length');
const lengthValue = document.getElementById('length-value');
const useUppercase = document.getElementById('use-uppercase');
const useLowercase = document.getElementById('use-lowercase');
const useNumbers = document.getElementById('use-numbers');
const useSpecial = document.getElementById('use-special');
const regenerateBtn = document.getElementById('regenerate-btn');
const copyGeneratedBtn = document.getElementById('copy-generated-btn');
const usePasswordBtn = document.getElementById('use-password-btn');

const toast = document.getElementById('toast');

// Utility Functions
function showToast(message, duration = 2500) {
  toast.textContent = message;
  toast.style.display = 'block';
  
  setTimeout(() => {
    toast.style.display = 'none';
  }, duration);
}

function generateId() {
  return Date.now().toString(36) + Math.random().toString(36).substr(2);
}

// Auth Functions
async function initAuth() {
  const hasMasterPassword = await electronAPI.checkHasMasterPassword();
  
  if (hasMasterPassword) {
    authTitle.textContent = '输入主密码';
    authSubtitle.textContent = '请输入主密码解锁您的密码保险箱';
    setupForm.style.display = 'none';
    loginForm.style.display = 'flex';
    
    await checkLockStatus();
  } else {
    authTitle.textContent = '设置主密码';
    authSubtitle.textContent = '请设置一个安全的主密码来保护您的数据';
    setupForm.style.display = 'flex';
    loginForm.style.display = 'none';
  }
}

async function checkLockStatus() {
  const status = await electronAPI.getLockStatus();
  
  if (status.isLocked) {
    loginBtn.disabled = true;
    loginPassword.disabled = true;
    lockMessage.style.display = 'block';
    lockMessage.textContent = `已锁定，请等待 ${status.remainingSeconds} 秒后再试`;
    
    startLockTimer(status.remainingSeconds);
  } else {
    loginBtn.disabled = false;
    loginPassword.disabled = false;
    lockMessage.style.display = 'none';
  }
}

function startLockTimer(seconds) {
  let remaining = seconds;
  
  if (lockTimer) {
    clearInterval(lockTimer);
  }
  
  lockTimer = setInterval(() => {
    remaining--;
    if (remaining <= 0) {
      clearInterval(lockTimer);
      lockTimer = null;
      checkLockStatus();
    } else {
      lockMessage.textContent = `已锁定，请等待 ${remaining} 秒后再试`;
    }
  }, 1000);
}

async function setupMasterPassword() {
  const password = setupPassword.value;
  const confirm = setupConfirm.value;
  
  setupError.textContent = '';
  
  if (!password || password.length < 4) {
    setupError.textContent = '主密码至少需要 4 个字符';
    return;
  }
  
  if (password !== confirm) {
    setupError.textContent = '两次输入的密码不一致';
    return;
  }
  
  const result = await electronAPI.setupMasterPassword(password);
  
  if (result.success) {
    showToast('主密码设置成功！');
    await loadData();
    showMainScreen();
  } else {
    setupError.textContent = result.error || '设置失败，请重试';
  }
}

async function verifyMasterPassword() {
  const password = loginPassword.value;
  
  loginError.textContent = '';
  
  if (!password) {
    loginError.textContent = '请输入主密码';
    return;
  }
  
  const result = await electronAPI.verifyMasterPassword(password);
  
  if (result.success) {
    showToast('解锁成功！');
    await loadData();
    showMainScreen();
  } else {
    if (result.isLocked) {
      loginBtn.disabled = true;
      loginPassword.disabled = true;
      lockMessage.style.display = 'block';
      lockMessage.textContent = `已锁定，请等待 ${result.remainingSeconds} 秒后再试`;
      startLockTimer(result.remainingSeconds);
    } else {
      loginError.textContent = `密码错误 (${result.failedAttempts}/${result.maxAttempts})`;
      loginPassword.value = '';
      loginPassword.focus();
    }
  }
}

function showMainScreen() {
  authScreen.style.display = 'none';
  mainScreen.style.display = 'flex';
  renderCategories();
  renderEntries();
}

function showAuthScreen() {
  mainScreen.style.display = 'none';
  authScreen.style.display = 'flex';
  setupPassword.value = '';
  setupConfirm.value = '';
  loginPassword.value = '';
  setupError.textContent = '';
  loginError.textContent = '';
  initAuth();
}

// Data Functions
async function loadData() {
  const result = await electronAPI.loadData();
  
  if (result.success) {
    appData = result.data;
  } else {
    showToast('加载数据失败: ' + result.error);
  }
}

async function saveData() {
  const result = await electronAPI.saveData(appData);
  
  if (!result.success) {
    showToast('保存数据失败: ' + result.error);
  }
}

// Category Functions
function renderCategories() {
  categoryList.innerHTML = '';
  
  const allLi = document.createElement('li');
  allLi.className = currentCategoryId === 'all' ? 'active' : '';
  allLi.innerHTML = `
    <span class="category-icon">📋</span>
    <span class="category-name">全部</span>
    <span class="category-count">${appData.entries.length}</span>
  `;
  allLi.addEventListener('click', () => selectCategory('all'));
  categoryList.appendChild(allLi);
  
  appData.categories.forEach(category => {
    const count = appData.entries.filter(e => e.categoryId === category.id).length;
    const li = document.createElement('li');
    li.className = currentCategoryId === category.id ? 'active' : '';
    li.innerHTML = `
      <span class="category-icon">${category.icon}</span>
      <span class="category-name">${category.name}</span>
      <span class="category-count">${count}</span>
    `;
    li.addEventListener('click', () => selectCategory(category.id));
    categoryList.appendChild(li);
  });
  
  updateCategorySelect();
}

function updateCategorySelect() {
  entryCategory.innerHTML = '';
  
  appData.categories.forEach(category => {
    const option = document.createElement('option');
    option.value = category.id;
    option.textContent = `${category.icon} ${category.name}`;
    entryCategory.appendChild(option);
  });
}

function selectCategory(categoryId) {
  currentCategoryId = categoryId;
  currentEntryId = null;
  
  if (categoryId === 'all') {
    currentCategoryName.textContent = '全部条目';
  } else {
    const category = appData.categories.find(c => c.id === categoryId);
    currentCategoryName.textContent = category ? `${category.icon} ${category.name}` : '全部条目';
  }
  
  renderCategories();
  renderEntries();
  showEmptyDetail();
}

// Entry Functions
function getFilteredEntries() {
  let entries = [...appData.entries];
  
  if (currentCategoryId !== 'all') {
    entries = entries.filter(e => e.categoryId === currentCategoryId);
  }
  
  const searchTerm = searchInput.value.trim().toLowerCase();
  if (searchTerm) {
    entries = entries.filter(e => 
      e.website.toLowerCase().includes(searchTerm) ||
      e.username.toLowerCase().includes(searchTerm)
    );
  }
  
  return entries;
}

function renderEntries() {
  const entries = getFilteredEntries();
  
  entriesList.innerHTML = '';
  
  if (entries.length === 0) {
    const li = document.createElement('li');
    li.className = 'empty-state';
    li.textContent = searchInput.value ? '未找到匹配的条目' : '暂无条目';
    entriesList.appendChild(li);
    return;
  }
  
  entries.forEach(entry => {
    const li = document.createElement('li');
    li.className = currentEntryId === entry.id ? 'active' : '';
    li.innerHTML = `
      <div class="entry-item-name">${escapeHtml(entry.website)}</div>
      <div class="entry-item-username">${escapeHtml(entry.username) || '未设置用户名'}</div>
    `;
    li.addEventListener('click', () => selectEntry(entry.id));
    entriesList.appendChild(li);
  });
}

function escapeHtml(text) {
  if (!text) return '';
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

function selectEntry(entryId) {
  const entry = appData.entries.find(e => e.id === entryId);
  if (!entry) return;
  
  currentEntryId = entryId;
  isEditing = false;
  originalEntry = { ...entry };
  
  renderEntries();
  showEntryDetail(entry);
}

function showEntryDetail(entry) {
  emptyDetail.style.display = 'none';
  entryForm.style.display = 'block';
  
  detailTitle.textContent = '编辑条目';
  saveEntryBtn.style.display = 'inline-block';
  deleteEntryBtn.style.display = 'inline-block';
  cancelEditBtn.style.display = 'inline-block';
  
  entryWebsite.value = entry.website || '';
  entryUsername.value = entry.username || '';
  entryPassword.value = entry.password || '';
  entryUrl.value = entry.url || '';
  entryNotes.value = entry.notes || '';
  entryCategory.value = entry.categoryId || appData.categories[0]?.id || '';
  
  entryPassword.type = 'password';
}

function showEmptyDetail() {
  entryForm.style.display = 'none';
  emptyDetail.style.display = 'flex';
  
  detailTitle.textContent = '条目详情';
  saveEntryBtn.style.display = 'none';
  deleteEntryBtn.style.display = 'none';
  cancelEditBtn.style.display = 'none';
}

function createNewEntry() {
  const newEntry = {
    id: generateId(),
    website: '',
    username: '',
    password: '',
    url: '',
    notes: '',
    categoryId: currentCategoryId === 'all' ? (appData.categories[0]?.id || '') : currentCategoryId,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString()
  };
  
  currentEntryId = newEntry.id;
  isEditing = true;
  originalEntry = null;
  
  appData.entries.push(newEntry);
  
  renderEntries();
  showEntryDetail(newEntry);
  entryWebsite.focus();
}

async function saveCurrentEntry() {
  if (!currentEntryId) return;
  
  const website = entryWebsite.value.trim();
  
  if (!website) {
    showToast('网站名称不能为空');
    entryWebsite.focus();
    return;
  }
  
  const entry = appData.entries.find(e => e.id === currentEntryId);
  if (!entry) return;
  
  entry.website = website;
  entry.username = entryUsername.value.trim();
  entry.password = entryPassword.value;
  entry.url = entryUrl.value.trim();
  entry.notes = entryNotes.value.trim();
  entry.categoryId = entryCategory.value;
  entry.updatedAt = new Date().toISOString();
  
  await saveData();
  
  originalEntry = { ...entry };
  isEditing = false;
  
  renderCategories();
  renderEntries();
  
  showToast('保存成功');
}

async function deleteCurrentEntry() {
  if (!currentEntryId) return;
  
  if (!confirm('确定要删除这个条目吗？')) return;
  
  const index = appData.entries.findIndex(e => e.id === currentEntryId);
  if (index > -1) {
    appData.entries.splice(index, 1);
    await saveData();
  }
  
  currentEntryId = null;
  originalEntry = null;
  isEditing = false;
  
  renderCategories();
  renderEntries();
  showEmptyDetail();
  
  showToast('已删除');
}

function cancelEdit() {
  if (!currentEntryId) return;
  
  if (originalEntry) {
    showEntryDetail(originalEntry);
    isEditing = false;
  } else {
    const index = appData.entries.findIndex(e => e.id === currentEntryId);
    if (index > -1) {
      appData.entries.splice(index, 1);
    }
    currentEntryId = null;
    renderEntries();
    showEmptyDetail();
  }
}

// Category Modal
function openCategoryModal() {
  newCategoryName.value = '';
  selectedIcon = '📁';
  
  iconPicker.querySelectorAll('.icon-option').forEach(btn => {
    btn.classList.remove('selected');
    if (btn.dataset.icon === selectedIcon) {
      btn.classList.add('selected');
    }
  });
  
  categoryModal.style.display = 'flex';
  newCategoryName.focus();
}

function closeCategoryModal() {
  categoryModal.style.display = 'none';
}

iconPicker.addEventListener('click', (e) => {
  const btn = e.target.closest('.icon-option');
  if (!btn) return;
  
  iconPicker.querySelectorAll('.icon-option').forEach(b => b.classList.remove('selected'));
  btn.classList.add('selected');
  selectedIcon = btn.dataset.icon;
});

async function addNewCategory() {
  const name = newCategoryName.value.trim();
  
  if (!name) {
    showToast('请输入分类名称');
    newCategoryName.focus();
    return;
  }
  
  const newCategory = {
    id: generateId(),
    name: name,
    icon: selectedIcon
  };
  
  appData.categories.push(newCategory);
  await saveData();
  
  renderCategories();
  closeCategoryModal();
  
  showToast('分类已添加');
}

// Password Generator Modal
function openPasswordModal() {
  updateGeneratedPassword();
  passwordModal.style.display = 'flex';
}

function closePasswordModal() {
  passwordModal.style.display = 'none';
}

async function updateGeneratedPassword() {
  const options = {
    length: parseInt(passwordLength.value),
    useUppercase: useUppercase.checked,
    useLowercase: useLowercase.checked,
    useNumbers: useNumbers.checked,
    useSpecial: useSpecial.checked
  };
  
  const result = await electronAPI.generatePassword(options);
  
  if (result.success) {
    generatedPassword.value = result.password;
  } else {
    showToast('生成密码失败: ' + result.error);
  }
}

function useGeneratedPassword() {
  if (generatedPassword.value) {
    entryPassword.value = generatedPassword.value;
    closePasswordModal();
  }
}

async function copyGeneratedPassword() {
  if (!generatedPassword.value) return;
  
  const result = await electronAPI.copyToClipboard(generatedPassword.value);
  
  if (result.success) {
    showToast('密码已复制 (3秒后自动清空)');
  } else {
    showToast('复制失败');
  }
}

// Copy Functions
async function copyField(field) {
  let text = '';
  
  if (field === 'username') {
    text = entryUsername.value;
  } else if (field === 'password') {
    text = entryPassword.value;
  }
  
  if (!text) {
    showToast('没有可复制的内容');
    return;
  }
  
  const result = await electronAPI.copyToClipboard(text);
  
  if (result.success) {
    const fieldName = field === 'username' ? '用户名' : '密码';
    showToast(`${fieldName}已复制 (3秒后自动清空)`);
  } else {
    showToast('复制失败');
  }
}

// Toggle Password Visibility
function togglePasswordVisibility(targetId) {
  const input = document.getElementById(targetId);
  if (!input) return;
  
  if (input.type === 'password') {
    input.type = 'text';
  } else {
    input.type = 'password';
  }
}

// Search Functions
function handleSearch() {
  const hasValue = searchInput.value.trim().length > 0;
  clearSearch.style.display = hasValue ? 'block' : 'none';
  
  renderEntries();
  
  if (currentEntryId) {
    const filtered = getFilteredEntries();
    const stillExists = filtered.some(e => e.id === currentEntryId);
    if (!stillExists) {
      currentEntryId = null;
      showEmptyDetail();
    }
  }
}

function clearSearchInput() {
  searchInput.value = '';
  clearSearch.style.display = 'none';
  handleSearch();
}

// Lock Function
async function lockVault() {
  const result = await electronAPI.lockVault();
  
  if (result.success) {
    appData = { categories: [], entries: [] };
    currentCategoryId = 'all';
    currentEntryId = null;
    showAuthScreen();
  }
}

// Event Listeners
document.addEventListener('DOMContentLoaded', () => {
  initAuth();
});

// Auth Events
setupBtn.addEventListener('click', setupMasterPassword);
setupConfirm.addEventListener('keypress', (e) => {
  if (e.key === 'Enter') setupMasterPassword();
});

loginBtn.addEventListener('click', verifyMasterPassword);
loginPassword.addEventListener('keypress', (e) => {
  if (e.key === 'Enter') verifyMasterPassword();
});

// Toggle Password Visibility
document.querySelectorAll('.toggle-password').forEach(btn => {
  btn.addEventListener('click', () => {
    const targetId = btn.dataset.target;
    togglePasswordVisibility(targetId);
  });
});

// Main Events
addCategoryBtn.addEventListener('click', openCategoryModal);
addEntryBtn.addEventListener('click', createNewEntry);

saveEntryBtn.addEventListener('click', saveCurrentEntry);
deleteEntryBtn.addEventListener('click', deleteCurrentEntry);
cancelEditBtn.addEventListener('click', cancelEdit);

// Search
searchInput.addEventListener('input', handleSearch);
clearSearch.addEventListener('click', clearSearchInput);

// Lock
lockBtn.addEventListener('click', lockVault);

// Category Modal
document.querySelectorAll('#category-modal .close-modal').forEach(btn => {
  btn.addEventListener('click', closeCategoryModal);
});
saveCategoryBtn.addEventListener('click', addNewCategory);
newCategoryName.addEventListener('keypress', (e) => {
  if (e.key === 'Enter') addNewCategory();
});

// Password Modal
document.querySelectorAll('#password-modal .close-modal').forEach(btn => {
  btn.addEventListener('click', closePasswordModal);
});
document.getElementById('generate-password-btn').addEventListener('click', openPasswordModal);
regenerateBtn.addEventListener('click', updateGeneratedPassword);
copyGeneratedBtn.addEventListener('click', copyGeneratedPassword);
usePasswordBtn.addEventListener('click', useGeneratedPassword);

passwordLength.addEventListener('input', () => {
  lengthValue.textContent = passwordLength.value;
  updateGeneratedPassword();
});

[useUppercase, useLowercase, useNumbers, useSpecial].forEach(checkbox => {
  checkbox.addEventListener('change', updateGeneratedPassword);
});

// Copy Buttons
document.querySelectorAll('.copy-btn[data-field]').forEach(btn => {
  btn.addEventListener('click', () => {
    const field = btn.dataset.field;
    copyField(field);
  });
});

// Click outside modal to close
categoryModal.addEventListener('click', (e) => {
  if (e.target === categoryModal) closeCategoryModal();
});

passwordModal.addEventListener('click', (e) => {
  if (e.target === passwordModal) closePasswordModal();
});
