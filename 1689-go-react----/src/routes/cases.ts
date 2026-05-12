import { Router, Response } from 'express';
import db from '../database';
import { AuthRequest, authenticateToken, requireSocialWorker, requireSupervisor } from '../middleware/auth';

const router = Router();
router.use(authenticateToken);

function addMonths(date: Date, months: number): Date {
  const result = new Date(date);
  result.setMonth(result.getMonth() + months);
  return result;
}

function formatDate(date: Date): string {
  return date.toISOString().split('T')[0];
}

router.get('/', requireSocialWorker, (req: AuthRequest, res: Response) => {
  const cases = db.prepare(`
    SELECT c.*, u.name as social_worker_name 
    FROM cases c 
    JOIN users u ON c.social_worker_id = u.id
    ORDER BY c.created_at DESC
  `).all();
  res.json(cases);
});

router.get('/:id', requireSocialWorker, (req: AuthRequest, res: Response) => {
  const caseData = db.prepare(`
    SELECT c.*, u.name as social_worker_name 
    FROM cases c 
    JOIN users u ON c.social_worker_id = u.id
    WHERE c.id = ?
  `).get(req.params.id) as any;

  if (!caseData) {
    res.status(404).json({ error: '个案不存在' });
    return;
  }

  const sessions = db.prepare('SELECT * FROM case_sessions WHERE case_id = ? ORDER BY session_date DESC').all(req.params.id);
  const extensions = db.prepare('SELECT * FROM case_extensions WHERE case_id = ? ORDER BY created_at DESC').all(req.params.id);

  res.json({ ...caseData, sessions, extensions });
});

router.post('/', requireSocialWorker, (req: AuthRequest, res: Response) => {
  const {
    client_name,
    client_info,
    problem_description,
    needs_assessment,
    intervention_plan,
    supervisor_approved,
  } = req.body;

  if (!client_name || !client_info || !client_info.trim()) {
    res.status(400).json({ error: '案主信息不能为空' });
    return;
  }

  if (!problem_description || !needs_assessment || !intervention_plan) {
    res.status(400).json({ error: '缺少必要字段' });
    return;
  }

  const socialWorkerId = req.user?.id;

  const activeCases = db.prepare(`
    SELECT COUNT(*) as count FROM cases 
    WHERE social_worker_id = ? AND status IN ('open', 'in_progress', 'extended')
  `).get(socialWorkerId) as { count: number };

  if (activeCases.count >= 15) {
    if (req.user?.role !== 'supervisor' && !supervisor_approved) {
      res.status(403).json({ error: '进行中个案已达15个上限，需要主管特批' });
      return;
    }
  }

  const today = new Date();
  const estimatedEnd = addMonths(today, 6);

  const info = db.prepare(`
    INSERT INTO cases 
    (social_worker_id, client_name, client_info, problem_description, needs_assessment, intervention_plan, 
     status, start_date, estimated_end_date, supervisor_approved)
    VALUES (?, ?, ?, ?, ?, ?, 'in_progress', ?, ?, ?)
  `).run(
    socialWorkerId,
    client_name,
    client_info,
    problem_description,
    needs_assessment,
    intervention_plan,
    formatDate(today),
    formatDate(estimatedEnd),
    supervisor_approved ? 1 : 0
  );

  const newCase = db.prepare('SELECT * FROM cases WHERE id = ?').get(info.lastInsertRowid);
  res.status(201).json(newCase);
});

router.put('/:id', requireSocialWorker, (req: AuthRequest, res: Response) => {
  const existingCase = db.prepare('SELECT * FROM cases WHERE id = ?').get(req.params.id) as any;

  if (!existingCase) {
    res.status(404).json({ error: '个案不存在' });
    return;
  }

  if (existingCase.status === 'archived') {
    res.status(405).json({ error: '已归档的个案不可修改' });
    return;
  }

  const {
    client_name,
    client_info,
    problem_description,
    needs_assessment,
    intervention_plan,
  } = req.body;

  db.prepare(`
    UPDATE cases SET 
    client_name = COALESCE(?, client_name),
    client_info = COALESCE(?, client_info),
    problem_description = COALESCE(?, problem_description),
    needs_assessment = COALESCE(?, needs_assessment),
    intervention_plan = COALESCE(?, intervention_plan)
    WHERE id = ?
  `).run(
    client_name,
    client_info,
    problem_description,
    needs_assessment,
    intervention_plan,
    req.params.id
  );

  const updatedCase = db.prepare('SELECT * FROM cases WHERE id = ?').get(req.params.id);
  res.json(updatedCase);
});

router.post('/:id/sessions', requireSocialWorker, (req: AuthRequest, res: Response) => {
  const existingCase = db.prepare('SELECT * FROM cases WHERE id = ?').get(req.params.id) as any;

  if (!existingCase) {
    res.status(404).json({ error: '个案不存在' });
    return;
  }

  if (existingCase.status === 'archived') {
    res.status(405).json({ error: '已归档的个案不可添加记录' });
    return;
  }

  const { session_date, duration, summary, next_plan } = req.body;

  if (!session_date || !session_date.trim()) {
    res.status(400).json({ error: '面谈日期不能为空' });
    return;
  }

  if (!duration || !summary || !next_plan) {
    res.status(400).json({ error: '缺少必要字段' });
    return;
  }

  const info = db.prepare(`
    INSERT INTO case_sessions (case_id, session_date, duration, summary, next_plan)
    VALUES (?, ?, ?, ?, ?)
  `).run(req.params.id, session_date, duration, summary, next_plan);

  const session = db.prepare('SELECT * FROM case_sessions WHERE id = ?').get(info.lastInsertRowid);
  res.status(201).json(session);
});

router.get('/:id/sessions', requireSocialWorker, (req: AuthRequest, res: Response) => {
  const existingCase = db.prepare('SELECT * FROM cases WHERE id = ?').get(req.params.id) as any;

  if (!existingCase) {
    res.status(404).json({ error: '个案不存在' });
    return;
  }

  const sessions = db.prepare('SELECT * FROM case_sessions WHERE case_id = ? ORDER BY session_date DESC').all(req.params.id);
  res.json(sessions);
});

router.post('/:id/extensions', requireSocialWorker, (req: AuthRequest, res: Response) => {
  const existingCase = db.prepare('SELECT * FROM cases WHERE id = ?').get(req.params.id) as any;

  if (!existingCase) {
    res.status(404).json({ error: '个案不存在' });
    return;
  }

  const { reason, new_end_date } = req.body;

  if (!reason || !new_end_date) {
    res.status(400).json({ error: '缺少延期原因或新截止日期' });
    return;
  }

  const info = db.prepare(`
    INSERT INTO case_extensions (case_id, reason, new_end_date)
    VALUES (?, ?, ?)
  `).run(req.params.id, reason, new_end_date);

  const extension = db.prepare('SELECT * FROM case_extensions WHERE id = ?').get(info.lastInsertRowid);
  res.status(201).json(extension);
});

router.put('/extensions/:extensionId/approve', requireSupervisor, (req: AuthRequest, res: Response) => {
  const extension = db.prepare('SELECT * FROM case_extensions WHERE id = ?').get(req.params.extensionId) as any;

  if (!extension) {
    res.status(404).json({ error: '延期申请不存在' });
    return;
  }

  const transaction = db.transaction(() => {
    db.prepare(`
      UPDATE case_extensions SET approved = 1, approved_by = ? WHERE id = ?
    `).run(req.user?.id, req.params.extensionId);

    db.prepare(`
      UPDATE cases SET status = 'extended', estimated_end_date = ? WHERE id = ?
    `).run(extension.new_end_date, extension.case_id);
  });

  transaction();

  res.json({ message: '延期申请已批准' });
});

router.post('/:id/close', requireSocialWorker, (req: AuthRequest, res: Response) => {
  const existingCase = db.prepare('SELECT * FROM cases WHERE id = ?').get(req.params.id) as any;

  if (!existingCase) {
    res.status(404).json({ error: '个案不存在' });
    return;
  }

  if (existingCase.status === 'closed' || existingCase.status === 'archived') {
    res.status(400).json({ error: '个案已经结案或已归档' });
    return;
  }

  const sessions = db.prepare('SELECT * FROM case_sessions WHERE case_id = ?').all(req.params.id);

  if (sessions.length === 0) {
    res.status(400).json({ error: '未完成介入计划，不能结案（无面谈记录）' });
    return;
  }

  const today = formatDate(new Date());

  db.prepare(`
    UPDATE cases SET status = 'closed', actual_end_date = ? WHERE id = ?
  `).run(today, req.params.id);

  const updatedCase = db.prepare('SELECT * FROM cases WHERE id = ?').get(req.params.id);
  res.json(updatedCase);
});

router.post('/:id/archive', requireSocialWorker, (req: AuthRequest, res: Response) => {
  const existingCase = db.prepare('SELECT * FROM cases WHERE id = ?').get(req.params.id) as any;

  if (!existingCase) {
    res.status(404).json({ error: '个案不存在' });
    return;
  }

  if (existingCase.status !== 'closed') {
    res.status(400).json({ error: '只有已结案的个案才能归档' });
    return;
  }

  const sessions = db.prepare('SELECT * FROM case_sessions WHERE case_id = ?').all(req.params.id);

  const archiveData = JSON.stringify({
    case: existingCase,
    sessions: sessions,
  });

  const transaction = db.transaction(() => {
    db.prepare(`
      INSERT INTO case_archives (case_id, archive_data) VALUES (?, ?)
    `).run(req.params.id, archiveData);

    db.prepare(`
      UPDATE cases SET status = 'archived' WHERE id = ?
    `).run(req.params.id);
  });

  transaction();

  const archivedCase = db.prepare('SELECT * FROM cases WHERE id = ?').get(req.params.id);
  res.json({ case: archivedCase, archived: true });
});

router.post('/:id/reopen-request', requireSocialWorker, (req: AuthRequest, res: Response) => {
  const existingCase = db.prepare('SELECT * FROM cases WHERE id = ?').get(req.params.id) as any;

  if (!existingCase) {
    res.status(404).json({ error: '个案不存在' });
    return;
  }

  if (existingCase.status !== 'archived') {
    res.status(400).json({ error: '只有已归档的个案才能申请重新打开' });
    return;
  }

  const { reason } = req.body;

  if (!reason) {
    res.status(400).json({ error: '请提供重新打开的原因' });
    return;
  }

  const existingPending = db.prepare(`
    SELECT * FROM case_reopen_requests WHERE case_id = ? AND status = 'pending'
  `).get(req.params.id);

  if (existingPending) {
    res.status(400).json({ error: '已有待审批的重新打开申请' });
    return;
  }

  const info = db.prepare(`
    INSERT INTO case_reopen_requests (case_id, requested_by, reason)
    VALUES (?, ?, ?)
  `).run(req.params.id, req.user?.id, reason);

  const request = db.prepare('SELECT * FROM case_reopen_requests WHERE id = ?').get(info.lastInsertRowid);
  res.status(201).json(request);
});

router.put('/reopen-requests/:requestId/approve', requireSupervisor, (req: AuthRequest, res: Response) => {
  const request = db.prepare('SELECT * FROM case_reopen_requests WHERE id = ?').get(req.params.requestId) as any;

  if (!request) {
    res.status(404).json({ error: '重新打开申请不存在' });
    return;
  }

  if (request.status !== 'pending') {
    res.status(400).json({ error: '申请已处理' });
    return;
  }

  const today = new Date();
  const estimatedEnd = addMonths(today, 6);

  const transaction = db.transaction(() => {
    db.prepare(`
      UPDATE case_reopen_requests 
      SET status = 'approved', approved_by = ?, approved_at = ?
      WHERE id = ?
    `).run(req.user?.id, today.toISOString(), req.params.requestId);

    db.prepare(`
      UPDATE cases 
      SET status = 'in_progress', 
          start_date = ?, 
          estimated_end_date = ?,
          actual_end_date = NULL
      WHERE id = ?
    `).run(formatDate(today), formatDate(estimatedEnd), request.case_id);
  });

  transaction();

  const reopenedCase = db.prepare('SELECT * FROM cases WHERE id = ?').get(request.case_id);
  res.json({ message: '个案已重新打开', case: reopenedCase });
});

export default router;
