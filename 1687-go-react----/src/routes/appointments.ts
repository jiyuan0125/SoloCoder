import express from 'express';
import {
  createAppointment,
  getAppointmentById,
  updateAppointmentStatus,
  getAppointmentsByCounselor,
  getAppointmentsByClient
} from '../services/appointmentService';
import { getCounselorById } from '../services/counselorService';
import { getClientByPhone } from '../services/clientService';

const router = express.Router();

router.post('/', (req, res) => {
  const {
    counselorId,
    timeSlotId,
    clientName,
    clientPhone,
    problemDescription,
    appointmentDate,
    startTime,
    endTime
  } = req.body;

  if (!clientName || !clientPhone) {
    return res.status(400).json({ error: '姓名和手机号不能为空' });
  }

  const counselor = getCounselorById(counselorId);
  if (!counselor) {
    return res.status(404).json({ error: '咨询师不存在' });
  }

  const result = createAppointment({
    counselorId,
    timeSlotId,
    clientName,
    clientPhone,
    problemDescription,
    appointmentDate,
    startTime,
    endTime
  });

  if (!result.success) {
    if (result.message?.includes('30天') || result.message?.includes('限制')) {
      return res.status(403).json({ error: result.message });
    }
    if (result.message?.includes('已被预约') || result.message?.includes('同一天')) {
      return res.status(409).json({ error: result.message });
    }
    return res.status(400).json({ error: result.message });
  }

  res.status(201).json(result.appointment);
});

router.get('/:id', (req, res) => {
  const appointment = getAppointmentById(req.params.id);
  if (!appointment) {
    return res.status(404).json({ error: '预约不存在' });
  }
  res.json(appointment);
});

router.patch('/:id/status', (req, res) => {
  const { status, userId, userType } = req.body;
  const appointment = getAppointmentById(req.params.id);

  if (!appointment) {
    return res.status(404).json({ error: '预约不存在' });
  }

  if (userType === 'counselor' && appointment.counselor_id !== userId) {
    return res.status(403).json({ error: '无权操作此预约' });
  }

  if (userType === 'client') {
    const client = getClientByPhone(appointment.client_phone);
    if (!client || client.id !== userId) {
      return res.status(403).json({ error: '无权操作此预约' });
    }
  }

  const result = updateAppointmentStatus(
    req.params.id,
    status,
    userId,
    userType
  );

  if (!result.success) {
    if (result.message?.includes('并发') || result.message?.includes('状态已被')) {
      return res.status(409).json({ error: result.message });
    }
    return res.status(400).json({ error: result.message });
  }

  if (result.message) {
    return res.status(200).json({ message: result.message });
  }

  res.json({ message: '状态更新成功' });
});

router.get('/counselor/:counselorId', (req, res) => {
  const counselor = getCounselorById(req.params.counselorId);
  if (!counselor) {
    return res.status(404).json({ error: '咨询师不存在' });
  }

  const appointments = getAppointmentsByCounselor(req.params.counselorId);
  res.json(appointments);
});

router.get('/client/:clientId', (req, res) => {
  const appointments = getAppointmentsByClient(req.params.clientId);
  res.json(appointments);
});

export default router;
