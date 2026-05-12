export type RoomType = 'standard' | 'deluxe' | 'suite';
export type ShiftType = 'morning' | 'afternoon' | 'night';
export type StaffRole = 'maternal_care' | 'pediatric_nurse';
export type FeedingType = 'breast' | 'formula' | 'mixed';
export type MotherStatus = 'booked' | 'checked_in' | 'checked_out';

export interface Mother {
  id: string;
  name: string;
  expected_due_date: string;
  room_type: RoomType;
  emergency_contact_name: string;
  emergency_contact_phone: string;
  check_in_date?: string;
  check_out_date?: string;
  booking_date: string;
  status: MotherStatus;
  total_deposit: number;
}

export interface Baby {
  id: string;
  mother_id: string;
  name: string;
  birth_weight: number;
  birth_length: number;
  feeding_type: FeedingType;
  birth_date: string;
}

export interface BabyLog {
  id: string;
  baby_id: string;
  log_date: string;
  feeding_time?: string;
  feeding_amount?: number;
  diaper_change?: boolean;
  temperature?: number;
  sleep_duration?: number;
  jaundice_index?: number;
  current_weight?: number;
  anomaly_notes?: string;
}

export interface Staff {
  id: string;
  name: string;
  role: StaffRole;
}

export interface Shift {
  id: string;
  shift_type: ShiftType;
  shift_date: string;
}

export interface ShiftAssignment {
  id: string;
  shift_id: string;
  staff_id: string;
}
