export type CareLevel = 'self-care' | 'semi-care' | 'full-care';

export const MONTHLY_FEES: Record<CareLevel, number> = {
  'self-care': 300000,
  'semi-care': 500000,
  'full-care': 800000
};

export const VALID_CARE_LEVELS: CareLevel[] = ['self-care', 'semi-care', 'full-care'];

export interface Elder {
  id: number;
  name: string;
  idCard: string;
  gender?: string;
  birthDate?: string;
  emergencyContact: string;
  emergencyContactPhone: string;
  medicalHistory?: string;
  careLevel: CareLevel;
  status: 'active' | 'discharged';
  checkInDate: string;
  checkOutDate?: string;
  roomNumber?: string;
  bedNumber?: string;
  depositAmount: number;
  createdAt: string;
}

export interface CheckInRequest {
  name: string;
  idCard: string;
  gender?: string;
  birthDate?: string;
  emergencyContact: string;
  emergencyContactPhone: string;
  medicalHistory?: string;
  careLevel: string;
  checkInDate?: string;
  roomNumber?: string;
  bedNumber?: string;
}

export interface CheckOutRequest {
  checkOutDate?: string;
}

export interface CarePlan {
  id: number;
  elderId: number;
  medicationReminder?: string;
  dietArrangement?: string;
  rehabilitationProject?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CarePlanChange {
  id: number;
  carePlanId: number;
  changeReason: string;
  approvedBy: string;
  changedAt: string;
  snapshot: string;
}

export interface CarePlanUpdateRequest {
  medicationReminder?: string;
  dietArrangement?: string;
  rehabilitationProject?: string;
  changeReason: string;
  approvedBy: string;
}

export interface CareRecord {
  id: number;
  elderId: number;
  recordDate: string;
  temperature: number;
  bloodPressure: string;
  diet?: string;
  specialNotes?: string;
  createdBy?: string;
  createdAt: string;
}

export interface CareRecordRequest {
  recordDate?: string;
  temperature: number | string;
  bloodPressure: string;
  diet?: string;
  specialNotes?: string;
  createdBy?: string;
}

export interface HealthWarning {
  id: number;
  elderId: number;
  warningType: string;
  description: string;
  snapshot: string;
  status: 'pending' | 'notified' | 'resolved';
  createdAt: string;
  notifiedAt?: string;
  notifiedTo?: string;
}

export interface DoctorNotification {
  id: number;
  warningId: number;
  doctorName: string;
  notifiedAt: string;
}

export interface FamilyMember {
  id: number;
  elderId: number;
  name: string;
  phone: string;
  relation: string;
  lastViewDate?: string;
  viewCount: number;
  createdAt: string;
}

export interface FamilyMessage {
  id: number;
  familyMemberId: number;
  elderId: number;
  content: string;
  createdAt: string;
}

export interface VisitAppointment {
  id: number;
  familyMemberId: number;
  elderId: number;
  visitDate: string;
  timeSlot: 'morning' | 'afternoon';
  visitorName: string;
  visitorPhone?: string;
  status: 'pending' | 'visited' | 'cancelled';
  createdAt: string;
}

export interface VisitAppointmentRequest {
  familyMemberId: number;
  elderId: number;
  visitDate: string;
  timeSlot: string;
  visitorName: string;
  visitorPhone?: string;
}

export interface BillingRecord {
  id: number;
  elderId: number;
  type: 'deposit' | 'charge' | 'refund' | 'deduction' | 'arrears';
  amount: number;
  description: string;
  relatedRecordId?: number;
  createdAt: string;
}

export interface OccupancyStats {
  totalBeds: number;
  occupiedBeds: number;
  occupancyRate: number;
}
