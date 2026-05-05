import {
  Cents,
  SalaryCalculationInput,
  SalaryCalculationResult,
  SalaryBreakdown,
  YearEndBonusOption,
} from '@salary/shared';
import {
  SOCIAL_INSURANCE_BASE_MIN,
  SOCIAL_INSURANCE_BASE_MAX,
  MINIMUM_WAGE,
  OVERTIME_CAP_RATIO,
  TAX_BRACKETS,
  OVERTIME_MULTIPLIERS,
} from '@salary/shared';
import { OvertimeHours } from '@salary/shared';

function roundToCents(value: number): Cents {
  return Math.round(value);
}

function max(a: Cents, b: Cents): Cents {
  return a > b ? a : b;
}

function min(a: Cents, b: Cents): Cents {
  return a < b ? a : b;
}

export function calculateBaseSalaryAfterLeave(baseSalary: Cents, leaveDeduction: Cents): Cents {
  const result = baseSalary - leaveDeduction;
  return max(result, 0);
}

export function calculateSocialInsuranceBase(baseSalaryAfterLeave: Cents): Cents {
  if (baseSalaryAfterLeave < SOCIAL_INSURANCE_BASE_MIN) {
    return SOCIAL_INSURANCE_BASE_MIN;
  }
  if (baseSalaryAfterLeave > SOCIAL_INSURANCE_BASE_MAX) {
    return SOCIAL_INSURANCE_BASE_MAX;
  }
  return baseSalaryAfterLeave;
}

export function calculateOvertimePay(
  baseSalary: Cents,
  overtimeHours: OvertimeHours
): { overtimePay: Cents; isCapped: boolean } {
  const hourlyRate = baseSalary / 21.75 / 8;
  
  const weekdayPay = overtimeHours.weekday * hourlyRate * OVERTIME_MULTIPLIERS.weekday;
  const weekendPay = overtimeHours.weekend * hourlyRate * OVERTIME_MULTIPLIERS.weekend;
  const holidayPay = overtimeHours.holiday * hourlyRate * OVERTIME_MULTIPLIERS.holiday;
  
  const totalOvertimePay = roundToCents(weekdayPay + weekendPay + holidayPay);
  const cap = roundToCents(baseSalary * OVERTIME_CAP_RATIO);
  
  if (totalOvertimePay <= cap) {
    return { overtimePay: totalOvertimePay, isCapped: false };
  }
  return { overtimePay: cap, isCapped: true };
}

export function calculatePerformanceBonus(baseSalary: Cents, performanceCoefficient: number): Cents {
  return roundToCents(baseSalary * performanceCoefficient);
}

export function calculateProgressiveTax(taxableIncome: Cents): Cents {
  if (taxableIncome <= 0) {
    return 0;
  }
  
  let tax = 0;
  let previousThreshold = 0;
  
  for (const bracket of TAX_BRACKETS) {
    if (taxableIncome <= previousThreshold) {
      break;
    }
    
    const bracketIncome = min(taxableIncome, bracket.threshold) - previousThreshold;
    
    if (bracketIncome > 0) {
      tax += bracketIncome * bracket.rate;
    }
    
    previousThreshold = bracket.threshold;
  }
  
  return roundToCents(tax);
}

export function calculateSeparateYearEndBonusTax(bonus: Cents): Cents {
  if (bonus <= 0) {
    return 0;
  }
  
  const monthlyBonus = bonus / 12;
  const monthlyTax = calculateProgressiveTax(roundToCents(monthlyBonus));
  return roundToCents(monthlyTax * 12);
}

export function calculateSalary(
  input: SalaryCalculationInput
): SalaryCalculationResult {
  const baseSalaryAfterLeave = calculateBaseSalaryAfterLeave(input.baseSalary, input.leaveDeduction);
  const socialInsuranceBase = calculateSocialInsuranceBase(baseSalaryAfterLeave);
  
  const socialInsuranceDeduction = roundToCents(socialInsuranceBase * input.socialInsuranceRate);
  const housingFundDeduction = roundToCents(socialInsuranceBase * input.housingFundRate);
  
  const performanceBonus = calculatePerformanceBonus(input.baseSalary, input.performanceCoefficient);
  const { overtimePay, isCapped } = calculateOvertimePay(input.baseSalary, input.overtimeHours);
  
  const preTaxIncome =
    baseSalaryAfterLeave +
    input.positionAllowance +
    performanceBonus +
    overtimePay -
    socialInsuranceDeduction -
    housingFundDeduction;
  
  const taxableIncome = max(preTaxIncome - input.taxThreshold, 0);
  const taxDeduction = calculateProgressiveTax(taxableIncome);
  
  const separateBonusTax = calculateSeparateYearEndBonusTax(input.yearEndBonus);
  
  const taxableIncomeWithBonus = max(
    preTaxIncome + input.yearEndBonus - input.taxThreshold,
    0
  );
  const taxWithBonus = calculateProgressiveTax(taxableIncomeWithBonus);
  const mergedBonusTax = taxWithBonus - taxDeduction;
  
  const actualSalarySeparate =
    preTaxIncome + input.yearEndBonus - taxDeduction - separateBonusTax;
  const actualSalaryMerged =
    preTaxIncome + input.yearEndBonus - taxWithBonus;
  
  let minimumWageSubsidy = 0;
  let finalNetSalary = min(actualSalarySeparate, actualSalaryMerged);
  
  if (finalNetSalary < MINIMUM_WAGE) {
    minimumWageSubsidy = MINIMUM_WAGE - finalNetSalary;
    finalNetSalary = MINIMUM_WAGE;
  }
  
  const recommendedOption: 'separate' | 'merged' = actualSalarySeparate >= actualSalaryMerged ? 'separate' : 'merged';
  
  const grossSalary =
    baseSalaryAfterLeave +
    input.positionAllowance +
    performanceBonus +
    overtimePay +
    input.yearEndBonus;
  
  const breakdown: SalaryBreakdown = {
    baseSalaryAfterLeave,
    performanceBonus,
    overtimePay,
    overtimePayCapped: isCapped,
    positionAllowance: input.positionAllowance,
    socialInsuranceBase,
    socialInsuranceDeduction,
    housingFundDeduction,
    preTaxIncome,
    taxDeduction,
    yearEndBonusTax: separateBonusTax,
    yearEndBonusMergedTax: mergedBonusTax,
    actualSalary: finalNetSalary,
    minimumWageSubsidy,
  };
  
  const separateOption: YearEndBonusOption = {
    method: 'separate',
    totalTax: taxDeduction + separateBonusTax,
    actualSalary: actualSalarySeparate + (actualSalarySeparate < MINIMUM_WAGE ? (MINIMUM_WAGE - actualSalarySeparate) : 0),
    yearEndBonusTax: separateBonusTax,
  };
  
  const mergedOption: YearEndBonusOption = {
    method: 'merged',
    totalTax: taxWithBonus,
    actualSalary: actualSalaryMerged + (actualSalaryMerged < MINIMUM_WAGE ? (MINIMUM_WAGE - actualSalaryMerged) : 0),
    yearEndBonusTax: mergedBonusTax,
  };
  
  return {
    employeeId: input.employeeId,
    employeeName: input.employeeName,
    month: input.month,
    grossSalary,
    netSalary: finalNetSalary,
    breakdown,
    yearEndBonusOptions: [separateOption, mergedOption],
    recommendedOption,
    timestamp: Date.now(),
  };
}

export function validateInput(input: SalaryCalculationInput): string | null {
  if (input.baseSalary < 0) {
    return '基本工资不能为负数';
  }
  if (input.positionAllowance < 0) {
    return '岗位津贴不能为负数';
  }
  if (input.performanceCoefficient < 0 || input.performanceCoefficient > 2) {
    return '绩效系数必须在 0 到 2 之间';
  }
  if (input.socialInsuranceRate < 0 || input.socialInsuranceRate > 1) {
    return '社保比例必须在 0 到 1 之间';
  }
  if (input.housingFundRate < 0 || input.housingFundRate > 1) {
    return '公积金比例必须在 0 到 1 之间';
  }
  if (input.taxThreshold < 0) {
    return '个税起征点不能为负数';
  }
  if (input.leaveDeduction < 0) {
    return '请假扣款不能为负数';
  }
  if (input.leaveDeduction > input.baseSalary) {
    return '请假扣款不能超过基本工资';
  }
  if (input.overtimeHours.weekday < 0 ||
      input.overtimeHours.weekend < 0 ||
      input.overtimeHours.holiday < 0) {
    return '加班小时数不能为负数';
  }
  if (input.yearEndBonus < 0) {
    return '年终奖不能为负数';
  }
  if (!input.employeeId || input.employeeId.trim() === '') {
    return '员工ID不能为空';
  }
  if (!input.month || !/^\d{4}-\d{2}$/.test(input.month)) {
    return '月份格式无效，应为 YYYY-MM';
  }
  
  return null;
}
