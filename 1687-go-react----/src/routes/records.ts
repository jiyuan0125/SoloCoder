import express from 'express';
import {
  createConsultationRecord,
  getRecordByAppointment,
  getRecordsByCounselor
} from '../services/recordService';
import { getCounselorById } from '../services/counselorService';

const router = express.Router();

router.post('/', (req, res) => {
  const {
    appointmentId,
    counselorId,
    consultationDate,
    duration,
    summary,
    followUpSuggestions
  } = req.body;

  const counselor = getCounselorById(counselorId);
  if (!counselor) {
    return res.status(404).json({ error: '咨询师不存在' });
  }

  const result = createConsultationRecord(
    appointmentId,
    counselorId,
    {
      consultationDate,
      duration: Number(duration),
      summary,
      followUpSuggestions
    }
  );

  if (!result.success) {
    if (result.message?.includes('无权') || result.message?.includes('无权限')) {
      return res.status(403).json({ error: result.message });
    }
    return res.status(400).json({ error: result.message });
  }

  res.status(201).json(result.record);
});

router.get('/appointment/:appointmentId', (req, res) => {
  const { userId, userType } = req.query;

  const result = getRecordByAppointment(
    req.params.appointmentId,
    userId as string,
    userType as 'counselor' | 'client'
  );

  if (!result.success) {
    if (result.message?.includes('无权') || result.message?.includes('不可见')) {
      return res.status(403).json({ error: result.message });
    }
    if (result.message?.includes('不存在')) {
      return res.status(404).json({ error: result.message });
    }
    return res.status(400).json({ error: result.message });
  }

  res.json(result.record);
});

router.get('/counselor/:counselorId', (req, res) => {
  const counselor = getCounselorById(req.params.counselorId);
  if (!counselor) {
    return res.status(404).json({ error: '咨询师不存在' });
  }

  const records = getRecordsByCounselor(req.params.counselorId);
  res.json(records);
});

export default router;
