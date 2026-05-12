import { Router, Request, Response } from 'express';
import db from '../database';
import { generateId, getNow } from '../utils';
import { Volunteer } from '../types';

const router = Router();

const PASSING_SCORE = 8;
const TOTAL_QUESTIONS = 10;

router.post('/entry-training', (req: Request, res: Response) => {
  const { volunteer_id, knowledge_score, safety_score } = req.body;

  const volunteer = db.prepare('SELECT * FROM volunteers WHERE id = ?').get(volunteer_id);
  if (!volunteer) {
    return res.status(404).json({ error: '志愿者不存在' });
  }

  if (knowledge_score < 0 || knowledge_score > TOTAL_QUESTIONS ||
      safety_score < 0 || safety_score > TOTAL_QUESTIONS) {
    return res.status(400).json({ error: '分数必须在 0 到 10 之间' });
  }

  const knowledgePassed = knowledge_score >= PASSING_SCORE;
  const safetyPassed = safety_score >= PASSING_SCORE;
  const passed = knowledgePassed && safetyPassed;

  const id = generateId();
  const now = getNow();

  db.prepare(`
    INSERT INTO training_records (id, volunteer_id, training_type, knowledge_score, safety_score, passed, completed_at)
    VALUES (?, ?, 'entry', ?, ?, ?, ?)
  `).run(id, volunteer_id, knowledge_score, safety_score, passed ? 1 : 0, now);

  if (passed) {
    db.prepare('UPDATE volunteers SET entry_training_completed = 1, entry_training_date = ?, last_annual_training_date = ? WHERE id = ?')
      .run(now, now, volunteer_id);
  }

  res.json({
    passed,
    knowledge: { score: knowledge_score, passed: knowledgePassed },
    safety: { score: safety_score, passed: safetyPassed },
    message: passed ? '入门培训通过' : `培训未通过，需要答对${PASSING_SCORE}道题`
  });
});

router.post('/annual-training', (req: Request, res: Response) => {
  const { volunteer_id, knowledge_score, safety_score } = req.body;

  const volunteer = db.prepare('SELECT * FROM volunteers WHERE id = ?').get(volunteer_id) as Volunteer | undefined;
  if (!volunteer) {
    return res.status(404).json({ error: '志愿者不存在' });
  }

  if (!volunteer.entry_training_completed) {
    return res.status(400).json({ error: '请先完成入门培训' });
  }

  if (knowledge_score < 0 || knowledge_score > TOTAL_QUESTIONS ||
      safety_score < 0 || safety_score > TOTAL_QUESTIONS) {
    return res.status(400).json({ error: '分数必须在 0 到 10 之间' });
  }

  const knowledgePassed = knowledge_score >= PASSING_SCORE;
  const safetyPassed = safety_score >= PASSING_SCORE;
  const passed = knowledgePassed && safetyPassed;

  const id = generateId();
  const now = getNow();

  db.prepare(`
    INSERT INTO training_records (id, volunteer_id, training_type, knowledge_score, safety_score, passed, completed_at)
    VALUES (?, ?, 'annual', ?, ?, ?, ?)
  `).run(id, volunteer_id, knowledge_score, safety_score, passed ? 1 : 0, now);

  if (passed) {
    db.prepare('UPDATE volunteers SET last_annual_training_date = ? WHERE id = ?')
      .run(now, volunteer_id);
  }

  res.json({
    passed,
    knowledge: { score: knowledge_score, passed: knowledgePassed },
    safety: { score: safety_score, passed: safetyPassed },
    message: passed ? '年度复训通过' : `培训未通过，需要答对${PASSING_SCORE}道题`
  });
});

router.get('/training-records/:volunteer_id', (req: Request, res: Response) => {
  const records = db.prepare('SELECT * FROM training_records WHERE volunteer_id = ? ORDER BY completed_at DESC')
    .all(req.params.volunteer_id);
  res.json(records);
});

export default router;
