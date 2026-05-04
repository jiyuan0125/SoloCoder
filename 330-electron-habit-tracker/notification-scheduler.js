const { Notification } = require('electron');
const path = require('path');
const dataStore = require('./data-store');

let checkInterval = null;
let lastNotificationWeek = null;

function getLastWeekStats() {
  const today = new Date();
  const lastWeekEnd = new Date(today);
  lastWeekEnd.setDate(today.getDate() - today.getDay());
  
  const lastWeekStart = new Date(lastWeekEnd);
  lastWeekStart.setDate(lastWeekEnd.getDate() - 6);
  
  const habits = dataStore.getHabits();
  const startYear = lastWeekStart.getFullYear();
  const startMonth = lastWeekStart.getMonth() + 1;
  const endYear = lastWeekEnd.getFullYear();
  const endMonth = lastWeekEnd.getMonth() + 1;
  
  let totalCheckins = 0;
  let completedCheckins = 0;
  
  const startDate = lastWeekStart.getDate();
  const endDate = lastWeekEnd.getDate();
  
  for (let d = startDate; d <= endDate; d++) {
    const dateStr = `${startYear}-${String(startMonth).padStart(2, '0')}-${String(d).padStart(2, '0')}`;
    
    habits.forEach(habit => {
      if (dataStore.shouldCheckinOnDate(habit, dateStr)) {
        totalCheckins++;
        
        const [year, month] = dateStr.split('-');
        const checkinData = dataStore.getCheckinData(parseInt(year), parseInt(month));
        
        if (checkinData[dateStr] && checkinData[dateStr][habit.id]) {
          completedCheckins++;
        }
      }
    });
  }
  
  const rate = totalCheckins > 0 ? Math.round((completedCheckins / totalCheckins) * 100) : 0;
  
  return {
    totalCheckins,
    completedCheckins,
    rate,
    startDate: lastWeekStart.toLocaleDateString('zh-CN'),
    endDate: lastWeekEnd.toLocaleDateString('zh-CN')
  };
}

function showWeeklyNotification() {
  const stats = getLastWeekStats();
  
  const title = '上周打卡总结';
  let body = `${stats.startDate} - ${stats.endDate}\n`;
  body += `打卡率: ${stats.rate}%\n`;
  body += `完成: ${stats.completedCheckins}/${stats.totalCheckins}`;
  
  if (stats.rate >= 80) {
    body += '\n🎉 太棒了！继续保持！';
  } else if (stats.rate >= 50) {
    body += '\n💪 还不错，再加把劲！';
  } else {
    body += '\n📝 下周要更加努力哦！';
  }
  
  new Notification({
    title: title,
    body: body,
    icon: path.join(__dirname, 'assets/icon.png')
  }).show();
}

function shouldSendNotification() {
  const now = new Date();
  const dayOfWeek = now.getDay();
  const hours = now.getHours();
  const currentWeek = Math.floor(now.getTime() / (7 * 24 * 60 * 60 * 1000));
  
  if (dayOfWeek === 1 && hours >= 8 && hours < 9) {
    if (lastNotificationWeek !== currentWeek) {
      lastNotificationWeek = currentWeek;
      return true;
    }
  }
  
  return false;
}

function start() {
  if (checkInterval) return;
  
  checkInterval = setInterval(() => {
    if (shouldSendNotification()) {
      showWeeklyNotification();
    }
  }, 60000);
  
  console.log('Notification scheduler started');
}

function stop() {
  if (checkInterval) {
    clearInterval(checkInterval);
    checkInterval = null;
    console.log('Notification scheduler stopped');
  }
}

module.exports = {
  start,
  stop,
  showWeeklyNotification,
  getLastWeekStats
};
