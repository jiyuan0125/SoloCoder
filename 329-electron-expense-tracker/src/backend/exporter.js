const fs = require('fs');

function exportToCSV(records, filePath) {
  try {
    if (records.length === 0) {
      return { success: false, message: '没有记录需要导出' };
    }

    const headers = ['日期', '类型', '分类', '金额', '备注', '创建时间'];
    
    let csvContent = headers.join(',') + '\n';
    
    records.forEach(record => {
      const type = record.type === 'income' ? '收入' : '支出';
      const amount = record.amount.toFixed(2);
      const note = record.note ? escapeCSV(record.note) : '';
      const date = record.date.split('T')[0];
      const createdAt = record.createdAt ? record.createdAt.replace('T', ' ').substring(0, 19) : '';
      
      const row = [
        date,
        type,
        record.category,
        amount,
        note,
        createdAt
      ].join(',');
      
      csvContent += row + '\n';
    });

    fs.writeFileSync(filePath, '\ufeff' + csvContent, 'utf8');
    
    return {
      success: true,
      message: `成功导出 ${records.length} 条记录`,
      filePath: filePath
    };
  } catch (error) {
    return {
      success: false,
      message: `导出失败: ${error.message}`
    };
  }
}

function escapeCSV(value) {
  if (typeof value !== 'string') {
    value = String(value);
  }
  
  if (value.includes(',') || value.includes('"') || value.includes('\n') || value.includes('\r')) {
    value = '"' + value.replace(/"/g, '""') + '"';
  }
  
  return value;
}

module.exports = {
  exportToCSV
};
