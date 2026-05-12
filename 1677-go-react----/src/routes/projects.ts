import express from 'express';
import { v4 as uuidv4 } from 'uuid';
import { db } from '../database';
import { Project, CreateProjectRequest } from '../types';
import dayjs from 'dayjs';

const router = express.Router();

router.post('/', (req, res) => {
  const body = req.body as CreateProjectRequest;
  
  if (!body.name || !body.description || body.target_amount === undefined || !body.start_date || !body.end_date) {
    return res.status(400).json({ error: '缺少必要字段' });
  }
  
  if (body.target_amount <= 0) {
    return res.status(400).json({ error: '目标金额必须大于0' });
  }
  
  const existing = db.prepare(
    'SELECT id FROM projects WHERE name = ?'
  ).get(body.name) as Project | undefined;
  
  if (existing) {
    return res.status(409).json({ error: '项目名称已存在' });
  }
  
  const id = uuidv4();
  const now = dayjs().format('YYYY-MM-DD HH:mm:ss');
  const hasQualification = body.has_tax_deductible_qualification ? 1 : 0;
  
  try {
    db.prepare(
      `INSERT INTO projects (id, name, description, target_amount, start_date, end_date, has_tax_deductible_qualification, created_at)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
    ).run(id, body.name, body.description, body.target_amount, body.start_date, body.end_date, hasQualification, now);
    
    const project = db.prepare('SELECT * FROM projects WHERE id = ?').get(id) as Project;
    return res.status(201).json(project);
  } catch (err) {
    return res.status(500).json({ error: '创建项目失败' });
  }
});

router.get('/:id', (req, res) => {
  const project = db.prepare('SELECT * FROM projects WHERE id = ?').get(req.params.id) as Project | undefined;
  
  if (!project) {
    return res.status(404).json({ error: '项目不存在' });
  }
  
  const donated = db.prepare(
    'SELECT COALESCE(SUM(amount), 0) as total FROM donations WHERE project_id = ?'
  ).get(req.params.id) as { total: number };
  
  return res.json({
    ...project,
    current_amount: donated.total
  });
});

router.get('/', (_req, res) => {
  const projects = db.prepare('SELECT * FROM projects ORDER BY created_at DESC').all() as Project[];
  const result = projects.map(project => {
    const donated = db.prepare(
      'SELECT COALESCE(SUM(amount), 0) as total FROM donations WHERE project_id = ?'
    ).get(project.id) as { total: number };
    return {
      ...project,
      current_amount: donated.total
    };
  });
  
  return res.json(result);
});

export default router;
