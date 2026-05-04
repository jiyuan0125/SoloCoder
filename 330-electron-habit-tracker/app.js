const electronAPI = window.electronAPI;

let currentYear = new Date().getFullYear();
let currentMonth = new Date().getMonth() + 1;
let habits = [];
let checkinData = {};

const monthNames = ['一月', '二月', '三月', '四月', '五月', '六月', 
                    '七月', '八月', '九月', '十月', '十一月', '十二月'];
const weekDayNames = ['日', '一', '二', '三', '四', '五', '六'];

document.addEventListener('DOMContentLoaded', async () => {
  await init();
  setupEventListeners();
});

async function init() {
  habits = await electronAPI.getHabits();
  checkinData = await electronAPI.getCheckinData(currentYear, currentMonth);
  
  updateMonthDisplay();
  renderHeatmap();
  renderHabitsList();
  renderStats();
}

function setupEventListeners() {
  document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      switchTab(btn.dataset.tab);
    });
  });

  document.getElementById('prevMonth').addEventListener('click', () => {
    navigateMonth(-1);
  });
  document.getElementById('nextMonth').addEventListener('click', () => {
    navigateMonth(1);
  });
  document.getElementById('prevMonthStats').addEventListener('click', () => {
    navigateMonth(-1);
  });
  document.getElementById('nextMonthStats').addEventListener('click', () => {
    navigateMonth(1);
  });

  document.getElementById('addHabitBtn').addEventListener('click', () => {
    openAddHabitModal();
  });
  document.getElementById('cancelAddHabit').addEventListener('click', () => {
    closeAddHabitModal();
  });
  document.getElementById('closeCheckinModal').addEventListener('click', () => {
    closeCheckinModal();
  });

  document.querySelectorAll('input[name="frequency"]').forEach(radio => {
    radio.addEventListener('change', handleFrequencyChange);
  });

  document.getElementById('addHabitForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    await handleAddHabit();
  });

  document.getElementById('exportBtn').addEventListener('click', async () => {
    const result = await electronAPI.exportCSV(currentYear, currentMonth);
    if (result.success) {
      alert(`数据已导出到: ${result.path}`);
    } else {
      alert('导出失败: ' + (result.error || '未知错误'));
    }
  });

  document.querySelectorAll('.modal').forEach(modal => {
    modal.addEventListener('click', (e) => {
      if (e.target === modal) {
        modal.classList.remove('active');
      }
    });
  });
}

function switchTab(tabName) {
  document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.classList.toggle('active', btn.dataset.tab === tabName);
  });
  document.querySelectorAll('.tab-content').forEach(content => {
    content.classList.toggle('active', content.id === `${tabName}-tab`);
  });
}

async function navigateMonth(delta) {
  currentMonth += delta;
  if (currentMonth > 12) {
    currentMonth = 1;
    currentYear++;
  } else if (currentMonth < 1) {
    currentMonth = 12;
    currentYear--;
  }
  
  checkinData = await electronAPI.getCheckinData(currentYear, currentMonth);
  updateMonthDisplay();
  renderHeatmap();
  renderStats();
}

function updateMonthDisplay() {
  const monthText = `${currentYear}年 ${monthNames[currentMonth - 1]}`;
  document.getElementById('currentMonth').textContent = monthText;
  document.getElementById('statsMonth').textContent = monthText;
}

function getDaysInMonth(year, month) {
  return new Date(year, month, 0).getDate();
}

function getFirstDayOfMonth(year, month) {
  return new Date(year, month - 1, 1).getDay();
}

async function renderHeatmap() {
  const container = document.getElementById('heatmap');
  container.innerHTML = '';
  
  const daysInMonth = getDaysInMonth(currentYear, currentMonth);
  const firstDay = getFirstDayOfMonth(currentYear, currentMonth);
  const today = new Date();
  const todayStr = today.toISOString().split('T')[0];
  
  const weekLabels = ['日', '一', '二', '三', '四', '五', '六'];
  for (let week = 0; week < 6; week++) {
    const weekLabel = document.createElement('div');
    weekLabel.className = 'heatmap-week-label';
    weekLabel.textContent = weekLabels[week % 7];
    container.appendChild(weekLabel);
    
    for (let day = 0; day < 7; day++) {
      const dayIndex = week * 7 + day - firstDay + 1;
      
      const dayElement = document.createElement('div');
      dayElement.className = 'heatmap-day';
      
      if (dayIndex >= 1 && dayIndex <= daysInMonth) {
        const dateStr = `${currentYear}-${String(currentMonth).padStart(2, '0')}-${String(dayIndex).padStart(2, '0')}`;
        dayElement.dataset.date = dateStr;
        
        if (dateStr === todayStr) {
          dayElement.classList.add('today');
        }
        
        const completion = getDayCompletion(dateStr);
        dayElement.style.backgroundColor = getHeatmapColor(completion);
        
        dayElement.addEventListener('click', () => {
          openCheckinModal(dateStr);
        });
      } else {
        dayElement.style.visibility = 'hidden';
      }
      
      container.appendChild(dayElement);
    }
  }
}

function getDayCompletion(dateStr) {
  if (!checkinData[dateStr]) return 0;
  
  const requiredHabits = habits.filter(habit => shouldCheckinOnDate(habit, dateStr));
  if (requiredHabits.length === 0) return 0;
  
  const completedCount = requiredHabits.filter(habit => 
    checkinData[dateStr][habit.id] === true
  ).length;
  
  return completedCount / requiredHabits.length;
}

function getHeatmapColor(completion) {
  if (completion === 0) return '#ebedf0';
  if (completion < 0.5) return '#9be9a8';
  if (completion < 1) return '#40c463';
  return '#216e39';
}

function shouldCheckinOnDate(habit, dateStr) {
  const date = new Date(dateStr);
  const dayOfWeek = date.getDay();
  
  switch (habit.frequency) {
    case 'daily':
      return true;
    case 'weekly':
      return habit.weeklyDays && habit.weeklyDays.includes(dayOfWeek);
    case 'custom':
      const createdDate = new Date(habit.createdAt);
      const diffDays = Math.floor((date - createdDate) / (1000 * 60 * 60 * 24));
      return diffDays >= 0 && diffDays % (habit.customInterval || 1) === 0;
    default:
      return true;
  }
}

async function renderHabitsList() {
  const container = document.getElementById('habitsList');
  const emptyState = document.getElementById('noHabits');
  
  if (habits.length === 0) {
    container.innerHTML = '';
    emptyState.style.display = 'block';
    return;
  }
  
  emptyState.style.display = 'none';
  
  const streakPromises = habits.map(habit => 
    electronAPI.getStreaks(habit.id)
  );
  const streaksList = await Promise.all(streakPromises);
  
  container.innerHTML = habits.map((habit, index) => {
    const streaks = streaksList[index];
    const trophyCount = Math.floor(streaks.current / 7);
    const trophies = trophyCount > 0 ? '<span class="trophy">' + '🏆'.repeat(trophyCount) + '</span>' : '';
    
    let frequencyText = '';
    switch (habit.frequency) {
      case 'daily':
        frequencyText = '每天';
        break;
      case 'weekly':
        if (habit.weeklyDays && habit.weeklyDays.length > 0) {
          frequencyText = '每周' + habit.weeklyDays.map(d => weekDayNames[d]).join('、');
        }
        break;
      case 'custom':
        frequencyText = `每${habit.customInterval || 1}天`;
        break;
    }
    
    return `
      <div class="habit-card" data-id="${habit.id}">
        <div class="habit-icon">${habit.icon}</div>
        <div class="habit-info">
          <div class="habit-name">${habit.name}${trophies}</div>
          <div class="habit-frequency">${frequencyText}</div>
        </div>
        <div class="habit-streaks">
          <div class="streak-item">
            <div class="streak-value">${streaks.current}</div>
            <div class="streak-label">当前连续</div>
          </div>
          <div class="streak-item">
            <div class="streak-value">${streaks.max}</div>
            <div class="streak-label">历史最长</div>
          </div>
        </div>
        <div class="habit-actions">
          <button class="btn btn-danger" onclick="deleteHabit('${habit.id}')">删除</button>
        </div>
      </div>
    `;
  }).join('');
}

async function deleteHabit(habitId) {
  const result = await electronAPI.deleteHabit(habitId);
  if (result.success) {
    habits = habits.filter(h => h.id !== habitId);
    renderHabitsList();
    renderHeatmap();
    renderStats();
  }
}

async function renderStats() {
  const stats = await electronAPI.getMonthlyStats(currentYear, currentMonth);
  
  document.getElementById('monthlyRate').textContent = Math.round(stats.monthlyRate) + '%';
  document.getElementById('totalDays').textContent = stats.totalRequiredDays;
  document.getElementById('completedDays').textContent = stats.completedDays;
  
  const chartContainer = document.getElementById('barChart');
  
  if (stats.habitStats.length === 0) {
    chartContainer.innerHTML = '<p style="color: #999; text-align: center; padding: 40px;">暂无统计数据</p>';
    return;
  }
  
  chartContainer.innerHTML = stats.habitStats.map(stat => `
    <div class="chart-row">
      <div class="chart-label">${stat.icon} ${stat.name}</div>
      <div class="chart-bar-container">
        <div class="chart-bar" style="width: ${Math.max(stat.rate, 2)}%;"></div>
        <span class="chart-percentage">${Math.round(stat.rate)}%</span>
      </div>
    </div>
  `).join('');
}

function openAddHabitModal() {
  document.getElementById('addHabitModal').classList.add('active');
  document.getElementById('addHabitForm').reset();
  handleFrequencyChange();
  
  document.querySelectorAll('.weekday').forEach(cb => cb.checked = false);
  document.getElementById('customInterval').value = 1;
}

function closeAddHabitModal() {
  document.getElementById('addHabitModal').classList.remove('active');
}

function handleFrequencyChange() {
  const selected = document.querySelector('input[name="frequency"]:checked').value;
  
  document.getElementById('weeklyOptions').classList.toggle('hidden', selected !== 'weekly');
  document.getElementById('customOptions').classList.toggle('hidden', selected !== 'custom');
}

async function handleAddHabit() {
  const name = document.getElementById('habitName').value.trim();
  const icon = document.getElementById('habitIcon').value;
  const frequency = document.querySelector('input[name="frequency"]:checked').value;
  
  let weeklyDays = [0, 1, 2, 3, 4, 5, 6];
  if (frequency === 'weekly') {
    weeklyDays = Array.from(document.querySelectorAll('.weekday:checked'))
      .map(cb => parseInt(cb.value));
    if (weeklyDays.length === 0) {
      alert('请至少选择一天');
      return;
    }
  }
  
  const customInterval = parseInt(document.getElementById('customInterval').value) || 1;
  
  const result = await electronAPI.addHabit({
    name,
    icon,
    frequency,
    weeklyDays,
    customInterval
  });
  
  if (result.success) {
    habits.push(result.habit);
    closeAddHabitModal();
    renderHabitsList();
    renderHeatmap();
    renderStats();
  }
}

async function openCheckinModal(dateStr) {
  const date = new Date(dateStr);
  const formattedDate = `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日 周${weekDayNames[date.getDay()]}`;
  document.getElementById('checkinDateTitle').textContent = formattedDate;
  
  const listContainer = document.getElementById('checkinList');
  
  if (habits.length === 0) {
    listContainer.innerHTML = '<p style="color: #999; text-align: center; padding: 20px;">还没有添加习惯</p>';
  } else {
    listContainer.innerHTML = habits.map(habit => {
      const isRequired = shouldCheckinOnDate(habit, dateStr);
      const completed = checkinData[dateStr] && checkinData[dateStr][habit.id] === true;
      const optionalClass = isRequired ? '' : 'optional';
      const optionalLabel = isRequired ? '' : '<span class="optional-label">(无需打卡)</span>';
      
      return `
        <div class="checkin-item ${optionalClass}">
          <input type="checkbox" 
                 data-habit-id="${habit.id}" 
                 data-date="${dateStr}"
                 ${completed ? 'checked' : ''}
                 ${!isRequired ? 'disabled' : ''}>
          <div class="habit-icon">${habit.icon}</div>
          <div class="habit-name">${habit.name}${optionalLabel}</div>
        </div>
      `;
    }).join('');
    
    listContainer.querySelectorAll('input[type="checkbox"]:not([disabled])').forEach(checkbox => {
      checkbox.addEventListener('change', async (e) => {
        const habitId = e.target.dataset.habitId;
        const date = e.target.dataset.date;
        const completed = e.target.checked;
        
        await electronAPI.saveCheckin({ date, habitId, completed });
        
        checkinData = await electronAPI.getCheckinData(currentYear, currentMonth);
        renderHeatmap();
        renderStats();
      });
    });
  }
  
  document.getElementById('checkinModal').classList.add('active');
}

function closeCheckinModal() {
  document.getElementById('checkinModal').classList.remove('active');
}

window.deleteHabit = deleteHabit;
