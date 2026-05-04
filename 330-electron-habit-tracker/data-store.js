const fs = require('fs');
const path = require('path');
const { Parser } = require('json2csv');

const userDataPath = process.env.HOME || process.env.USERPROFILE;
const dataDir = path.join(userDataPath, '.habit-tracker');
const habitsFile = path.join(dataDir, 'habits.json');

function ensureDataDir() {
  if (!fs.existsSync(dataDir)) {
    fs.mkdirSync(dataDir, { recursive: true });
  }
}

function getCheckinDataPath(year, month) {
  return path.join(dataDir, `checkin-${year}-${String(month).padStart(2, '0')}.json`);
}

function getHabits() {
  ensureDataDir();
  if (!fs.existsSync(habitsFile)) {
    return [];
  }
  try {
    const data = fs.readFileSync(habitsFile, 'utf8');
    return JSON.parse(data);
  } catch (e) {
    return [];
  }
}

function saveHabits(habits) {
  ensureDataDir();
  fs.writeFileSync(habitsFile, JSON.stringify(habits, null, 2), 'utf8');
}

function addHabit(habit) {
  const habits = getHabits();
  const newHabit = {
    id: Date.now().toString(),
    name: habit.name,
    icon: habit.icon || '📌',
    frequency: habit.frequency || 'daily',
    weeklyDays: habit.weeklyDays || [0, 1, 2, 3, 4, 5, 6],
    customInterval: habit.customInterval || 1,
    createdAt: new Date().toISOString()
  };
  habits.push(newHabit);
  saveHabits(habits);
  return { success: true, habit: newHabit };
}

function deleteHabit(habitId) {
  const habits = getHabits();
  const index = habits.findIndex(h => h.id === habitId);
  
  if (index === -1) {
    return { success: false, error: 'Habit not found' };
  }
  
  habits.splice(index, 1);
  saveHabits(habits);
  
  removeCheckinDataForHabit(habitId);
  
  return { success: true };
}

function removeCheckinDataForHabit(habitId) {
  const files = fs.readdirSync(dataDir).filter(f => f.startsWith('checkin-') && f.endsWith('.json'));
  files.forEach(file => {
    const filePath = path.join(dataDir, file);
    try {
      const data = JSON.parse(fs.readFileSync(filePath, 'utf8'));
      let modified = false;
      Object.keys(data).forEach(date => {
        if (data[date][habitId] !== undefined) {
          delete data[date][habitId];
          modified = true;
        }
      });
      if (modified) {
        fs.writeFileSync(filePath, JSON.stringify(data, null, 2), 'utf8');
      }
    } catch (e) {
      console.error('Error cleaning checkin data:', e);
    }
  });
}

function getCheckinData(year, month) {
  const filePath = getCheckinDataPath(year, month);
  if (!fs.existsSync(filePath)) {
    return {};
  }
  try {
    const data = fs.readFileSync(filePath, 'utf8');
    return JSON.parse(data);
  } catch (e) {
    return {};
  }
}

function saveCheckin(checkin) {
  const { date, habitId, completed } = checkin;
  const [year, month] = date.split('-');
  const filePath = getCheckinDataPath(parseInt(year), parseInt(month));
  ensureDataDir();
  
  let data = {};
  if (fs.existsSync(filePath)) {
    try {
      data = JSON.parse(fs.readFileSync(filePath, 'utf8'));
    } catch (e) {
      data = {};
    }
  }
  
  if (!data[date]) {
    data[date] = {};
  }
  data[date][habitId] = completed;
  
  fs.writeFileSync(filePath, JSON.stringify(data, null, 2), 'utf8');
  return { success: true };
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

function getStreaks(habitId) {
  const habits = getHabits();
  const habit = habits.find(h => h.id === habitId);
  
  if (!habit) {
    return { current: 0, max: 0 };
  }
  
  let currentStreak = 0;
  let maxStreak = 0;
  let tempStreak = 0;
  
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  
  for (let i = 0; i < 365; i++) {
    const checkDate = new Date(today);
    checkDate.setDate(today.getDate() - i);
    const dateStr = checkDate.toISOString().split('T')[0];
    
    if (shouldCheckinOnDate(habit, dateStr)) {
      const [year, month] = dateStr.split('-');
      const checkinData = getCheckinData(parseInt(year), parseInt(month));
      const completed = checkinData[dateStr] && checkinData[dateStr][habitId];
      
      if (completed) {
        tempStreak++;
        if (i === 0) {
          currentStreak = tempStreak;
        }
        if (tempStreak > maxStreak) {
          maxStreak = tempStreak;
        }
      } else {
        if (i === 0 || currentStreak > 0) {
          if (currentStreak === 0) {
            currentStreak = tempStreak;
          }
        }
        tempStreak = 0;
      }
    }
  }
  
  return {
    current: currentStreak,
    max: Math.max(maxStreak, currentStreak)
  };
}

function getMonthlyStats(year, month) {
  const habits = getHabits();
  const checkinData = getCheckinData(year, month);
  
  const daysInMonth = new Date(year, month, 0).getDate();
  
  let totalRequiredDays = 0;
  let completedDays = 0;
  
  const habitStats = habits.map(habit => {
    let habitTotalDays = 0;
    let habitCompletedDays = 0;
    
    for (let day = 1; day <= daysInMonth; day++) {
      const dateStr = `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
      
      if (shouldCheckinOnDate(habit, dateStr)) {
        habitTotalDays++;
        totalRequiredDays++;
        
        if (checkinData[dateStr] && checkinData[dateStr][habit.id]) {
          habitCompletedDays++;
          completedDays++;
        }
      }
    }
    
    return {
      id: habit.id,
      name: habit.name,
      icon: habit.icon,
      totalDays: habitTotalDays,
      completedDays: habitCompletedDays,
      rate: habitTotalDays > 0 ? (habitCompletedDays / habitTotalDays) * 100 : 0
    };
  });
  
  const monthlyRate = totalRequiredDays > 0 ? (completedDays / totalRequiredDays) * 100 : 0;
  
  return {
    monthlyRate,
    totalRequiredDays,
    completedDays,
    habitStats
  };
}

function exportToCSV(year, month) {
  const habits = getHabits();
  const checkinData = getCheckinData(year, month);
  const daysInMonth = new Date(year, month, 0).getDate();
  
  const fields = ['日期'];
  habits.forEach(h => fields.push(h.name));
  
  const data = [];
  for (let day = 1; day <= daysInMonth; day++) {
    const dateStr = `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
    const row = { 日期: dateStr };
    
    habits.forEach(habit => {
      if (shouldCheckinOnDate(habit, dateStr)) {
        const completed = checkinData[dateStr] && checkinData[dateStr][habit.id];
        row[habit.name] = completed ? '已完成' : '未完成';
      } else {
        row[habit.name] = '无需打卡';
      }
    });
    
    data.push(row);
  }
  
  try {
    const parser = new Parser({ fields });
    const csv = parser.parse(data);
    const fileName = `习惯打卡记录-${year}-${String(month).padStart(2, '0')}.csv`;
    const savePath = path.join(dataDir, fileName);
    
    fs.writeFileSync(savePath, '\uFEFF' + csv, 'utf8');
    return { success: true, path: savePath };
  } catch (e) {
    console.error('CSV export error:', e);
    return { success: false, error: e.message };
  }
}

module.exports = {
  getHabits,
  addHabit,
  deleteHabit,
  getCheckinData,
  saveCheckin,
  getStreaks,
  getMonthlyStats,
  exportToCSV,
  shouldCheckinOnDate
};
