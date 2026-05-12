export interface Activity {
  id: string;
  name: string;
  category: string;
  date_time: string;
  location: string;
  max_volunteers: number;
  registration_deadline: string;
  created_at: string;
  updated_at: string;
}

export interface Volunteer {
  id: string;
  name: string;
  phone: string;
  id_card: string;
  age: number;
  guardian_name?: string;
  guardian_phone?: string;
  emergency_contact: string;
  entry_training_completed: number;
  entry_training_date?: string;
  last_annual_training_date?: string;
  created_at: string;
}

export interface Registration {
  id: string;
  activity_id: string;
  volunteer_id: string;
  status: string;
  is_special_position: number;
  hours: number;
  registered_at: string;
  reviewed_at?: string;
}

export interface TrainingRecord {
  id: string;
  volunteer_id: string;
  training_type: string;
  knowledge_score: number;
  safety_score: number;
  passed: number;
  completed_at: string;
}

export interface Point {
  id: string;
  volunteer_id: string;
  amount: number;
  description: string;
  source: string;
  created_at: string;
}

export interface Reminder {
  id: string;
  volunteer_id: string;
  type: string;
  message: string;
  due_date: string;
  is_read: number;
  created_at: string;
}

export interface Notification {
  id: string;
  volunteer_id: string;
  activity_id?: string;
  content: string;
  sent_at?: string;
  status: string;
}
