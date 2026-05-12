import { Router, Request, Response } from 'express';
import db from '../db';
import { nowISO, calculateEvaluationGrade } from '../utils';

const router = Router();

interface ScoreRequest {
  organizationId: number;
  evaluationYear: number;
  score1: number;
  score2: number;
  score3: number;
  score4: number;
}

router.post('/submit', (req: Request, res: Response) => {
  const { organizationId, evaluationYear, score1, score2, score3, score4 } = req.body as ScoreRequest;

  if (!organizationId || !evaluationYear || score1 === undefined || score2 === undefined || score3 === undefined || score4 === undefined) {
    return res.status(400).json({ error: '缺少必填字段' });
  }

  const scores = [score1, score2, score3, score4];
  if (scores.some(s => s < 0 || s > 25)) {
    return res.status(400).json({ error: '各维度评分必须在 0-25 之间' });
  }

  const org = db.prepare('SELECT * FROM organizations WHERE id = ?').get(organizationId);
  if (!org) {
    return res.status(404).json({ error: '组织不存在' });
  }

  const orgData = org as any;
  if (orgData.status === '已注销') {
    return res.status(405).json({ error: '组织已注销，不可操作' });
  }
  if (orgData.status !== '已注册') {
    return res.status(400).json({ error: '只有已注册组织才能参加评估' });
  }

  const lastInspection = db.prepare(`
    SELECT result FROM annual_inspections 
    WHERE organization_id = ? 
    ORDER BY year DESC 
    LIMIT 1
  `).get(organizationId);

  if (!lastInspection || (lastInspection as any).result === '不合格') {
    return res.status(400).json({ error: '年检不合格，不能参加评估，请先通过整改补检' });
  }

  const existing = db.prepare('SELECT id FROM evaluations WHERE organization_id = ? AND evaluation_year = ?').get(organizationId, evaluationYear);
  if (existing) {
    return res.status(409).json({ error: '该年度评估已存在' });
  }

  const totalScore = score1 + score2 + score3 + score4;
  const grade = calculateEvaluationGrade(totalScore);
  const canAccept = grade === '5A' || grade === '4A' || grade === '3A';

  const now = nowISO();
  const insert = db.prepare(`
    INSERT INTO evaluations (organization_id, evaluation_year, score1, score2, score3, score4, total_score, grade, created_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
  `);
  const result = insert.run(organizationId, evaluationYear, score1, score2, score3, score4, totalScore, grade, now);

  res.status(201).json({
    id: result.lastInsertRowid,
    totalScore,
    grade,
    canAcceptGovernmentServices: canAccept,
    message: '评估已完成'
  });
});

router.get('/organization/:orgId', (req: Request, res: Response) => {
  const evals = db.prepare('SELECT * FROM evaluations WHERE organization_id = ? ORDER BY evaluation_year DESC').all(req.params.orgId);
  res.json(evals);
});

router.get('/:id', (req: Request, res: Response) => {
  const evalRecord = db.prepare('SELECT * FROM evaluations WHERE id = ?').get(req.params.id);
  if (!evalRecord) {
    return res.status(404).json({ error: '评估记录不存在' });
  }
  const data = evalRecord as any;
  res.json({
    ...data,
    canAcceptGovernmentServices: data.grade === '5A' || data.grade === '4A' || data.grade === '3A'
  });
});

export default router;
