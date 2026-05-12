import { Router, Request, Response } from 'express';
import db from '../db';

const router = Router();

const REPAIR_CATEGORIES = ['水电', '门窗', '电梯', '公共设施', '其他'];
const EMERGENCY_CATEGORIES = ['水电', '电梯'];
const STATUS_TRANSITIONS: Record<string, string[]> = {
  '待接单': ['已分配'],
  '已分配': ['处理中'],
  '处理中': ['已完成'],
  '已完成': []
};

interface Repair {
  id: number;
  category: string;
  location: string;
  description: string;
  phone: string;
  status: string;
  priority: string;
  assigned_staff_id: number | null;
  assigned_at: number | null;
  completed_at: number | null;
  evaluation_score: number | null;
  evaluation_comment: string | null;
  created_at: number;
}

interface Staff {
  id: number;
  name: string;
  phone: string;
  created_at: number;
}

router.post('/', (req: Request, res: Response) => {
  const { category, location, description, phone } = req.body;

  if (!category || !REPAIR_CATEGORIES.includes(category)) {
    return res.status(400).json({ error: '报修类别无效，有效类别：水电、门窗、电梯、公共设施、其他' });
  }

  if (!location || location.trim() === '') {
    return res.status(400).json({ error: '位置不能为空' });
  }

  if (!description || description.trim() === '') {
    return res.status(400).json({ error: '问题描述不能为空' });
  }

  if (!phone || phone.trim() === '') {
    return res.status(400).json({ error: '联系电话不能为空' });
  }

  const priority = EMERGENCY_CATEGORIES.includes(category) ? '紧急' : '普通';
  const now = Date.now();

  const result = db.prepare(`
    INSERT INTO repairs (category, location, description, phone, status, priority, created_at)
    VALUES (?, ?, ?, ?, '待接单', ?, ?)
  `).run(category, location.trim(), description.trim(), phone.trim(), priority, now);

  const repair = db.prepare('SELECT * FROM repairs WHERE id = ?').get(result.lastInsertRowid) as Repair;
  res.status(201).json(formatRepair(repair));
});

router.get('/', (req: Request, res: Response) => {
  const rows = db.prepare('SELECT * FROM repairs ORDER BY created_at DESC').all() as Repair[];
  res.json(rows.map(formatRepair));
});

router.get('/:id', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(404).json({ error: '工单不存在' });
  }

  const repair = db.prepare('SELECT * FROM repairs WHERE id = ?').get(id) as Repair | undefined;
  if (!repair) {
    return res.status(404).json({ error: '工单不存在' });
  }

  res.json(formatRepair(repair));
});

router.all('/:id/*', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(404).json({ error: '工单不存在' });
  }

  const repair = db.prepare('SELECT * FROM repairs WHERE id = ?').get(id) as Repair | undefined;
  if (!repair) {
    return res.status(404).json({ error: '工单不存在' });
  }

  const path = req.params[0] || '';
  const segments = path.split('/').filter(s => s.length > 0);

  if (segments.length === 2 && segments[0] === 'assign' && req.method === 'POST') {
    return handleAssign(repair, segments[1], res);
  }

  if (segments.length === 1 && segments[0] === 'complete' && req.method === 'POST') {
    return handleComplete(repair, res);
  }

  if (segments.length === 1 && segments[0] === 'process' && req.method === 'POST') {
    return handleProcess(repair, res);
  }

  if (segments.length === 1 && segments[0] === 'evaluate' && req.method === 'POST') {
    return handleEvaluate(repair, req, res);
  }

  res.status(404).json({ error: '路径段数不对或操作无效' });
});

function handleAssign(repair: Repair, staffIdStr: string, res: Response) {
  const staffId = parseInt(staffIdStr);
  if (isNaN(staffId)) {
    return res.status(404).json({ error: '维修人员不存在' });
  }

  const staff = db.prepare('SELECT * FROM staff WHERE id = ?').get(staffId) as Staff | undefined;
  if (!staff) {
    return res.status(404).json({ error: '维修人员不存在' });
  }

  if (repair.status !== '待接单') {
    if (repair.assigned_staff_id !== null) {
      return res.status(409).json({ error: '工单已被分配' });
    }
    return res.status(400).json({ error: '工单状态不允许分配' });
  }

  const now = Date.now();
  db.prepare(`
    UPDATE repairs 
    SET status = '已分配', assigned_staff_id = ?, assigned_at = ?
    WHERE id = ?
  `).run(staffId, now, repair.id);

  const updated = db.prepare('SELECT * FROM repairs WHERE id = ?').get(repair.id) as Repair;
  res.json(formatRepair(updated));
}

function handleProcess(repair: Repair, res: Response) {
  if (repair.status !== '已分配') {
    return res.status(400).json({ error: '非法状态跳转，仅已分配工单可开始处理' });
  }

  db.prepare(`UPDATE repairs SET status = '处理中' WHERE id = ?`).run(repair.id);
  const updated = db.prepare('SELECT * FROM repairs WHERE id = ?').get(repair.id) as Repair;
  res.json(formatRepair(updated));
}

function handleComplete(repair: Repair, res: Response) {
  if (repair.status !== '处理中') {
    return res.status(400).json({ error: '非法状态跳转，仅处理中工单可完成' });
  }

  const now = Date.now();
  db.prepare(`
    UPDATE repairs 
    SET status = '已完成', completed_at = ?
    WHERE id = ?
  `).run(now, repair.id);

  const updated = db.prepare('SELECT * FROM repairs WHERE id = ?').get(repair.id) as Repair;
  res.json(formatRepair(updated));
}

function handleEvaluate(repair: Repair, req: Request, res: Response) {
  if (repair.status !== '已完成') {
    return res.status(400).json({ error: '仅已完成工单可评价' });
  }

  const { score, comment } = req.body;
  if (score === undefined || score < 1 || score > 5) {
    return res.status(400).json({ error: '评价分数必须在1-5之间' });
  }

  db.prepare(`
    UPDATE repairs 
    SET evaluation_score = ?, evaluation_comment = ?
    WHERE id = ?
  `).run(score, comment || '', repair.id);

  const updated = db.prepare('SELECT * FROM repairs WHERE id = ?').get(repair.id) as Repair;
  res.json(formatRepair(updated));
}

function formatRepair(repair: Repair) {
  return {
    id: repair.id,
    category: repair.category,
    location: repair.location,
    description: repair.description,
    phone: repair.phone,
    status: repair.status,
    priority: repair.priority,
    assigned_staff_id: repair.assigned_staff_id,
    assigned_at: repair.assigned_at,
    completed_at: repair.completed_at,
    evaluation_score: repair.evaluation_score,
    evaluation_comment: repair.evaluation_comment,
    created_at: repair.created_at,
    response_deadline: repair.priority === '紧急' 
      ? repair.created_at + 2 * 60 * 60 * 1000 
      : repair.created_at + 24 * 60 * 60 * 1000
  };
}

router.get('/staff/list', (_req: Request, res: Response) => {
  const staff = db.prepare('SELECT * FROM staff').all();
  res.json(staff);
});

export default router;
