import { v4 as uuidv4 } from 'uuid';
import db from '../database';
import { ConsultationRecord } from '../types';
import { getAppointmentById } from './appointmentService';

export function createConsultationRecord(
  appointmentId: string,
  counselorId: string,
  data: {
    consultationDate: string;
    duration: number;
    summary: string;
    followUpSuggestions: string;
  }
): { success: boolean; message?: string; record?: ConsultationRecord } {
  const appointment = getAppointmentById(appointmentId);
  if (!appointment) {
    return { success: false, message: '预约不存在' };
  }

  if (appointment.counselor_id !== counselorId) {
    return { success: false, message: '无权限访问该预约' };
  }

  const existing = db.prepare(
    'SELECT * FROM consultation_records WHERE appointment_id = ?'
  ).get(appointmentId);

  if (existing) {
    return { success: false, message: '该预约已有咨询记录' };
  }

  const id = uuidv4();
  db.prepare(`
    INSERT INTO consultation_records (
      id, appointment_id, counselor_id, client_id,
      consultation_date, duration, summary, follow_up_suggestions
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
  `).run(
    id, appointmentId, counselorId, appointment.client_id,
    data.consultationDate, data.duration, data.summary, data.followUpSuggestions
  );

  return { success: true, record: getRecordById(id) as ConsultationRecord };
}

export function getRecordById(id: string): ConsultationRecord | undefined {
  return db.prepare('SELECT * FROM consultation_records WHERE id = ?').get(id) as ConsultationRecord | undefined;
}

export function getRecordsByCounselor(counselorId: string): ConsultationRecord[] {
  return db.prepare(`
    SELECT * FROM consultation_records 
    WHERE counselor_id = ? 
    ORDER BY consultation_date DESC
  `).all(counselorId) as ConsultationRecord[];
}

export function getRecordByAppointment(
  appointmentId: string,
  userId: string,
  userType: 'counselor' | 'client'
): { success: boolean; message?: string; record?: ConsultationRecord } {
  const appointment = getAppointmentById(appointmentId);
  if (!appointment) {
    return { success: false, message: '预约不存在' };
  }

  if (userType === 'client') {
    return { success: false, message: '咨询记录对来访者不可见' };
  }

  if (appointment.counselor_id !== userId) {
    return { success: false, message: '无权访问该来访者档案' };
  }

  const record = db.prepare(
    'SELECT * FROM consultation_records WHERE appointment_id = ?'
  ).get(appointmentId) as ConsultationRecord | undefined;

  if (!record) {
    return { success: false, message: '暂无咨询记录' };
  }

  return { success: true, record };
}
