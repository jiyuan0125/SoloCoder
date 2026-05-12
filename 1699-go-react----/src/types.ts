export interface VaccineBatch {
  id: number;
  name: string;
  manufacturer: string;
  batchNumber: string;
  specification: string;
  expiryDate: string;
  stock: number;
  storageTemperature: string;
  safeStock: number;
  status: 'normal' | 'near_expiry' | 'expired' | 'suspended';
}

export interface Reservation {
  id: number;
  residentId: string;
  residentName: string;
  vaccineName: string;
  date: string;
  timeSlot: string;
  status: 'active' | 'cancelled' | 'completed';
  isBooster: boolean;
  mergedFrom?: string;
}

export interface VaccinationRecord {
  id: number;
  certificateNumber: string;
  reservationId?: number;
  residentId: string;
  residentName: string;
  vaccineName: string;
  batchNumber: string;
  injectionSite: string;
  dose: number;
  vaccinationDate: string;
  nextVaccinationDate?: string;
  isOverdue: boolean;
}

export interface AdverseReaction {
  id: number;
  recordId: number;
  reactionId: string;
  symptoms: string;
  severity: 'mild' | 'moderate' | 'severe';
  reportDate: string;
  isHandled: boolean;
}

export interface Notification {
  id: number;
  type: 'stock' | 'appointment_merge' | 'overdue' | 'emergency';
  message: string;
  targetId: string;
  createdAt: string;
  isRead: boolean;
}
