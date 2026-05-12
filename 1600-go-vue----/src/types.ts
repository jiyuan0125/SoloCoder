export enum TaskStatus {
  PENDING = '待排期',
  SCHEDULED = '已排期',
  OBSERVING = '观测中',
  COMPLETED = '已完成',
  ARCHIVED = '已归档',
}

export enum CancelReason {
  WEATHER = '天气原因',
  EQUIPMENT_FAULT = '设备故障',
}

export interface Equipment {
  id: number;
  name: string;
  description: string;
  created_at: string;
}

export interface ObservationTask {
  id: number;
  target_object: string;
  start_time: string | null;
  end_time: string | null;
  equipment_id: number | null;
  equipment_name: string | null;
  responsible_person: string;
  status: TaskStatus;
  scheduled_month: string;
  is_canceled: boolean;
  cancel_reason: CancelReason | null;
  fault_description: string | null;
  created_at: string;
  updated_at: string;
}

export interface EquipmentBooking {
  id: number;
  equipment_id: number;
  task_id: number;
  start_time: string;
  end_time: string;
  created_at: string;
}

export interface PublicEvent {
  id: number;
  title: string;
  start_time: string;
  end_time: string;
  max_capacity: number;
  current_bookings: number;
  is_special: boolean;
  is_canceled: boolean;
  created_at: string;
}

export interface Booking {
  id: number;
  event_id: number;
  phone: string;
  seats: number;
  is_waitlist: boolean;
  created_at: string;
}
