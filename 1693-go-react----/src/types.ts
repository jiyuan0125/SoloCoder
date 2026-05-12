export type PovertyLevel = 'general' | 'low_income' | 'special_difficult';

export type PovertyCause = 
  | '因病' 
  | '因残' 
  | '因学' 
  | '因灾' 
  | '缺劳动力' 
  | '缺资金' 
  | '缺技术' 
  | '缺土地' 
  | '缺水' 
  | '交通不便' 
  | '自身发展动力不足' 
  | '其他';

export type MeasureType = 'industry' | 'employment' | 'education' | 'health' | 'support';

export type EvaluationResult = 'effective' | 'partially_effective' | 'ineffective';

export type HouseholdStatus = 'poverty' | 'out_of_poverty' | 'return_monitoring';

export interface FamilyMember {
  name: string;
  relation: string;
  age: number;
  idCard: string;
  healthStatus: string;
  educationLevel: string;
  employmentStatus: string;
}

export interface HouseholdAsset {
  house: string;
  land: string;
  livestock: string;
  vehicle: string;
  other: string;
}

export interface LaborForce {
  total: number;
  employed: number;
  skilled: number;
  disabled: number;
}

export interface TwoWorriesThreeGuarantees {
  worryAboutFood: boolean;
  worryAboutClothing: boolean;
  compulsoryEducation: boolean;
  basicMedicalCare: boolean;
  housingSafety: boolean;
}

export interface Household {
  id: string;
  householdId: string;
  county: string;
  township: string;
  village: string;
  headOfHousehold: string;
  contactPhone: string;
  familySize: number;
  familyMembers: FamilyMember[];
  povertyCauses: PovertyCause[];
  mainPovertyCause: PovertyCause;
  povertyLevel: PovertyLevel;
  assets: HouseholdAsset;
  laborForce: LaborForce;
  yearlyIncomePerCapita: number;
  twoWorriesThreeGuarantees: TwoWorriesThreeGuarantees;
  status: HouseholdStatus;
  outOfPovertyDate?: string;
  monitoringEndDate?: string;
  helperName: string;
  helperUnit: string;
  helperPhone: string;
  archiveYear: number;
  createdAt: string;
  updatedAt: string;
}

export interface AssistancePlan {
  id: string;
  householdId: string;
  measures: AssistanceMeasure[];
  adjustmentReason?: string;
  approvedByTownship?: boolean;
  approvalDate?: string;
  status: 'draft' | 'executing' | 'completed' | 'suspended' | 'revised';
  createdBy: string;
  createdAt: string;
  updatedAt: string;
}

export interface AssistanceMeasure {
  id: string;
  type: MeasureType;
  name: string;
  description: string;
  target: string;
  startDate: string;
  endDate: string;
  evaluations: EvaluationRecord[];
  needRedraft: boolean;
}

export interface EvaluationRecord {
  id: string;
  measureId: string;
  quarter: string;
  result: EvaluationResult;
  evaluator: string;
  evaluationDate: string;
  notes: string;
  consecutiveIneffectiveCount: number;
}

export interface HouseholdHistory {
  id: string;
  householdId: string;
  eventType: 'archive_update' | 'status_change' | 'out_of_poverty' | 'return_to_poverty';
  previousStatus?: HouseholdStatus;
  newStatus?: HouseholdStatus;
  description: string;
  operator: string;
  operationDate: string;
}

export interface PlanHistory {
  id: string;
  planId: string;
  householdId: string;
  eventType: 'adjustment' | 'approval' | 'restart' | 'suspend' | 'complete';
  previousStatus?: string;
  newStatus?: string;
  reason?: string;
  operator: string;
  operationDate: string;
}

export interface DashboardStats {
  level: 'county' | 'township' | 'village';
  name: string;
  totalHouseholds: number;
  outOfPovertyCount: number;
  returnMonitoringCount: number;
  povertyCauseDistribution: Record<PovertyCause, number>;
}
