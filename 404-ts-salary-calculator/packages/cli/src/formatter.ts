import {
  SalaryCalculationResult,
  SalaryRecord,
  SalaryComparisonResult,
  YearEndBonusOption,
} from '@salary/shared';

function formatCents(cents: number): string {
  const yuan = (cents / 100).toFixed(2);
  return `${yuan} 元`;
}

function formatPercentage(value: number | null): string {
  if (value === null) return 'N/A';
  return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`;
}

export function formatSalaryResult(result: SalaryCalculationResult): string {
  const { breakdown, yearEndBonusOptions, recommendedOption } = result;
  
  let output = `
========================================
        薪资计算结果
========================================
员工ID: ${result.employeeId}
员工姓名: ${result.employeeName}
月份: ${result.month}

────────────────────────────────────────
【应发工资明细】
────────────────────────────────────────
基本工资（扣请假后）: ${formatCents(breakdown.baseSalaryAfterLeave)}
岗位津贴: ${formatCents(breakdown.positionAllowance)}
绩效奖金: ${formatCents(breakdown.performanceBonus)}
加班费: ${formatCents(breakdown.overtimePay)}${breakdown.overtimePayCapped ? ' (已封顶)' : ''}
────────────────────────────────────────
社保缴纳基数: ${formatCents(breakdown.socialInsuranceBase)}
社保扣除: ${formatCents(breakdown.socialInsuranceDeduction)}
公积金扣除: ${formatCents(breakdown.housingFundDeduction)}
────────────────────────────────────────
税前收入: ${formatCents(breakdown.preTaxIncome)}
个税扣除: ${formatCents(breakdown.taxDeduction)}
`;
  
  if (breakdown.minimumWageSubsidy > 0) {
    output += `最低工资补贴: ${formatCents(breakdown.minimumWageSubsidy)}\n`;
  }
  
  output += `
────────────────────────────────────────
【年终奖计税方案对比】
────────────────────────────────────────
`;
  
  for (const option of yearEndBonusOptions) {
    const isRecommended = option.method === recommendedOption ? ' ⭐ 推荐' : '';
    output += `
方案: ${option.method === 'separate' ? '单独计税' : '合并计税'}${isRecommended}
  年终奖税额: ${formatCents(option.yearEndBonusTax)}
  总税额: ${formatCents(option.totalTax)}
  实发工资: ${formatCents(option.actualSalary)}
`;
  }
  
  output += `
========================================
应发工资总额: ${formatCents(result.grossSalary)}
实发工资: ${formatCents(result.netSalary)}
========================================
`;
  
  return output;
}

export function formatRecords(records: SalaryRecord[]): string {
  if (records.length === 0) {
    return '未找到任何薪资记录。';
  }
  
  let output = `
========================================
        薪资记录列表 (共 ${records.length} 条)
========================================
`;
  
  for (const record of records) {
    output += `
【记录ID】 ${record.id}
员工: ${record.employeeName} (${record.employeeId})
月份: ${record.month}
应发: ${formatCents(record.grossSalary)}
实发: ${formatCents(record.netSalary)}
创建时间: ${new Date(record.createdAt).toLocaleString()}
────────────────────────────────────────
`;
  }
  
  return output;
}

export function formatComparison(comparison: SalaryComparisonResult): string {
  const { monthOverMonth, yearOverYear } = comparison;
  
  let output = `
========================================
        薪资对比分析
========================================
员工ID: ${comparison.employeeId}
员工姓名: ${comparison.employeeName}
月份: ${comparison.month}

────────────────────────────────────────
【环比对比（vs 上月）】
────────────────────────────────────────
本月实发: ${formatCents(monthOverMonth.current)}
上月实发: ${monthOverMonth.previous !== null ? formatCents(monthOverMonth.previous) : 'N/A'}
差额: ${monthOverMonth.difference !== null ? formatCents(monthOverMonth.difference) : 'N/A'}
变化率: ${formatPercentage(monthOverMonth.percentage)}

────────────────────────────────────────
【同比对比（vs 去年同期）】
────────────────────────────────────────
本月实发: ${formatCents(yearOverYear.current)}
去年同期: ${yearOverYear.lastYearSameMonth !== null ? formatCents(yearOverYear.lastYearSameMonth) : 'N/A'}
差额: ${yearOverYear.difference !== null ? formatCents(yearOverYear.difference) : 'N/A'}
变化率: ${formatPercentage(yearOverYear.percentage)}
========================================
`;
  
  return output;
}
