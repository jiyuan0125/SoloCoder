export type Cents = number;

export interface OvertimeHours {
  weekday: number;
  weekend: number;
  holiday: number;
}

export interface SalaryCalculationInput {
  baseSalary: Cents;
  positionAllowance: Cents;
  performanceCoefficient: number;
  socialInsuranceRate: number;
  housingFundRate: number;
  taxThreshold: Cents;
  leaveDeduction: Cents;
  overtimeHours: OvertimeHours;
  yearEndBonus: Cents;
  employeeId: string;
  employeeName: string;
  month: string;
}

export interface SalaryBreakdown {
  baseSalaryAfterLeave: Cents;
  performanceBonus: Cents;
  overtimePay: Cents;
  overtimePayCapped: boolean;
  positionAllowance: Cents;
  socialInsuranceBase: Cents;
  socialInsuranceDeduction: Cents;
  housingFundDeduction: Cents;
  preTaxIncome: Cents;
  taxDeduction: Cents;
  yearEndBonusTax: Cents;
  yearEndBonusMergedTax: Cents;
  actualSalary: Cents;
  minimumWageSubsidy: Cents;
}

export interface YearEndBonusOption {
  method: 'separate' | 'merged';
  totalTax: Cents;
  actualSalary: Cents;
  yearEndBonusTax: Cents;
}

export interface SalaryCalculationResult {
  employeeId: string;
  employeeName: string;
  month: string;
  grossSalary: Cents;
  netSalary: Cents;
  breakdown: SalaryBreakdown;
  yearEndBonusOptions: YearEndBonusOption[];
  recommendedOption: 'separate' | 'merged';
  timestamp: number;
}

export interface SalaryRecord extends SalaryCalculationResult {
  id: string;
  createdAt: number;
}

export interface MonthComparison {
  current: Cents;
  previous: Cents | null;
  difference: Cents | null;
  percentage: number | null;
}

export interface YoYComparison {
  current: Cents;
  lastYearSameMonth: Cents | null;
  difference: Cents | null;
  percentage: number | null;
}

export interface SalaryComparisonResult {
  employeeId: string;
  employeeName: string;
  month: string;
  monthOverMonth: MonthComparison;
  yearOverYear: YoYComparison;
}
