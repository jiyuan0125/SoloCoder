const database = require('./database');

function getMonthlyStats(year, month) {
  const records = database.getRecordsByMonth(year, month);
  
  let totalIncome = 0;
  let totalExpense = 0;
  
  records.forEach(record => {
    if (record.type === 'income') {
      totalIncome += record.amount;
    } else {
      totalExpense += record.amount;
    }
  });
  
  return {
    totalIncome: totalIncome.toFixed(2),
    totalExpense: totalExpense.toFixed(2),
    balance: (totalIncome - totalExpense).toFixed(2),
    recordCount: records.length
  };
}

function getDailyExpenses(year, month) {
  const records = database.getRecordsByMonth(year, month);
  const dailyExpenses = {};
  
  records.forEach(record => {
    const dateStr = record.date.split('T')[0];
    if (!dailyExpenses[dateStr]) {
      dailyExpenses[dateStr] = {
        income: 0,
        expense: 0,
        total: 0
      };
    }
    
    if (record.type === 'income') {
      dailyExpenses[dateStr].income += record.amount;
    } else {
      dailyExpenses[dateStr].expense += record.amount;
    }
    dailyExpenses[dateStr].total += record.amount;
  });
  
  return dailyExpenses;
}

function getCategoryStats(year, month) {
  const records = database.getRecordsByMonth(year, month);
  const categoryStats = {};
  
  records.forEach(record => {
    const category = record.category;
    if (!categoryStats[category]) {
      categoryStats[category] = {
        category,
        type: record.type,
        amount: 0,
        count: 0
      };
    }
    
    categoryStats[category].amount += record.amount;
    categoryStats[category].count += 1;
  });
  
  const result = Object.values(categoryStats);
  result.sort((a, b) => b.amount - a.amount);
  
  return result;
}

function getBudgetStatus(year, month) {
  const budgets = database.getBudgets(year, month);
  const records = database.getRecordsByMonth(year, month);
  
  const categoryExpenses = {};
  
  records.forEach(record => {
    if (record.type === 'expense') {
      const category = record.category;
      if (!categoryExpenses[category]) {
        categoryExpenses[category] = 0;
      }
      categoryExpenses[category] += record.amount;
    }
  });
  
  const budgetStatus = [];
  
  for (const [category, budgetAmount] of Object.entries(budgets)) {
    const spent = categoryExpenses[category] || 0;
    const percentage = budgetAmount > 0 ? (spent / budgetAmount) * 100 : 0;
    
    let status = 'normal';
    if (percentage >= 100) {
      status = 'over';
    } else if (percentage >= 80) {
      status = 'warning';
    }
    
    budgetStatus.push({
      category,
      budget: budgetAmount,
      spent: parseFloat(spent.toFixed(2)),
      percentage: parseFloat(percentage.toFixed(1)),
      status,
      remaining: budgetAmount - spent
    });
  }
  
  return budgetStatus;
}

module.exports = {
  getMonthlyStats,
  getDailyExpenses,
  getCategoryStats,
  getBudgetStatus
};
