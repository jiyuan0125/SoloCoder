import express from 'express';
import {
  registerCounselor,
  getCounselorById,
  getAllCounselors,
  reviewCounselor,
  setTimeSlot,
  getAvailableTimeSlots
} from '../services/counselorService';

const router = express.Router();

router.post('/', (req, res) => {
  const { name, certificate, specialty, experience, fee } = req.body;
  
  if (!name || !certificate || !specialty || experience === undefined || fee === undefined) {
    return res.status(400).json({ error: '缺少必要字段' });
  }

  const counselor = registerCounselor({
    name,
    certificate,
    specialty,
    experience: Number(experience),
    fee: Number(fee)
  });

  res.status(201).json(counselor);
});

router.get('/', (req, res) => {
  const counselors = getAllCounselors();
  res.json(counselors);
});

router.get('/:id', (req, res) => {
  const counselor = getCounselorById(req.params.id);
  if (!counselor) {
    return res.status(404).json({ error: '咨询师不存在' });
  }
  res.json(counselor);
});

router.post('/:id/review', (req, res) => {
  const { approved } = req.body;
  const counselor = reviewCounselor(req.params.id, approved);
  
  if (!counselor) {
    return res.status(404).json({ error: '咨询师不存在' });
  }
  
  res.json(counselor);
});

router.post('/:id/time-slots', (req, res) => {
  const counselor = getCounselorById(req.params.id);
  if (!counselor) {
    return res.status(404).json({ error: '咨询师不存在' });
  }

  const { isRecurring, dayOfWeek, date, startTime, endTime } = req.body;

  if (isRecurring && dayOfWeek === undefined) {
    return res.status(400).json({ error: '重复时段需要指定星期几' });
  }

  if (!isRecurring && !date) {
    return res.status(400).json({ error: '非重复时段需要指定日期' });
  }

  if (!startTime || !endTime) {
    return res.status(400).json({ error: '需要指定开始和结束时间' });
  }

  setTimeSlot(req.params.id, {
    isRecurring: Boolean(isRecurring),
    dayOfWeek: dayOfWeek !== undefined ? Number(dayOfWeek) : undefined,
    date,
    startTime,
    endTime
  });

  res.status(201).json({ message: '时段设置成功' });
});

router.get('/:id/time-slots', (req, res) => {
  const { date } = req.query;
  const counselor = getCounselorById(req.params.id);
  
  if (!counselor) {
    return res.status(404).json({ error: '咨询师不存在' });
  }

  if (!date) {
    return res.status(400).json({ error: '需要指定日期参数' });
  }

  const slots = getAvailableTimeSlots(req.params.id, date as string);
  res.json(slots);
});

export default router;
