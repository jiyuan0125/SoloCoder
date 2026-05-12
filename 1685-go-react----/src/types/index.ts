export enum ServiceLevel {
  LEVEL_1 = 'LEVEL_1',
  LEVEL_2 = 'LEVEL_2',
  LEVEL_3 = 'LEVEL_3',
}

export enum CallType {
  EMERGENCY = 'EMERGENCY',
  DAILY = 'DAILY',
}

export enum CallStatus {
  PENDING = 'PENDING',
  RESPONDED = 'RESPONDED',
  NOTIFY_COMMUNITY = 'NOTIFY_COMMUNITY',
  NOTIFY_STREET = 'NOTIFY_STREET',
  CLOSED = 'CLOSED',
}

export enum VisitStatus {
  SCHEDULED = 'SCHEDULED',
  COMPLETED = 'COMPLETED',
  MISSED = 'MISSED',
  REPORTED = 'REPORTED',
}

export interface Elder {
  id: string;
  name: string;
  idCard: string;
  phone: string;
  address: string;
  emergencyContact: string;
  emergencyContactPhone: string;
  serviceLevel: ServiceLevel;
  transitionUntil: string | null;
  previousLevel: ServiceLevel | null;
  visitPlanNeedsManualAdjustment: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface Assessment {
  id: string;
  elderId: string;
  selfCareScore: number;
  cognitiveScore: number;
  totalScore: number;
  serviceLevel: ServiceLevel;
  assessmentDate: string;
  createdAt: string;
}

export interface CallRecord {
  id: string;
  elderId: string;
  type: CallType;
  status: CallStatus;
  calledAt: string;
  respondedAt: string | null;
  responseTime: number | null;
  staffId: string | null;
  createdAt: string;
}

export interface VisitPlan {
  id: string;
  elderId: string;
  scheduledDate: string;
  status: VisitStatus;
  completedAt: string | null;
  notes: string | null;
  createdAt: string;
}

export interface HealthRecord {
  id: string;
  elderId: string;
  systolic: number | null;
  diastolic: number | null;
  bloodSugar: number | null;
  recordDate: string;
  isAbnormal: boolean;
  medicalAdvice: string | null;
  createdAt: string;
}

export interface Notification {
  id: string;
  callRecordId: string;
  recipientType: 'COMMUNITY' | 'STREET';
  notifiedAt: string;
  message: string;
}

export interface CreateElderInput {
  name: string;
  idCard: string;
  phone: string;
  address: string;
  emergencyContact: string;
  emergencyContactPhone: string;
}

export interface CreateAssessmentInput {
  elderId: string;
  selfCareScore: number;
  cognitiveScore: number;
  assessmentDate?: string;
}

export interface CreateCallInput {
  elderId: string;
  type: CallType;
}

export interface RespondToCallInput {
  callId: string;
  staffId: string;
}

export interface CreateHealthRecordInput {
  elderId: string;
  systolic: number | null;
  diastolic: number | null;
  bloodSugar: number | null;
  recordDate?: string;
}

export interface CompleteVisitInput {
  planId: string;
  notes?: string;
  visited: boolean;
  contactedByPhone: boolean;
}
