import { v4 as uuidv4 } from 'uuid';
import db from '../database';
import { Appointment, AppointmentStatus } from '../types';
import { getCounselorById } from './counselorService';
import { getOrCreateClient, isClientBanned, incrementNoShow } from './clientService';
import { sendAppointmentConfirmation } from './notificationService';

const STATUS_FLOW: Record<AppointmentStatus, AppointmentStatus[]> = {
  pending_confirmation: ['confirmed', 'cancelled'],
  confirmed: ['in_progress', 'cancelled', 'no_show'],
  in_progress: ['completed'],
  completed: [],
  cancelled: [],
  no_show: []
};

function getHoursUntilAppointment(appointmentDate: string, startTime: string): number {
  const now = new Date();
  const appointmentDateTime = new Date(`${appointmentDate}T${startTime}`);
  const diffMs = appointmentDateTime.getTime() - now.getTime();
  return diffMs / (1000 * 60 * 60);
}

function getMinutesSinceStart(appointmentDate: string, startTime: string): number {
  const now = new Date();
  const appointmentDateTime = new Date(`${appointmentDate}T${startTime}`);
  const diffMs = now.getTime() - appointmentDateTime.getTime();
  return diffMs / (1000 * 60);
}

export function createAppointment(data: {
  counselorId: string;
  timeSlotId: string;
  clientName: string;
  clientPhone: string;
  problemDescription: string;
  appointmentDate: string;
  startTime: string;
  endTime: string;
}): { success: boolean; message?: string; appointment?: Appointment } {
  const counselor = getCounselorById(data.counselorId);
  if (!counselor) {
    return { success: false, message: '咨询师不存在' };
  }

  if (counselor.status !== 'approved') {
    return { success: false, message: '咨询师未审核通过' };
  }

  if (!data.clientName || !data.clientPhone) {
    return { success: false, message: '姓名和手机号不能为空' };
  }

  const client = getOrCreateClient(data.clientName, data.clientPhone);

  if (isClientBanned(client.id)) {
    return { success: false, message: '您已被限制预约，请30天后再试' };
  }

  const sameDayAppointment = db.prepare(`
    SELECT * FROM appointments 
    WHERE counselor_id = ? 
    AND client_id = ? 
    AND appointment_date = ?
    AND status NOT IN ('cancelled', 'no_show')
  `).get(data.counselorId, client.id, data.appointmentDate);

  if (sameDayAppointment) {
    return { success: false, message: '同一用户同一天对同一咨询师只能预约一次' };
  }

  const timeSlot = db.prepare('SELECT * FROM time_slots WHERE id = ?').get(data.timeSlotId) as any;
  if (!timeSlot) {
    return { success: false, message: '时段不存在' };
  }

  const existingBooking = db.prepare(`
    SELECT * FROM appointments 
    WHERE time_slot_id = ? 
    AND appointment_date = ?
    AND status NOT IN ('cancelled', 'no_show')
  `).get(data.timeSlotId, data.appointmentDate);

  if (existingBooking) {
    return { success: false, message: '该时段已被预约' };
  }

  if (timeSlot.is_booked === 1) {
    return { success: false, message: '该时段已被预约' };
  }

  const id = uuidv4();
  db.prepare(`
    INSERT INTO appointments (
      id, counselor_id, client_id, client_name, client_phone,
      problem_description, time_slot_id, appointment_date,
      start_time, end_time
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `).run(
    id, data.counselorId, client.id, data.clientName, data.clientPhone,
    data.problemDescription, data.timeSlotId, data.appointmentDate,
    data.startTime, data.endTime
  );

  return { success: true, appointment: getAppointmentById(id) as Appointment };
}

export function getAppointmentById(id: string): Appointment | undefined {
  const row = db.prepare('SELECT * FROM appointments WHERE id = ?').get(id) as any;
  if (!row) return undefined;
  return row as Appointment;
}

export function updateAppointmentStatus(
  appointmentId: string,
  newStatus: AppointmentStatus,
  userId: string,
  userType: 'counselor' | 'client'
): { success: boolean; message?: string } {
  const appointment = getAppointmentById(appointmentId);
  if (!appointment) {
    return { success: false, message: '预约不存在' };
  }

  const allowedNext = STATUS_FLOW[appointment.status];
  if (!allowedNext.includes(newStatus)) {
    return { success: false, message: '非法的状态流转' };
  }

  if (newStatus === 'cancelled') {
    const hoursUntil = getHoursUntilAppointment(appointment.appointment_date, appointment.start_time);
    const minHours = userType === 'counselor' ? 24 : 12;
    
    if (hoursUntil < minHours) {
      return { success: false, message: `不足${minHours}小时不能取消预约` };
    }
  }

  if (newStatus === 'no_show') {
    const minutesSince = getMinutesSinceStart(appointment.appointment_date, appointment.start_time);
    if (minutesSince > 15) {
      return { success: false, message: '已过15分钟不能再标爽约' };
    }
  }

  const existingAppointment = db.prepare(
    'SELECT * FROM appointments WHERE id = ?'
  ).get(appointmentId) as any;
  
  if (existingAppointment.status !== appointment.status) {
    return { success: false, message: '状态已被更新' };
  }

  db.prepare(`
    UPDATE appointments 
    SET status = ?, updated_at = CURRENT_TIMESTAMP
    WHERE id = ? AND status = ?
  `).run(newStatus, appointmentId, appointment.status);

  const result = db.prepare('SELECT changes() as changes').get() as any;
  if (result.changes === 0) {
    return { success: false, message: '并发更新失败' };
  }

  if (newStatus === 'confirmed') {
    const counselor = getCounselorById(appointment.counselor_id);
    let notificationSuccess = true;
    
    try {
      if (counselor) {
        notificationSuccess = sendAppointmentConfirmation(appointment.client_phone, {
          counselorName: counselor.name,
          date: appointment.appointment_date,
          startTime: appointment.start_time
        }) as unknown as boolean;
      }
    } catch {
      notificationSuccess = false;
    }

    return { 
      success: true, 
      message: notificationSuccess ? undefined : '预约成功但通知发送失败' 
    };
  }

  if (newStatus === 'no_show') {
    const newCount = incrementNoShow(appointment.client_id);
    if (newCount >= 3) {
      return { success: true, message: '已标记爽约，累计3次，30天内不能预约' };
    }
  }

  if (newStatus === 'cancelled') {
    db.prepare(
      'UPDATE time_slots SET is_booked = 0 WHERE id = ?'
    ).run(appointment.time_slot_id);
  }

  return { success: true };
}

export function getAppointmentsByCounselor(counselorId: string): Appointment[] {
  return db.prepare('SELECT * FROM appointments WHERE counselor_id = ? ORDER BY appointment_date DESC')
    .all(counselorId) as Appointment[];
}

export function getAppointmentsByClient(clientId: string): Appointment[] {
  return db.prepare('SELECT * FROM appointments WHERE client_id = ? ORDER BY appointment_date DESC')
    .all(clientId) as Appointment[];
}
