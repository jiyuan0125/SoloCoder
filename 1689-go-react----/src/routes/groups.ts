import { Router, Response } from 'express';
import db from '../database';
import { AuthRequest, authenticateToken, requireSocialWorker } from '../middleware/auth';

const router = Router();
router.use(authenticateToken);

router.get('/', requireSocialWorker, (req: AuthRequest, res: Response) => {
  const groups = db.prepare(`
    SELECT g.*, u.name as creator_name 
    FROM group_activities g 
    JOIN users u ON g.created_by = u.id
    ORDER BY g.created_at DESC
  `).all();
  res.json(groups);
});

router.get('/:id', requireSocialWorker, (req: AuthRequest, res: Response) => {
  const group = db.prepare(`
    SELECT g.*, u.name as creator_name 
    FROM group_activities g 
    JOIN users u ON g.created_by = u.id
    WHERE g.id = ?
  `).get(req.params.id) as any;

  if (!group) {
    res.status(404).json({ error: '活动不存在' });
    return;
  }

  const members = db.prepare(`
    SELECT m.*, u.name as approver_name 
    FROM group_members m
    LEFT JOIN users u ON m.approved_by = u.id
    WHERE m.group_activity_id = ?
    ORDER BY m.id DESC
  `).all(req.params.id);

  const attendance = db.prepare(`
    SELECT * FROM group_attendance WHERE group_activity_id = ?
    ORDER BY session_number, member_id
  `).all(req.params.id);

  res.json({ ...group, members, attendance });
});

router.post('/', requireSocialWorker, (req: AuthRequest, res: Response) => {
  const {
    name,
    goal,
    activity_count,
    time_details,
    location,
    max_participants,
  } = req.body;

  if (!name || !goal || !activity_count || !time_details || !location || !max_participants) {
    res.status(400).json({ error: '缺少必要字段' });
    return;
  }

  if (max_participants < 6 || max_participants > 12) {
    res.status(400).json({ error: '人数上限必须在6到12人之间' });
    return;
  }

  const info = db.prepare(`
    INSERT INTO group_activities 
    (name, goal, activity_count, time_details, location, max_participants, status, created_by)
    VALUES (?, ?, ?, ?, ?, ?, 'planning', ?)
  `).run(
    name,
    goal,
    activity_count,
    time_details,
    location,
    max_participants,
    req.user?.id
  );

  const newGroup = db.prepare('SELECT * FROM group_activities WHERE id = ?').get(info.lastInsertRowid);
  res.status(201).json(newGroup);
});

router.put('/:id', requireSocialWorker, (req: AuthRequest, res: Response) => {
  const group = db.prepare('SELECT * FROM group_activities WHERE id = ?').get(req.params.id) as any;

  if (!group) {
    res.status(404).json({ error: '活动不存在' });
    return;
  }

  const {
    name,
    goal,
    activity_count,
    time_details,
    location,
    max_participants,
    status,
  } = req.body;

  if (max_participants !== undefined && (max_participants < 6 || max_participants > 12)) {
    res.status(400).json({ error: '人数上限必须在6到12人之间' });
    return;
  }

  db.prepare(`
    UPDATE group_activities SET 
    name = COALESCE(?, name),
    goal = COALESCE(?, goal),
    activity_count = COALESCE(?, activity_count),
    time_details = COALESCE(?, time_details),
    location = COALESCE(?, location),
    max_participants = COALESCE(?, max_participants),
    status = COALESCE(?, status)
    WHERE id = ?
  `).run(
    name,
    goal,
    activity_count,
    time_details,
    location,
    max_participants,
    status,
    req.params.id
  );

  const updatedGroup = db.prepare('SELECT * FROM group_activities WHERE id = ?').get(req.params.id);
  res.json(updatedGroup);
});

router.post('/:id/members', (req: AuthRequest, res: Response) => {
  const group = db.prepare('SELECT * FROM group_activities WHERE id = ?').get(req.params.id) as any;

  if (!group) {
    res.status(404).json({ error: '活动不存在' });
    return;
  }

  const { name, contact_info } = req.body;

  if (!name) {
    res.status(400).json({ error: '成员姓名不能为空' });
    return;
  }

  const approvedCount = db.prepare(`
    SELECT COUNT(*) as count FROM group_members 
    WHERE group_activity_id = ? AND status = 'approved'
  `).get(req.params.id) as { count: number };

  if (approvedCount.count >= group.max_participants) {
    res.status(400).json({ error: '活动人数已达上限' });
    return;
  }

  const info = db.prepare(`
    INSERT INTO group_members (group_activity_id, name, contact_info, status)
    VALUES (?, ?, ?, 'pending')
  `).run(req.params.id, name, contact_info);

  const member = db.prepare('SELECT * FROM group_members WHERE id = ?').get(info.lastInsertRowid);
  res.status(201).json(member);
});

router.put('/:id/members/:memberId/approve', requireSocialWorker, (req: AuthRequest, res: Response) => {
  const group = db.prepare('SELECT * FROM group_activities WHERE id = ?').get(req.params.id) as any;

  if (!group) {
    res.status(404).json({ error: '活动不存在' });
    return;
  }

  const member = db.prepare('SELECT * FROM group_members WHERE id = ? AND group_activity_id = ?').get(req.params.memberId, req.params.id) as any;

  if (!member) {
    res.status(404).json({ error: '成员不存在' });
    return;
  }

  const approvedCount = db.prepare(`
    SELECT COUNT(*) as count FROM group_members 
    WHERE group_activity_id = ? AND status = 'approved'
  `).get(req.params.id) as { count: number };

  if (approvedCount.count >= group.max_participants) {
    res.status(400).json({ error: '活动人数已达上限' });
    return;
  }

  db.prepare(`
    UPDATE group_members SET status = 'approved', approved_by = ?, approved_at = CURRENT_TIMESTAMP
    WHERE id = ?
  `).run(req.user?.id, req.params.memberId);

  const updatedMember = db.prepare('SELECT * FROM group_members WHERE id = ?').get(req.params.memberId);
  res.json(updatedMember);
});

router.post('/:id/attendance', requireSocialWorker, (req: AuthRequest, res: Response) => {
  const group = db.prepare('SELECT * FROM group_activities WHERE id = ?').get(req.params.id) as any;

  if (!group) {
    res.status(404).json({ error: '活动不存在' });
    return;
  }

  const { session_number, records } = req.body;

  if (!session_number || !records || !Array.isArray(records)) {
    res.status(400).json({ error: '缺少必要字段' });
    return;
  }

  const transaction = db.transaction(() => {
    for (const record of records) {
      const { member_id, attended, leave_approved } = record;

      db.prepare(`
        INSERT INTO group_attendance (group_activity_id, member_id, session_number, attended, leave_approved)
        VALUES (?, ?, ?, ?, ?)
      `).run(req.params.id, member_id, session_number, attended ? 1 : 0, leave_approved ? 1 : 0);

      const member = db.prepare('SELECT * FROM group_members WHERE id = ?').get(member_id) as any;

      if (member) {
        let newAbsences = member.consecutive_absences;

        if (attended || leave_approved) {
          newAbsences = 0;
        } else {
          newAbsences += 1;
        }

        if (newAbsences >= 2) {
          db.prepare(`
            UPDATE group_members SET status = 'removed', consecutive_absences = ? WHERE id = ?
          `).run(newAbsences, member_id);
        } else {
          db.prepare(`
            UPDATE group_members SET consecutive_absences = ? WHERE id = ?
          `).run(newAbsences, member_id);
        }
      }
    }
  });

  transaction();

  if (session_number >= group.activity_count) {
    db.prepare(`
      UPDATE group_activities SET status = 'completed', completed_count = ? WHERE id = ?
    `).run(session_number, req.params.id);
  } else {
    db.prepare(`
      UPDATE group_activities SET completed_count = ? WHERE id = ?
    `).run(session_number, req.params.id);
  }

  res.json({ message: '考勤记录已保存', session_number });
});

router.post('/:id/summary', requireSocialWorker, (req: AuthRequest, res: Response) => {
  const group = db.prepare('SELECT * FROM group_activities WHERE id = ?').get(req.params.id) as any;

  if (!group) {
    res.status(404).json({ error: '活动不存在' });
    return;
  }

  if (group.status !== 'completed') {
    res.status(400).json({ error: '只有已完成的活动才能生成总结报告' });
    return;
  }

  const { summary_report } = req.body;

  if (!summary_report) {
    res.status(400).json({ error: '总结报告内容不能为空' });
    return;
  }

  db.prepare(`
    UPDATE group_activities SET summary_report = ? WHERE id = ?
  `).run(summary_report, req.params.id);

  const updatedGroup = db.prepare('SELECT * FROM group_activities WHERE id = ?').get(req.params.id);
  res.json(updatedGroup);
});

export default router;
