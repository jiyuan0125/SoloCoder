const fs = require('fs');
const path = require('path');

const DATA_DIR = path.join(__dirname, '../../data');
const CATEGORIES_FILE = path.join(DATA_DIR, 'categories.json');
const BUDGETS_DIR = path.join(DATA_DIR, 'budgets');
const RECORDS_DIR = path.join(DATA_DIR, 'records');

const DEFAULT_CATEGORIES = [
  { name: '餐饮', type: 'expense', isDefault: true },
  { name: '交通', type: 'expense', isDefault: true },
  { name: '购物', type: 'expense', isDefault: true },
  { name: '娱乐', type: 'expense', isDefault: true },
  { name: '居住', type: 'expense', isDefault: true },
  { name: '医疗', type: 'expense', isDefault: true },
  { name: '教育', type: 'expense', isDefault: true },
  { name: '工资', type: 'income', isDefault: true },
  { name: '兼职', type: 'income', isDefault: true }
];

function ensureDirExists(dir) {
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
}

function initialize() {
  ensureDirExists(DATA_DIR);
  ensureDirExists(BUDGETS_DIR);
  ensureDirExists(RECORDS_DIR);

  if (!fs.existsSync(CATEGORIES_FILE)) {
    fs.writeFileSync(CATEGORIES_FILE, JSON.stringify(DEFAULT_CATEGORIES, null, 2));
  }
}

function getRecordsFilePath(year, month) {
  const monthStr = month.toString().padStart(2, '0');
  return path.join(RECORDS_DIR, `records_${year}_${monthStr}.json`);
}

function readRecords(year, month) {
  const filePath = getRecordsFilePath(year, month);
  if (fs.existsSync(filePath)) {
    return JSON.parse(fs.readFileSync(filePath, 'utf8'));
  }
  return [];
}

function writeRecords(year, month, records) {
  const filePath = getRecordsFilePath(year, month);
  fs.writeFileSync(filePath, JSON.stringify(records, null, 2));
}

function generateId() {
  return Date.now().toString() + Math.random().toString(36).substr(2, 9);
}

function addRecord(record) {
  const date = new Date(record.date);
  const year = date.getFullYear();
  const month = date.getMonth() + 1;
  
  const records = readRecords(year, month);
  const newRecord = {
    id: generateId(),
    amount: parseFloat(record.amount),
    type: record.type,
    category: record.category,
    note: record.note || '',
    date: record.date,
    createdAt: new Date().toISOString()
  };
  
  records.push(newRecord);
  writeRecords(year, month, records);
  return newRecord;
}

function updateRecord(id, record) {
  const oldRecord = findRecordById(id);
  if (!oldRecord) return null;

  const oldDate = new Date(oldRecord.date);
  const oldYear = oldDate.getFullYear();
  const oldMonth = oldDate.getMonth() + 1;

  const newDate = new Date(record.date);
  const newYear = newDate.getFullYear();
  const newMonth = newDate.getMonth() + 1;

  let records = readRecords(oldYear, oldMonth);
  const index = records.findIndex(r => r.id === id);
  
  if (index === -1) return null;

  if (oldYear === newYear && oldMonth === newMonth) {
    records[index] = {
      ...records[index],
      ...record,
      amount: parseFloat(record.amount)
    };
    writeRecords(oldYear, oldMonth, records);
    return records[index];
  } else {
    const updatedRecord = {
      ...records[index],
      ...record,
      amount: parseFloat(record.amount)
    };
    
    records.splice(index, 1);
    writeRecords(oldYear, oldMonth, records);

    const newRecords = readRecords(newYear, newMonth);
    newRecords.push(updatedRecord);
    writeRecords(newYear, newMonth, newRecords);

    return updatedRecord;
  }
}

function deleteRecord(id) {
  const record = findRecordById(id);
  if (!record) return false;

  const date = new Date(record.date);
  const year = date.getFullYear();
  const month = date.getMonth() + 1;

  let records = readRecords(year, month);
  const index = records.findIndex(r => r.id === id);
  
  if (index === -1) return false;

  records.splice(index, 1);
  writeRecords(year, month, records);
  return true;
}

function findRecordById(id) {
  for (const yearDir of fs.readdirSync(RECORDS_DIR)) {
    const match = yearDir.match(/records_(\d{4})_(\d{2})\.json/);
    if (match) {
      const year = parseInt(match[1]);
      const month = parseInt(match[2]);
      const records = readRecords(year, month);
      const found = records.find(r => r.id === id);
      if (found) return found;
    }
  }
  return null;
}

function getRecordsByDate(date) {
  const d = new Date(date);
  const year = d.getFullYear();
  const month = d.getMonth() + 1;
  const dateStr = date.split('T')[0];

  const records = readRecords(year, month);
  return records.filter(r => r.date.split('T')[0] === dateStr);
}

function getRecordsByMonth(year, month) {
  return readRecords(year, month);
}

function searchRecords(filters) {
  let results = [];
  const { category, minAmount, maxAmount, keyword, startDate, endDate, type } = filters;

  const years = [];
  if (startDate && endDate) {
    const start = new Date(startDate);
    const end = new Date(endDate);
    for (let y = start.getFullYear(); y <= end.getFullYear(); y++) {
      years.push(y);
    }
  } else {
    years.push(new Date().getFullYear());
  }

  for (const year of years) {
    for (let month = 1; month <= 12; month++) {
      const records = readRecords(year, month);
      results = results.concat(records);
    }
  }

  results = results.filter(record => {
    if (category && record.category !== category) return false;
    if (type && record.type !== type) return false;
    
    const amount = record.amount;
    if (minAmount !== undefined && minAmount !== null && amount < parseFloat(minAmount)) return false;
    if (maxAmount !== undefined && maxAmount !== null && amount > parseFloat(maxAmount)) return false;
    
    if (keyword && !record.note.includes(keyword)) return false;
    
    const recordDate = new Date(record.date.split('T')[0]);
    if (startDate) {
      const start = new Date(startDate);
      if (recordDate < start) return false;
    }
    if (endDate) {
      const end = new Date(endDate);
      if (recordDate > end) return false;
    }
    
    return true;
  });

  return results;
}

function getCategories() {
  if (fs.existsSync(CATEGORIES_FILE)) {
    return JSON.parse(fs.readFileSync(CATEGORIES_FILE, 'utf8'));
  }
  return DEFAULT_CATEGORIES;
}

function addCategory(category) {
  const categories = getCategories();
  
  if (categories.find(c => c.name === category.name && c.type === category.type)) {
    return null;
  }

  categories.push({
    name: category.name,
    type: category.type,
    isDefault: false
  });

  fs.writeFileSync(CATEGORIES_FILE, JSON.stringify(categories, null, 2));
  return categories[categories.length - 1];
}

function deleteCategory(name) {
  const categories = getCategories();
  const index = categories.findIndex(c => c.name === name && !c.isDefault);
  
  if (index === -1) return false;

  categories.splice(index, 1);
  fs.writeFileSync(CATEGORIES_FILE, JSON.stringify(categories, null, 2));
  return true;
}

function getBudgetsFilePath(year, month) {
  const monthStr = month.toString().padStart(2, '0');
  return path.join(BUDGETS_DIR, `budget_${year}_${monthStr}.json`);
}

function getBudgets(year, month) {
  const filePath = getBudgetsFilePath(year, month);
  if (fs.existsSync(filePath)) {
    return JSON.parse(fs.readFileSync(filePath, 'utf8'));
  }
  return {};
}

function setBudget(year, month, category, amount) {
  const budgets = getBudgets(year, month);
  budgets[category] = parseFloat(amount);
  
  const filePath = getBudgetsFilePath(year, month);
  fs.writeFileSync(filePath, JSON.stringify(budgets, null, 2));
  return budgets;
}

module.exports = {
  initialize,
  addRecord,
  updateRecord,
  deleteRecord,
  getRecordsByDate,
  getRecordsByMonth,
  searchRecords,
  getCategories,
  addCategory,
  deleteCategory,
  getBudgets,
  setBudget
};
