export type ChronicDiseaseType = 'hypertension' | 'diabetes' | 'hyperlipidemia' | 'chronic_kidney_disease' | 'cardiovascular_disease';

export type RiskLevel = 'low' | 'medium' | 'high';

export interface BasicInfo {
  name: string;
  gender: 'male' | 'female' | 'other';
  age: number;
  phone: string;
  idCard: string;
  address?: string;
}

export interface InitialMetrics {
  systolicBP: number;
  diastolicBP: number;
  fastingBloodSugar: number;
  HbA1c?: number;
  weight: number;
  height: number;
}

export interface Patient {
  id: string;
  basicInfo: BasicInfo;
  chronicDiseases: ChronicDiseaseType[];
  diagnosisDate: Date;
  medicalHistory: string[];
  initialMetrics: InitialMetrics;
  riskLevel: RiskLevel;
  followUpIntervalDays: number;
  nextFollowUpDate?: Date;
  lastFollowUpDate?: Date;
  createdAt: Date;
  updatedAt: Date;
  medicationAdherence: 'good' | 'poor';
  nonCompliantCount: number;
}

export interface FollowUpMetrics {
  systolicBP: number;
  diastolicBP: number;
  fastingBloodSugar: number;
  HbA1c?: number;
  weight?: number;
  otherMetrics?: Record<string, number>;
}

export interface Medication {
  medicationId: string;
  name: string;
  genericName: string;
  dosage: string;
  frequency: string;
  startDate: Date;
  endDate?: Date;
  remainingDays?: number;
  prescribedBy: string;
}

export interface LifestyleAssessment {
  smoking: 'none' | 'light' | 'heavy';
  alcohol: 'none' | 'light' | 'moderate' | 'heavy';
  exercise: 'none' | 'occasional' | 'regular';
  diet: 'unhealthy' | 'moderate' | 'healthy';
  sleepHours: number;
}

export interface FollowUpRecord {
  id: string;
  patientId: string;
  date: Date;
  symptoms: string[];
  medications: Medication[];
  metrics: FollowUpMetrics;
  lifestyleAssessment: LifestyleAssessment;
  medicationTaken: boolean;
  notes?: string;
  createdAt: Date;
}

export interface Contraindication {
  drugA: string;
  drugB: string;
  description: string;
  severity: 'mild' | 'moderate' | 'severe';
}

export interface RefillReminder {
  id: string;
  patientId: string;
  medicationId: string;
  medicationName: string;
  expirationDate: Date;
  reminderDate: Date;
  sent: boolean;
  sentAt?: Date;
  createdAt: Date;
}

export interface ControlRate {
  hypertension: number;
  diabetes: number;
  overall: number;
  totalPatients: number;
  controlledPatients: number;
  period: 'month' | 'quarter' | 'year';
  calculatedAt: Date;
}

export interface FollowUpCompletionRate {
  period: 'month' | 'quarter' | 'year';
  totalExpected: number;
  completed: number;
  rate: number;
  calculatedAt: Date;
}

export interface BudgetItem {
  id: string;
  name: string;
  amount: number;
  settledAmount: number;
  category: 'followup' | 'medication' | 'lab' | 'other';
  status: 'pending' | 'settled';
}

export interface Budget {
  id: string;
  period: string;
  totalAmount: number;
  originalTotal: number;
  items: BudgetItem[];
  createdAt: Date;
  updatedAt: Date;
}

export interface EntityConfig {
  id: string;
  type: 'chronic_disease' | 'medication' | 'followup_template';
  code: string;
  name: string;
  config: Record<string, unknown>;
  active: boolean;
  createdAt: Date;
}
