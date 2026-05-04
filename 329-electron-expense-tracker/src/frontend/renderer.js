class ExpenseTrackerApp {
  constructor() {
    this.currentType = 'expense';
    this.currentYear = new Date().getFullYear();
    this.currentMonth = new Date().getMonth() + 1;
    this.selectedDate = new Date().toISOString().split('T')[0];
    this.categories = [];
    this.editingRecordId = null;
    this.deletingRecordId = null;
    this.isSearchMode = false;
    
    this.init();
  }

  async init() {
    await this.loadCategories();
    this.initDateInputs();
    this.initEventListeners();
    await this.renderCalendar();
    await this.loadSelectedDateRecords();
    await this.loadBudgetStatus();
  }

  async loadCategories() {
    this.categories = await window.electronAPI.getCategories();
    this.updateCategorySelects();
  }

  updateCategorySelects() {
    const categorySelect = document.getElementById('category');
    const searchCategorySelect = document.getElementById('search-category');
    const editCategorySelect = document.getElementById('edit-category');

    this.updateCategorySelect(categorySelect, this.currentType);
    this.updateCategorySelect(editCategorySelect, this.currentType, true);
    
    searchCategorySelect.innerHTML = '<option value="">全部</option>';
    this.categories.forEach(cat => {
      const option = document.createElement('option');
      option.value = cat.name;
      option.textContent = cat.name;
      searchCategorySelect.appendChild(option);
    });
  }

  updateCategorySelect(select, type, includeAll = false) {
    select.innerHTML = includeAll ? '<option value="">请选择</option>' : '';
    
    const filteredCategories = this.categories.filter(cat => cat.type === type);
    filteredCategories.forEach(cat => {
      const option = document.createElement('option');
      option.value = cat.name;
      option.textContent = cat.name;
      select.appendChild(option);
    });
  }

  initDateInputs() {
    const today = new Date().toISOString().split('T')[0];
    document.getElementById('date').value = today;
  }

  initEventListeners() {
    document.getElementById('btn-expense').addEventListener('click', () => this.switchType('expense'));
    document.getElementById('btn-income').addEventListener('click', () => this.switchType('income'));
    document.getElementById('btn-add').addEventListener('click', () => this.addRecord());

    document.getElementById('prev-month').addEventListener('click', () => this.changeMonth(-1));
    document.getElementById('next-month').addEventListener('click', () => this.changeMonth(1));

    document.getElementById('btn-search').addEventListener('click', () => this.toggleSearchPanel());
    document.getElementById('btn-stats').addEventListener('click', () => this.showStatsPanel());
    document.getElementById('btn-export').addEventListener('click', () => this.exportToCSV());
    document.getElementById('btn-close-stats').addEventListener('click', () => this.hideStatsPanel());

    document.getElementById('btn-do-search').addEventListener('click', () => this.doSearch());
    document.getElementById('btn-clear-search').addEventListener('click', () => this.clearSearch());

    document.getElementById('btn-set-budget').addEventListener('click', () => this.showBudgetModal());
    document.getElementById('btn-close-budget').addEventListener('click', () => this.hideBudgetModal());
    document.getElementById('btn-cancel-budget').addEventListener('click', () => this.hideBudgetModal());
    document.getElementById('btn-save-budget').addEventListener('click', () => this.saveBudgets());

    document.getElementById('btn-close-edit').addEventListener('click', () => this.hideEditModal());
    document.getElementById('btn-cancel-edit').addEventListener('click', () => this.hideEditModal());
    document.getElementById('btn-save-edit').addEventListener('click', () => this.saveEdit());
    document.getElementById('edit-btn-expense').addEventListener('click', () => this.switchEditType('expense'));
    document.getElementById('edit-btn-income').addEventListener('click', () => this.switchEditType('income'));

    document.getElementById('btn-confirm-delete').addEventListener('click', () => this.confirmDelete());
    document.getElementById('btn-cancel-delete').addEventListener('click', () => this.hideDeleteModal());

    this.initAmountInputValidation('amount');
    this.initAmountInputValidation('edit-amount');
    this.initAmountInputValidation('search-min-amount');
    this.initAmountInputValidation('search-max-amount');
  }

  initAmountInputValidation(id) {
    const input = document.getElementById(id);
    input.addEventListener('input', (e) => {
      let value = e.target.value;
      value = value.replace(/[^\d.]/g, '');
      
      const dotIndex = value.indexOf('.');
      if (dotIndex !== -1) {
        const integerPart = value.substring(0, dotIndex);
        let decimalPart = value.substring(dotIndex + 1);
        decimalPart = decimalPart.replace(/[^0-9]/g, '').substring(0, 2);
        value = integerPart + '.' + decimalPart;
      }
      
      if (value.startsWith('0') && value.length > 1 && value[1] !== '.') {
        value = value.substring(1);
      }
      
      e.target.value = value;
    });

    input.addEventListener('keypress', (e) => {
      const char = e.key;
      if (char === '-' || char === '+') {
        e.preventDefault();
      }
    });
  }

  switchType(type) {
    this.currentType = type;
    document.getElementById('btn-expense').classList.toggle('active', type === 'expense');
    document.getElementById('btn-income').classList.toggle('active', type === 'income');
    this.updateCategorySelects();
  }

  switchEditType(type) {
    const buttons = document.querySelectorAll('#edit-modal .type-switch button');
    buttons[0].classList.toggle('active', type === 'expense');
    buttons[1].classList.toggle('active', type === 'income');
    this.updateCategorySelect(document.getElementById('edit-category'), type);
  }

  async addRecord() {
    const amount = document.getElementById('amount').value;
    const category = document.getElementById('category').value;
    const note = document.getElementById('note').value;
    const date = document.getElementById('date').value;

    if (!amount || parseFloat(amount) <= 0) {
      alert('请输入有效的金额');
      return;
    }

    if (!category) {
      alert('请选择分类');
      return;
    }

    const record = {
      amount: parseFloat(amount),
      type: this.currentType,
      category: category,
      note: note,
      date: date
    };

    await window.electronAPI.addRecord(record);

    document.getElementById('amount').value = '';
    document.getElementById('note').value = '';

    await this.renderCalendar();
    
    if (date === this.selectedDate) {
      await this.loadSelectedDateRecords();
    }
    
    await this.loadBudgetStatus();
  }

  changeMonth(delta) {
    this.currentMonth += delta;
    
    if (this.currentMonth > 12) {
      this.currentMonth = 1;
      this.currentYear++;
    } else if (this.currentMonth < 1) {
      this.currentMonth = 12;
      this.currentYear--;
    }

    this.renderCalendar();
    this.loadBudgetStatus();
  }

  async renderCalendar() {
    const calendarEl = document.getElementById('calendar');
    const monthLabel = document.getElementById('current-month');
    
    monthLabel.textContent = `${this.currentYear}年${this.currentMonth}月`;

    const firstDay = new Date(this.currentYear, this.currentMonth - 1, 1);
    const lastDay = new Date(this.currentYear, this.currentMonth, 0);
    const startDay = firstDay.getDay();
    const totalDays = lastDay.getDate();

    const dailyExpenses = await window.electronAPI.getDailyExpenses(this.currentYear, this.currentMonth);

    let allExpenses = [];
    for (const date in dailyExpenses) {
      allExpenses.push(dailyExpenses[date].expense);
    }
    const maxExpense = allExpenses.length > 0 ? Math.max(...allExpenses) : 0;

    const today = new Date();
    const todayStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;

    let html = `
      <div class="calendar-weekday">
        <div>日</div>
        <div>一</div>
        <div>二</div>
        <div>三</div>
        <div>四</div>
        <div>五</div>
        <div>六</div>
      </div>
      <div class="calendar-days">
    `;

    const prevMonthLastDay = new Date(this.currentYear, this.currentMonth - 1, 0).getDate();
    for (let i = startDay - 1; i >= 0; i--) {
      const day = prevMonthLastDay - i;
      html += `<div class="calendar-day other-month"><span class="day-number">${day}</span></div>`;
    }

    for (let day = 1; day <= totalDays; day++) {
      const dateStr = `${this.currentYear}-${String(this.currentMonth).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
      const dayData = dailyExpenses[dateStr];
      
      let classes = ['calendar-day'];
      let amountHtml = '';

      if (dateStr === todayStr) {
        classes.push('today');
      }
      if (dateStr === this.selectedDate) {
        classes.push('selected');
      }

      if (dayData && dayData.expense > 0) {
        const expensePercent = maxExpense > 0 ? (dayData.expense / maxExpense) : 0;
        let level = Math.ceil(expensePercent * 5);
        if (level === 0) level = 1;
        classes.push(`expense-level-${level}`);
        amountHtml = `<span class="day-amount">¥${dayData.expense.toFixed(0)}</span>`;
      }

      html += `
        <div class="${classes.join(' ')}" data-date="${dateStr}">
          <span class="day-number">${day}</span>
          ${amountHtml}
        </div>
      `;
    }

    const remainingCells = 42 - (startDay + totalDays);
    for (let day = 1; day <= remainingCells; day++) {
      html += `<div class="calendar-day other-month"><span class="day-number">${day}</span></div>`;
    }

    html += '</div>';
    calendarEl.innerHTML = html;

    calendarEl.querySelectorAll('.calendar-day:not(.other-month)').forEach(dayEl => {
      dayEl.addEventListener('click', () => {
        const date = dayEl.dataset.date;
        this.selectDate(date);
      });
    });
  }

  async selectDate(date) {
    this.selectedDate = date;
    this.isSearchMode = false;
    document.getElementById('search-panel').classList.add('hidden');
    document.getElementById('selected-date-title').textContent = `${date} 明细`;
    
    await this.renderCalendar();
    await this.loadSelectedDateRecords();
  }

  async loadSelectedDateRecords() {
    const records = await window.electronAPI.getRecordsByDate(this.selectedDate);
    this.renderRecords(records);
  }

  renderRecords(records) {
    const listEl = document.getElementById('records-list');
    
    if (records.length === 0) {
      listEl.innerHTML = '<div class="empty-state">暂无记录</div>';
      return;
    }

    records.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));

    let html = '';
    records.forEach(record => {
      const amountPrefix = record.type === 'income' ? '+' : '-';
      const amountClass = record.type === 'income' ? 'income' : 'expense';
      const time = record.date.split('T')[0];
      
      html += `
        <div class="record-item" data-id="${record.id}">
          <div class="record-info">
            <div class="record-header">
              <span class="record-category">${record.category}</span>
              <span class="record-amount ${amountClass}">${amountPrefix}¥${record.amount.toFixed(2)}</span>
            </div>
            <div class="record-footer">
              <span class="record-note">${record.note || '无备注'}</span>
              <span class="record-time">${time}</span>
            </div>
          </div>
          <div class="record-actions">
            <button class="btn-secondary" onclick="app.editRecord('${record.id}')">编辑</button>
            <button class="btn-danger" onclick="app.deleteRecord('${record.id}')">删除</button>
          </div>
        </div>
      `;
    });

    listEl.innerHTML = html;
  }

  async loadBudgetStatus() {
    const budgetStatus = await window.electronAPI.getBudgetStatus(this.currentYear, this.currentMonth);
    const listEl = document.getElementById('budget-list');

    if (budgetStatus.length === 0) {
      listEl.innerHTML = '<div style="padding: 10px 0; color: #999; font-size: 12px;">暂无预算设置</div>';
      return;
    }

    let html = '';
    budgetStatus.forEach(item => {
      let statusClass = '';
      if (item.status === 'warning') statusClass = 'status-warning';
      if (item.status === 'over') statusClass = 'status-over';
      
      html += `
        <div class="budget-item ${statusClass}">
          <span class="category-name">${item.category}</span>
          <span class="category-amount">¥${item.spent.toFixed(2)} / ¥${item.budget.toFixed(2)}</span>
        </div>
      `;
    });

    listEl.innerHTML = html;
  }

  toggleSearchPanel() {
    const panel = document.getElementById('search-panel');
    panel.classList.toggle('hidden');
  }

  async doSearch() {
    const filters = {
      category: document.getElementById('search-category').value || null,
      type: document.getElementById('search-type').value || null,
      minAmount: document.getElementById('search-min-amount').value || null,
      maxAmount: document.getElementById('search-max-amount').value || null,
      keyword: document.getElementById('search-keyword').value || null,
      startDate: document.getElementById('search-start-date').value || null,
      endDate: document.getElementById('search-end-date').value || null
    };

    const records = await window.electronAPI.searchRecords(filters);
    this.isSearchMode = true;
    document.getElementById('selected-date-title').textContent = '搜索结果';
    this.renderRecords(records);
  }

  async clearSearch() {
    document.getElementById('search-category').value = '';
    document.getElementById('search-type').value = '';
    document.getElementById('search-min-amount').value = '';
    document.getElementById('search-max-amount').value = '';
    document.getElementById('search-keyword').value = '';
    document.getElementById('search-start-date').value = '';
    document.getElementById('search-end-date').value = '';
    
    this.isSearchMode = false;
    document.getElementById('selected-date-title').textContent = `${this.selectedDate} 明细`;
    await this.loadSelectedDateRecords();
  }

  async showStatsPanel() {
    const panel = document.getElementById('stats-panel');
    panel.classList.remove('hidden');

    const stats = await window.electronAPI.getMonthlyStats(this.currentYear, this.currentMonth);
    const categoryStats = await window.electronAPI.getCategoryStats(this.currentYear, this.currentMonth);

    document.getElementById('stat-income').textContent = `¥${stats.totalIncome}`;
    document.getElementById('stat-expense').textContent = `¥${stats.totalExpense}`;
    document.getElementById('stat-balance').textContent = `¥${stats.balance}`;
    document.getElementById('stat-count').textContent = stats.recordCount;

    const expenseStats = categoryStats.filter(s => s.type === 'expense');
    const incomeStats = categoryStats.filter(s => s.type === 'income');

    this.renderChart('expense-chart', expenseStats, 'expense');
    this.renderChart('income-chart', incomeStats, 'income');
  }

  hideStatsPanel() {
    document.getElementById('stats-panel').classList.add('hidden');
  }

  renderChart(containerId, stats, type) {
    const container = document.getElementById(containerId);
    
    if (stats.length === 0) {
      container.innerHTML = '<div class="empty-state">暂无数据</div>';
      return;
    }

    const maxAmount = Math.max(...stats.map(s => s.amount));
    const total = stats.reduce((sum, s) => sum + s.amount, 0);

    let html = '';
    stats.forEach(stat => {
      const percent = maxAmount > 0 ? (stat.amount / maxAmount) * 100 : 0;
      const percentOfTotal = total > 0 ? (stat.amount / total * 100).toFixed(1) : 0;
      
      html += `
        <div class="chart-item">
          <span class="chart-item-name">${stat.category}</span>
          <div class="chart-item-bar">
            <div class="chart-item-fill ${type}" style="width: ${percent}%;"></div>
          </div>
          <span class="chart-item-amount">¥${stat.amount.toFixed(2)} (${percentOfTotal}%)</span>
        </div>
      `;
    });

    container.innerHTML = html;
  }

  async showBudgetModal() {
    const modal = document.getElementById('budget-modal');
    const formList = document.getElementById('budget-form-list');
    
    const expenseCategories = this.categories.filter(c => c.type === 'expense');
    const budgets = await window.electronAPI.getBudgets(this.currentYear, this.currentMonth);

    let html = '';
    expenseCategories.forEach(cat => {
      const budget = budgets[cat.name] || 0;
      html += `
        <div class="budget-form-item">
          <label>${cat.name}</label>
          <input type="text" class="budget-input" data-category="${cat.name}" 
                 value="${budget > 0 ? budget.toFixed(2) : ''}" placeholder="设为0表示不设置">
        </div>
      `;
    });

    formList.innerHTML = html;
    modal.classList.remove('hidden');
  }

  hideBudgetModal() {
    document.getElementById('budget-modal').classList.add('hidden');
  }

  async saveBudgets() {
    const inputs = document.querySelectorAll('.budget-input');
    
    for (const input of inputs) {
      const category = input.dataset.category;
      const value = input.value;
      
      if (value && parseFloat(value) > 0) {
        await window.electronAPI.setBudget(this.currentYear, this.currentMonth, category, parseFloat(value));
      }
    }

    this.hideBudgetModal();
    await this.loadBudgetStatus();
  }

  async editRecord(id) {
    this.editingRecordId = id;
    
    let record = null;
    
    const allRecords = await window.electronAPI.searchRecords({
      category: null,
      type: null,
      minAmount: null,
      maxAmount: null,
      keyword: null,
      startDate: '2000-01-01',
      endDate: '2099-12-31'
    });
    
    record = allRecords.find(r => r.id === id);
    
    if (!record) {
      alert('找不到该记录');
      return;
    }

    document.getElementById('edit-amount').value = record.amount.toFixed(2);
    document.getElementById('edit-note').value = record.note || '';
    document.getElementById('edit-date').value = record.date.split('T')[0];

    this.switchEditType(record.type);
    document.getElementById('edit-category').value = record.category;

    document.getElementById('edit-modal').classList.remove('hidden');
  }

  hideEditModal() {
    document.getElementById('edit-modal').classList.add('hidden');
    this.editingRecordId = null;
  }

  async saveEdit() {
    const amount = document.getElementById('edit-amount').value;
    const note = document.getElementById('edit-note').value;
    const date = document.getElementById('edit-date').value;
    const category = document.getElementById('edit-category').value;
    
    const editType = document.querySelector('#edit-modal .type-switch button.active').id === 'edit-btn-expense' ? 'expense' : 'income';

    if (!amount || parseFloat(amount) <= 0) {
      alert('请输入有效的金额');
      return;
    }

    if (!category) {
      alert('请选择分类');
      return;
    }

    const record = {
      amount: parseFloat(amount),
      type: editType,
      category: category,
      note: note,
      date: date
    };

    await window.electronAPI.updateRecord(this.editingRecordId, record);

    this.hideEditModal();
    await this.renderCalendar();
    
    if (this.isSearchMode) {
      await this.doSearch();
    } else {
      await this.loadSelectedDateRecords();
    }
    
    await this.loadBudgetStatus();
  }

  deleteRecord(id) {
    this.deletingRecordId = id;
    document.getElementById('delete-modal').classList.remove('hidden');
  }

  hideDeleteModal() {
    document.getElementById('delete-modal').classList.add('hidden');
    this.deletingRecordId = null;
  }

  async confirmDelete() {
    if (this.deletingRecordId) {
      await window.electronAPI.deleteRecord(this.deletingRecordId);
      
      this.hideDeleteModal();
      await this.renderCalendar();
      
      if (this.isSearchMode) {
        await this.doSearch();
      } else {
        await this.loadSelectedDateRecords();
      }
      
      await this.loadBudgetStatus();
    }
  }

  async exportToCSV() {
    const result = await window.electronAPI.exportToCSV(this.currentYear, this.currentMonth);
    
    if (result.success) {
      alert(result.message);
    } else if (result.message !== '取消导出') {
      alert(result.message);
    }
  }
}

const app = new ExpenseTrackerApp();
