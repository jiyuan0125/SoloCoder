export type InvestigationResult = 'normal' | 'need_isolation' | 'need_hospital';
export type IsolationType = 'home' | 'centralized';
export type IsolationStatus = 'active' | 'pending_discharge' | 'discharged' | 'transferred_hospital';
export type TestResult = 'positive' | 'negative' | 'pending';

export interface Person {
  id: string;
  name: string;
  id_card: string;
  phone: string;
  address: string;
  is_key_person: boolean;
  created_at: string;
}

export interface Investigation {
  id: string;
  person_id: string;
  investigation_date: string;
  investigation_method: string;
  travel_history: string;
  health_status: string;
  result: InvestigationResult;
  created_at: string;
}

export interface IsolationRecord {
  id: string;
  person_id: string;
  isolation_type: IsolationType;
  start_date: string;
  end_date: string;
  status: IsolationStatus;
  created_at: string;
}

export interface HealthRecord {
  id: string;
  person_id: string;
  isolation_id?: string;
  record_date: string;
  temperature: number;
  symptoms: string;
  is_normal: boolean;
  created_at: string;
}

export interface TestRecord {
  id: string;
  person_id: string;
  isolation_id?: string;
  test_day: number;
  test_date: string;
  result: TestResult;
  created_at: string;
}

export interface DischargeNotification {
  id: string;
  isolation_id: string;
  person_id: string;
  notification_date: string;
  health_status: string;
  doctor_confirmed: boolean;
  doctor_name?: string;
  confirmed_at?: string;
  created_at: string;
}
