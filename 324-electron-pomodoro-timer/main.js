const { app, BrowserWindow, ipcMain, Notification, Tray, Menu } = require('electron');
const path = require('path');
const fs = require('fs');

let mainWindow;
let tray;
let timer = null;
let isRunning = false;

const CONFIG = {
  workDuration: 25 * 60,
  shortBreakDuration: 5 * 60,
  longBreakDuration: 15 * 60,
  pomodorosBeforeLongBreak: 4
};

let state = {
  mode: 'work',
  currentTime: CONFIG.workDuration,
  totalTime: CONFIG.workDuration,
  completedPomodoros: 0,
  currentTaskIndex: -1,
  tasks: [],
  statistics: {
    completedPomodoros: 0,
    totalFocusTime: 0,
    longestConsecutive: 0,
    currentConsecutive: 0
  },
  isTimerRunning: false,
  timerStartTimestamp: null,
  wasTimerRunning: false,
  wasInBreak: false,
  savedRemainingTime: null
};

const DATA_DIR = path.join(app.getPath('userData'), 'pomodoro-data');

function getTodayDateString() {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;
}

function getTodayDataPath() {
  if (!fs.existsSync(DATA_DIR)) {
    fs.mkdirSync(DATA_DIR, { recursive: true });
  }
  return path.join(DATA_DIR, `${getTodayDateString()}.json`);
}

function loadTodayData() {
  const dataPath = getTodayDataPath();
  if (fs.existsSync(dataPath)) {
    try {
      const data = JSON.parse(fs.readFileSync(dataPath, 'utf8'));
      state.tasks = data.tasks || [];
      state.statistics = data.statistics || {
        completedPomodoros: 0,
        totalFocusTime: 0,
        longestConsecutive: 0,
        currentConsecutive: 0
      };
      state.completedPomodoros = data.completedPomodoros || 0;
      
      if (data.savedState) {
        state.mode = data.savedState.mode || 'work';
        state.currentTime = data.savedState.currentTime || getModeDuration(state.mode);
        state.totalTime = getModeDuration(state.mode);
        state.currentTaskIndex = data.savedState.currentTaskIndex ?? -1;
        state.isTimerRunning = data.savedState.isTimerRunning || false;
        state.savedRemainingTime = data.savedState.savedRemainingTime;
        state.wasTimerRunning = state.isTimerRunning;
        state.wasInBreak = state.mode !== 'work';
      }
      
      return true;
    } catch (error) {
      console.error('Failed to load data:', error);
      return false;
    }
  }
  return false;
}

function saveTodayData() {
  const dataPath = getTodayDataPath();
  const dataToSave = {
    tasks: state.tasks,
    statistics: state.statistics,
    completedPomodoros: state.completedPomodoros,
    savedState: {
      mode: state.mode,
      currentTime: state.currentTime,
      currentTaskIndex: state.currentTaskIndex,
      isTimerRunning: state.isTimerRunning,
      savedRemainingTime: state.savedRemainingTime
    }
  };
  
  try {
    fs.writeFileSync(dataPath, JSON.stringify(dataToSave, null, 2));
  } catch (error) {
    console.error('Failed to save data:', error);
  }
}

function getModeDuration(mode) {
  switch (mode) {
    case 'work':
      return CONFIG.workDuration;
    case 'shortBreak':
      return CONFIG.shortBreakDuration;
    case 'longBreak':
      return CONFIG.longBreakDuration;
    default:
      return CONFIG.workDuration;
  }
}

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 800,
    height: 600,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'preload.js')
    },
    icon: path.join(__dirname, 'icon.png'),
    title: '番茄钟',
    resizable: true,
    minimumWidth: 600,
    minimumHeight: 500
  });

  mainWindow.loadFile(path.join(__dirname, 'renderer', 'index.html'));

  mainWindow.on('close', (event) => {
    if (isRunning) {
      event.preventDefault();
      mainWindow.hide();
    }
  });

  mainWindow.on('closed', () => {
    mainWindow = null;
  });
}

function createTray() {
  const iconPath = path.join(__dirname, 'icon.png');
  tray = new Tray(fs.existsSync(iconPath) ? iconPath : path.join(__dirname, 'assets', 'icon.png'));
  
  const contextMenu = Menu.buildFromTemplate([
    {
      label: '显示/隐藏',
      click: () => {
        if (mainWindow) {
          if (mainWindow.isVisible()) {
            mainWindow.hide();
          } else {
            mainWindow.show();
            mainWindow.focus();
          }
        }
      }
    },
    {
      label: '开始/暂停',
      click: () => {
        if (isRunning) {
          pauseTimer();
        } else {
          startTimer();
        }
      }
    },
    {
      label: '退出',
      click: () => {
        if (timer) {
          clearInterval(timer);
          timer = null;
        }
        saveTodayData();
        app.quit();
      }
    }
  ]);
  
  tray.setToolTip('番茄钟');
  tray.setContextMenu(contextMenu);
  
  tray.on('double-click', () => {
    if (mainWindow) {
      if (mainWindow.isVisible()) {
        mainWindow.hide();
      } else {
        mainWindow.show();
        mainWindow.focus();
      }
    }
  });
}

function startTimer() {
  if (isRunning) return;
  
  isRunning = true;
  state.isTimerRunning = true;
  
  if (state.mode === 'work') {
    state.timerStartTimestamp = Date.now();
  }
  
  if (mainWindow) {
    mainWindow.webContents.send('timer-state-changed', { isRunning: true });
  }
  
  timer = setInterval(() => {
    state.currentTime--;
    saveTodayData();
    
    if (state.currentTime <= 0) {
      timerComplete();
    } else {
      if (mainWindow) {
        mainWindow.webContents.send('timer-tick', {
          mode: state.mode,
          currentTime: state.currentTime,
          totalTime: state.totalTime,
          isRunning: isRunning
        });
      }
      updateTrayTitle();
    }
  }, 1000);
}

function pauseTimer() {
  if (!isRunning) return;
  
  isRunning = false;
  state.isTimerRunning = false;
  
  if (timer) {
    clearInterval(timer);
    timer = null;
  }
  
  if (state.mode === 'work' && state.timerStartTimestamp) {
    const elapsedSeconds = Math.floor((Date.now() - state.timerStartTimestamp) / 1000);
    state.statistics.totalFocusTime += elapsedSeconds;
    state.timerStartTimestamp = null;
  }
  
  saveTodayData();
  
  if (mainWindow) {
    mainWindow.webContents.send('timer-state-changed', { isRunning: false });
  }
}

function timerComplete() {
  if (timer) {
    clearInterval(timer);
    timer = null;
  }
  
  isRunning = false;
  state.isTimerRunning = false;
  
  if (state.mode === 'work') {
    state.completedPomodoros++;
    state.statistics.completedPomodoros++;
    state.statistics.currentConsecutive++;
    
    if (state.statistics.currentConsecutive > state.statistics.longestConsecutive) {
      state.statistics.longestConsecutive = state.statistics.currentConsecutive;
    }
    
    if (state.currentTaskIndex >= 0 && state.currentTaskIndex < state.tasks.length) {
      state.tasks[state.currentTaskIndex].pomodorosCompleted++;
    }
    
    showNotification('番茄钟完成！', '太棒了！休息一下吧。');
    
    if (state.completedPomodoros % CONFIG.pomodorosBeforeLongBreak === 0) {
      switchToMode('longBreak');
    } else {
      switchToMode('shortBreak');
    }
  } else {
    showNotification('休息结束', '准备好开始下一个番茄钟了吗？', true);
    switchToMode('work');
  }
  
  saveTodayData();
  
  if (mainWindow) {
    mainWindow.webContents.send('timer-complete', {
      mode: state.mode,
      completedPomodoros: state.completedPomodoros,
      tasks: state.tasks,
      statistics: state.statistics
    });
  }
}

function skipCurrentSession() {
  if (timer) {
    clearInterval(timer);
    timer = null;
  }
  
  isRunning = false;
  state.isTimerRunning = false;
  
  const previousMode = state.mode;
  
  if (previousMode === 'work') {
    if (state.currentTaskIndex >= 0 && state.currentTaskIndex < state.tasks.length) {
      state.tasks[state.currentTaskIndex].pomodorosCompleted++;
    }
    state.statistics.currentConsecutive = 0;
    
    if (state.completedPomodoros % CONFIG.pomodorosBeforeLongBreak === 0) {
      switchToMode('longBreak');
    } else {
      switchToMode('shortBreak');
    }
  } else {
    if (mainWindow) {
      mainWindow.webContents.send('break-skipped');
    }
    switchToMode('work');
  }
  
  saveTodayData();
  
  if (mainWindow) {
    mainWindow.webContents.send('session-skipped', {
      previousMode: previousMode,
      newMode: state.mode
    });
  }
}

function switchToMode(mode) {
  state.mode = mode;
  state.totalTime = getModeDuration(mode);
  state.currentTime = state.totalTime;
  state.isTimerRunning = false;
  isRunning = false;
  
  if (mainWindow) {
    mainWindow.webContents.send('mode-changed', {
      mode: state.mode,
      currentTime: state.currentTime,
      totalTime: state.totalTime
    });
  }
  
  updateTrayTitle();
  saveTodayData();
}

function showNotification(title, body, clickToResume = false) {
  const notification = new Notification({
    title: title,
    body: body,
    silent: false
  });
  
  if (clickToResume) {
    notification.on('click', () => {
      if (mainWindow) {
        mainWindow.show();
        mainWindow.focus();
      }
      if (state.mode === 'work' && !isRunning) {
        startTimer();
      }
    });
  }
  
  notification.show();
}

function updateTrayTitle() {
  if (!tray) return;
  
  const minutes = Math.floor(state.currentTime / 60);
  const seconds = state.currentTime % 60;
  const timeString = `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
  
  const modeIndicator = state.mode === 'work' ? '专注' : '休息';
  tray.setTitle(`${modeIndicator} ${timeString}`);
}

function initializeApp() {
  const hasSavedData = loadTodayData();
  state.totalTime = getModeDuration(state.mode);
  
  if (!hasSavedData) {
    state.tasks = [];
    state.statistics = {
      completedPomodoros: 0,
      totalFocusTime: 0,
      longestConsecutive: 0,
      currentConsecutive: 0
    };
    state.completedPomodoros = 0;
    state.currentTaskIndex = -1;
  }
  
  createWindow();
  createTray();
  
  if (mainWindow && hasSavedData) {
    mainWindow.webContents.once('dom-ready', () => {
      mainWindow.webContents.send('initial-state-loaded', {
        mode: state.mode,
        currentTime: state.currentTime,
        totalTime: state.totalTime,
        completedPomodoros: state.completedPomodoros,
        currentTaskIndex: state.currentTaskIndex,
        tasks: state.tasks,
        statistics: state.statistics,
        isRunning: isRunning,
        wasTimerRunning: state.wasTimerRunning,
        wasInBreak: state.wasInBreak,
        savedRemainingTime: state.savedRemainingTime
      });
    });
  }
}

app.whenReady().then(initializeApp);

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    if (timer) {
      clearInterval(timer);
      timer = null;
    }
    saveTodayData();
    app.quit();
  }
});

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    createWindow();
  } else if (mainWindow) {
    mainWindow.show();
    mainWindow.focus();
  }
});

app.on('before-quit', () => {
  saveTodayData();
});

ipcMain.handle('get-initial-state', () => {
  return {
    mode: state.mode,
    currentTime: state.currentTime,
    totalTime: state.totalTime,
    completedPomodoros: state.completedPomodoros,
    currentTaskIndex: state.currentTaskIndex,
    tasks: state.tasks,
    statistics: state.statistics,
    isRunning: isRunning
  };
});

ipcMain.handle('start-timer', () => {
  startTimer();
  return { success: true };
});

ipcMain.handle('pause-timer', () => {
  pauseTimer();
  return { success: true };
});

ipcMain.handle('skip-session', () => {
  skipCurrentSession();
  return { success: true };
});

ipcMain.handle('add-task', (event, taskData) => {
  if (!taskData.name || taskData.name.trim() === '') {
    return { success: false, error: '任务名称不能为空' };
  }
  
  const taskName = taskData.name.trim().substring(0, 50);
  const newTask = {
    id: Date.now(),
    name: taskName,
    completed: false,
    pomodorosCompleted: 0,
    createdAt: Date.now()
  };
  
  state.tasks.push(newTask);
  saveTodayData();
  
  return { success: true, task: newTask, tasks: state.tasks };
});

ipcMain.handle('delete-task', (event, taskId) => {
  const taskIndex = state.tasks.findIndex(t => t.id === taskId);
  
  if (taskIndex === -1) {
    return { success: false, error: '任务不存在' };
  }
  
  if (state.currentTaskIndex === taskIndex) {
    state.currentTaskIndex = -1;
  } else if (taskIndex < state.currentTaskIndex) {
    state.currentTaskIndex--;
  }
  
  state.tasks.splice(taskIndex, 1);
  saveTodayData();
  
  return { success: true, tasks: state.tasks, currentTaskIndex: state.currentTaskIndex };
});

ipcMain.handle('toggle-task-completion', (event, taskId) => {
  const taskIndex = state.tasks.findIndex(t => t.id === taskId);
  
  if (taskIndex === -1) {
    return { success: false, error: '任务不存在' };
  }
  
  state.tasks[taskIndex].completed = !state.tasks[taskIndex].completed;
  saveTodayData();
  
  return { success: true, task: state.tasks[taskIndex], tasks: state.tasks };
});

ipcMain.handle('select-task', (event, taskIndex) => {
  if (taskIndex < -1 || taskIndex >= state.tasks.length) {
    return { success: false, error: '无效的任务索引' };
  }
  
  state.currentTaskIndex = taskIndex;
  saveTodayData();
  
  const selectedTask = taskIndex >= 0 ? state.tasks[taskIndex] : null;
  return { 
    success: true, 
    currentTaskIndex: state.currentTaskIndex,
    selectedTask: selectedTask
  };
});

ipcMain.handle('reorder-tasks', (event, newOrder) => {
  if (!Array.isArray(newOrder) || newOrder.length !== state.tasks.length) {
    return { success: false, error: '无效的排序数据' };
  }
  
  const reorderedTasks = [];
  for (const taskId of newOrder) {
    const task = state.tasks.find(t => t.id === taskId);
    if (task) {
      reorderedTasks.push(task);
    }
  }
  
  if (state.currentTaskIndex >= 0 && state.currentTaskIndex < state.tasks.length) {
    const currentTaskId = state.tasks[state.currentTaskIndex].id;
    state.currentTaskIndex = reorderedTasks.findIndex(t => t.id === currentTaskId);
  }
  
  state.tasks = reorderedTasks;
  saveTodayData();
  
  return { success: true, tasks: state.tasks, currentTaskIndex: state.currentTaskIndex };
});

ipcMain.handle('get-statistics', () => {
  return {
    completedPomodoros: state.statistics.completedPomodoros,
    totalFocusTime: state.statistics.totalFocusTime,
    longestConsecutive: state.statistics.longestConsecutive
  };
});

ipcMain.handle('set-mode', (event, mode) => {
  if (!['work', 'shortBreak', 'longBreak'].includes(mode)) {
    return { success: false, error: '无效的模式' };
  }
  
  if (isRunning) {
    pauseTimer();
  }
  
  switchToMode(mode);
  return { success: true };
});
