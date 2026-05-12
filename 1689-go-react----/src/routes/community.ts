import { Router, Response } from 'express';
import db from '../database';
import { AuthRequest, authenticateToken, requireSocialWorker } from '../middleware/auth';

const router = Router();
router.use(authenticateToken);

router.get('/', requireSocialWorker, (req: AuthRequest, res: Response) => {
  const services = db.prepare(`
    SELECT cs.*, u.name as creator_name 
    FROM community_services cs 
    JOIN users u ON cs.created_by = u.id
    ORDER BY cs.service_date DESC
  `).all();
  res.json(services);
});

router.post('/', requireSocialWorker, (req: AuthRequest, res: Response) => {
  const {
    service_date,
    content,
    participants,
    effect,
  } = req.body;

  if (!service_date || !content || participants === undefined || !effect) {
    res.status(400).json({ error: '缺少必要字段' });
    return;
  }

  const info = db.prepare(`
    INSERT INTO community_services (service_date, content, participants, effect, created_by)
    VALUES (?, ?, ?, ?, ?)
  `).run(service_date, content, participants, effect, req.user?.id);

  const service = db.prepare('SELECT * FROM community_services WHERE id = ?').get(info.lastInsertRowid);
  res.status(201).json(service);
});

router.put('/:id', requireSocialWorker, (req: AuthRequest, res: Response) => {
  const service = db.prepare('SELECT * FROM community_services WHERE id = ?').get(req.params.id) as any;

  if (!service) {
    res.status(404).json({ error: '社区服务记录不存在' });
    return;
  }

  const {
    service_date,
    content,
    participants,
    effect,
  } = req.body;

  db.prepare(`
    UPDATE community_services SET 
    service_date = COALESCE(?, service_date),
    content = COALESCE(?, content),
    participants = COALESCE(?, participants),
    effect = COALESCE(?, effect)
    WHERE id = ?
  `).run(service_date, content, participants, effect, req.params.id);

  const updatedService = db.prepare('SELECT * FROM community_services WHERE id = ?').get(req.params.id);
  res.json(updatedService);
});

export default router;
