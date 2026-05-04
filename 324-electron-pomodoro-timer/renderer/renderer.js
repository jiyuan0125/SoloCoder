let state = {
  mode: 'work',
  currentTime: 25 * 60,
  totalTime: 25 * 60,
  isRunning: false,
  completedPomodoros: 0,
  currentTaskIndex: -1,
  tasks: [],
  statistics: {
    completedPomodoros: 0,
    totalFocusTime: 0,
    longestConsecutive: 0
  }
};

const elements = {
  circularTimer: document.getElementById('circularTimer'),
  timerProgress: document.querySelector('.timer-progress'),
  timerMode: document.getElementById('timerMode'),
  timerTime: document.getElementById('timerTime'),
  timerStatus: document.getElementById('timerStatus'),
  currentTaskName: document.getElementById('currentTaskName'),
  startBtn: document.getElementById('startBtn'),
  pauseBtn: document.getElementById('pauseBtn'),
  skipBtn: document.getElementById('skipBtn'),
  pomodoroDots: document.getElementById('pomodoroDots'),
  pomodoroLabel: document.getElementById('pomodoroLabel'),
  taskInput: document.getElementById('taskInput'),
  addTaskBtn: document.getElementById('addTaskBtn'),
  tasksList: document.getElementById('tasksList'),
  statPomodoros: document.getElementById('statPomodoros'),
  statTotalTime: document.getElementById('statTotalTime'),
  statLongestStreak: document.getElementById('statLongestStreak'),
  notification: document.getElementById('notification'),
  notificationMessage: document.getElementById('notificationMessage')
};

const CIRCUMFERENCE = 2 * Math.PI * 90;

function formatTime(seconds) {
  const mins = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
}

function showNotification(message) {
  elements.notificationMessage.textContent = message;
  elements.notification.classList.add('show');
  
  setTimeout(() => {
    elements.notification.classList.remove('show');
  }, 3000);
}

function updateTimerDisplay() {
  elements.timerTime.textContent = formatTime(state.currentTime);
  elements.timerStatus.textContent = state.isRunning ? '进行中' : '暂停';
  
  if (state.mode === 'work' && state.currentTime <= 60) {
    elements.timerTime.classList.add('warning');
  } else {
    elements.timerTime.classList.remove('warning');
  }
  
  const progress = 1 - (state.currentTime / state.totalTime);
  const dashoffset = CIRCUMFERENCE * (1 - progress);
  elements.timerProgress.style.strokeDashoffset = dashoffset;
  
  updateTimerColor();
  updateModeDisplay();
}

function updateTimerColor() {
  if (state.mode === 'work') {
    const progress = state.currentTime / state.totalTime;
    
    if (progress > 0.5) {
      elements.timerProgress.style.stroke = '#4CAF50';
    } else if (progress > 0.25) {
      elements.timerProgress.style.stroke = '#FF9800';
    } else {
      elements.timerProgress.style.stroke = '#f44336';
    }
  } else {
    elements.timerProgress.style.stroke = '#2196F3';
  }
}

function updateModeDisplay() {
  switch (state.mode) {
    case 'work':
      elements.timerMode.textContent = '专注时间';
      break;
    case 'shortBreak':
      elements.timerMode.textContent = '短休息';
      break;
    case 'longBreak':
      elements.timerMode.textContent = '长休息';
      break;
  }
}

function updateButtonStates() {
  if (state.isRunning) {
    elements.startBtn.style.display = 'none';
    elements.pauseBtn.style.display = 'inline-block';
  } else {
    elements.startBtn.style.display = 'inline-block';
    elements.pauseBtn.style.display = 'none';
  }
}

function updateCurrentTask() {
  if (state.currentTaskIndex >= 0 && state.currentTaskIndex < state.tasks.length) {
    const task = state.tasks[state.currentTaskIndex];
    elements.currentTaskName.textContent = task.name;
  } else {
    elements.currentTaskName.textContent = '无任务';
  }
}

function updatePomodoroDots() {
  const dots = elements.pomodoroDots.querySelectorAll('.pomodoro-dot');
  const currentCycle = state.completedPomodoros % 4;
  
  dots.forEach((dot, index) => {
    if (index < currentCycle) {
      dot.classList.add('completed');
    } else {
      dot.classList.remove('completed');
    }
  });
  
  elements.pomodoroLabel.textContent = `已完成 ${state.completedPomodoros} 个番茄钟`;
}

function renderTasks() {
  elements.tasksList.innerHTML = '';
  
  state.tasks.forEach((task, index) => {
    const taskElement = document.createElement('div');
    taskElement.className = `task-item ${task.completed ? 'completed' : ''} ${index === state.currentTaskIndex ? 'selected' : ''}`;
    taskElement.dataset.taskId = task.id;
    taskElement.dataset.index = index;
    taskElement.draggable = true;
    
    taskElement.innerHTML = `
      <div class="task-handle">
        <span></span>
        <span></span>
        <span></span>
      </div>
      <input type="checkbox" class="task-checkbox" ${task.completed ? 'checked' : ''}>
      <span class="task-name">${escapeHtml(task.name)}</span>
      <div class="task-pomodoros">
        <span class="pomodoro-count">${task.pomodorosCompleted} 🍅</span>
      </div>
      <button class="task-delete">×</button>
    `;
    
    setupTaskEventListeners(taskElement, task, index);
    elements.tasksList.appendChild(taskElement);
  });
}

function setupTaskEventListeners(taskElement, task, index) {
  const checkbox = taskElement.querySelector('.task-checkbox');
  const deleteBtn = taskElement.querySelector('.task-delete');
  
  checkbox.addEventListener('change', async (e) => {
    e.stopPropagation();
    const result = await window.electronAPI.toggleTaskCompletion(task.id);
    if (result.success) {
      state.tasks = result.tasks;
      renderTasks();
    }
  });
  
  deleteBtn.addEventListener('click', async (e) => {
    e.stopPropagation();
    const result = await window.electronAPI.deleteTask(task.id);
    if (result.success) {
      state.tasks = result.tasks;
      state.currentTaskIndex = result.currentTaskIndex;
      renderTasks();
      updateCurrentTask();
      showNotification('任务已删除');
    }
  });
  
  taskElement.addEventListener('click', async (e) => {
    if (e.target === checkbox || e.target === deleteBtn || e.target.closest('.task-handle')) {
      return;
    }
    
    const result = await window.electronAPI.selectTask(index);
    if (result.success) {
      state.currentTaskIndex = result.currentTaskIndex;
      renderTasks();
      updateCurrentTask();
    }
  });
  
  taskElement.addEventListener('dragstart', handleDragStart);
  taskElement.addEventListener('dragend', handleDragEnd);
  taskElement.addEventListener('dragover', handleDragOver);
  taskElement.addEventListener('drop', handleDrop);
}

let draggedElement = null;

function handleDragStart(e) {
  draggedElement = this;
  this.classList.add('dragging');
  e.dataTransfer.effectAllowed = 'move';
  e.dataTransfer.setData('text/plain', this.dataset.taskId);
}

function handleDragEnd(e) {
  this.classList.remove('dragging');
  draggedElement = null;
}

function handleDragOver(e) {
  e.preventDefault();
  e.dataTransfer.dropEffect = 'move';
}

async function handleDrop(e) {
  e.preventDefault();
  
  if (!draggedElement || draggedElement === this) return;
  
  const fromIndex = parseInt(draggedElement.dataset.index);
  const toIndex = parseInt(this.dataset.index);
  
  if (fromIndex === toIndex) return;
  
  const newOrder = state.tasks.map(t => t.id);
  const [removed] = newOrder.splice(fromIndex, 1);
  newOrder.splice(toIndex, 0, removed);
  
  const result = await window.electronAPI.reorderTasks(newOrder);
  if (result.success) {
    state.tasks = result.tasks;
    state.currentTaskIndex = result.currentTaskIndex;
    renderTasks();
    updateCurrentTask();
  }
}

function updateStatistics() {
  elements.statPomodoros.textContent = state.statistics.completedPomodoros;
  
  const totalMinutes = Math.floor(state.statistics.totalFocusTime / 60);
  if (totalMinutes >= 60) {
    const hours = Math.floor(totalMinutes / 60);
    const mins = totalMinutes % 60;
    elements.statTotalTime.textContent = `${hours}小时${mins > 0 ? mins + '分钟' : ''}`;
  } else {
    elements.statTotalTime.textContent = `${totalMinutes}分钟`;
  }
  
  elements.statLongestStreak.textContent = state.statistics.longestConsecutive;
}

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

async function initializeApp() {
  elements.timerProgress.style.strokeDasharray = CIRCUMFERENCE;
  elements.timerProgress.style.strokeDashoffset = 0;
  
  const initialState = await window.electronAPI.getInitialState();
  
  state = {
    ...state,
    ...initialState
  };
  
  updateTimerDisplay();
  updateButtonStates();
  updateCurrentTask();
  updatePomodoroDots();
  renderTasks();
  updateStatistics();
  
  setupEventListeners();
  setupIPCHandlers();
}

function setupEventListeners() {
  elements.startBtn.addEventListener('click', async () => {
    await window.electronAPI.startTimer();
    state.isRunning = true;
    updateButtonStates();
    updateTimerDisplay();
  });
  
  elements.pauseBtn.addEventListener('click', async () => {
    await window.electronAPI.pauseTimer();
    state.isRunning = false;
    updateButtonStates();
    updateTimerDisplay();
  });
  
  elements.skipBtn.addEventListener('click', async () => {
    if (state.mode !== 'work') {
      showNotification('提前结束休息');
    }
    await window.electronAPI.skipSession();
  });
  
  elements.addTaskBtn.addEventListener('click', addTask);
  
  elements.taskInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
      addTask();
    }
  });
}

async function addTask() {
  const taskName = elements.taskInput.value.trim();
  
  if (!taskName) {
    showNotification('任务名称不能为空');
    return;
  }
  
  if (taskName.length > 50) {
    showNotification('任务名称不能超过50个字符');
    return;
  }
  
  const result = await window.electronAPI.addTask({ name: taskName });
  
  if (result.success) {
    state.tasks = result.tasks;
    elements.taskInput.value = '';
    renderTasks();
    showNotification('任务添加成功');
  } else {
    showNotification(result.error);
  }
}

function setupIPCHandlers() {
  window.electronAPI.onTimerTick((data) => {
    state.mode = data.mode;
    state.currentTime = data.currentTime;
    state.totalTime = data.totalTime;
    state.isRunning = data.isRunning;
    updateTimerDisplay();
  });
  
  window.electronAPI.onTimerStateChanged((data) => {
    state.isRunning = data.isRunning;
    updateButtonStates();
    updateTimerDisplay();
  });
  
  window.electronAPI.onTimerComplete((data) => {
    state.mode = data.mode;
    state.currentTime = data.totalTime || getModeDuration(data.mode);
    state.totalTime = getModeDuration(data.mode);
    state.isRunning = false;
    state.completedPomodoros = data.completedPomodoros;
    state.tasks = data.tasks;
    state.statistics = data.statistics;
    
    updateTimerDisplay();
    updateButtonStates();
    updatePomodoroDots();
    renderTasks();
    updateStatistics();
  });
  
  window.electronAPI.onModeChanged((data) => {
    state.mode = data.mode;
    state.currentTime = data.currentTime;
    state.totalTime = data.totalTime;
    state.isRunning = false;
    
    updateTimerDisplay();
    updateButtonStates();
  });
  
  window.electronAPI.onSessionSkipped((data) => {
    const previousMode = data.previousMode;
    const newMode = data.newMode;
    
    state.mode = newMode;
    state.currentTime = getModeDuration(newMode);
    state.totalTime = getModeDuration(newMode);
    state.isRunning = false;
    
    updateTimerDisplay();
    updateButtonStates();
  });
  
  window.electronAPI.onInitialStateLoaded((data) => {
    state = {
      ...state,
      mode: data.mode,
      currentTime: data.currentTime,
      totalTime: data.totalTime,
      completedPomodoros: data.completedPomodoros,
      currentTaskIndex: data.currentTaskIndex,
      tasks: data.tasks,
      statistics: data.statistics,
      isRunning: data.isRunning
    };
    
    updateTimerDisplay();
    updateButtonStates();
    updateCurrentTask();
    updatePomodoroDots();
    renderTasks();
    updateStatistics();
  });
  
  window.electronAPI.onBreakSkipped(() => {
    showNotification('提前结束休息');
  });
}

function getModeDuration(mode) {
  switch (mode) {
    case 'work':
      return 25 * 60;
    case 'shortBreak':
      return 5 * 60;
    case 'longBreak':
      return 15 * 60;
    default:
      return 25 * 60;
  }
}

document.addEventListener('DOMContentLoaded', initializeApp);
