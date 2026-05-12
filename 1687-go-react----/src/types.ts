export type AppointmentStatus = 'pending_confirmation' | 'confirmed' | 'in_progress' | 'completed' | 'cancelled' | 'no_show';

export interface Counselor {
  id: string;
  name: string;
  certificate: string;
  specialty: string;
  experience: number;
  fee: number;
  status: 'pending_review' | 'approved' | 'rejected';
  created_at: string;
}

export interface TimeSlot {
  id: string;
  counselor_id: string;
  is_recurring: boolean;
  day_of_week?: number;
  date?: string;
  start_time: string;
  end_time: string;
  is_booked: boolean;
  created_at: string;
}

export interface Client {
  id: string;
  name: string;
  phone: string;
  no_show_count: number;
  banned_until?: string;
  created_at: string;
}

export interface Appointment {
  id: string;
  counselor_id: string;
  client_id: string;
  client_name: string;
  client_phone: string;
  problem_description: string;
  time_slot_id: string;
  appointment_date: string;
  start_time: string;
  end_time: string;
  status: AppointmentStatus;
  created_at: string;
  updated_at: string;
}

export interface ConsultationRecord {
  id: string;
  appointment_id: string;
  counselor_id: string;
  client_id: string;
  consultation_date: string;
  duration: number;
  summary: string;
  follow_up_suggestions: string;
  created_at: string;
}
