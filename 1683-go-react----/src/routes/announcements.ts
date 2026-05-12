import { Router, Request, Response } from 'express';
import db from '../db';

const router = Router();

const ANNOUNCEMENT_TYPES = ['通知', '紧急', '活动预告', '维修公告'];
const SCOPE_TYPES = ['全小区', '指定楼栋'];
const HOUR_24 = 24 * 60 * 60 * 1000;
const DAY_7 = 7 * 24 * 60 * 60 * 1000;
const DAY_3 = 3 * 24 * 60 * 60 * 1000;

interface Announcement {
  id: number;
  title: string;
  content: string;
  type: string;
  scope_type: string;
  scope_buildings: string | null;
  is_withdrawn: number;
  created_at: number;
  updated_at: number;
  edit_count: number;
  edit_summary: string | null;
}

router.post('/', (req: Request, res: Response) => {
  const { title, content, type, scope_type, scope_buildings } = req.body;

  if (!title || title.trim() === '' || !content || content.trim() === '') {
    return res.status(400).json({ error: '标题和正文不能为空' });
  }

  if (!type || !ANNOUNCEMENT_TYPES.includes(type)) {
    return res.status(400).json({ error: '公告类型无效，有效类型：通知、紧急、活动预告、维修公告' });
  }

  if (!scope_type || !SCOPE_TYPES.includes(scope_type)) {
    return res.status(400).json({ error: '发布范围无效' });
  }

  if (scope_type === '指定楼栋' && (!scope_buildings || scope_buildings.trim() === '')) {
    return res.status(400).json({ error: '指定楼栋时需填写楼栋信息' });
  }

  const now = Date.now();
  const stmt = db.prepare(`
    INSERT INTO announcements (title, content, type, scope_type, scope_buildings, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?)
  `);

  const result = stmt.run(
    title.trim(),
    content.trim(),
    type,
    scope_type,
    scope_type === '指定楼栋' ? scope_buildings.trim() : null,
    now,
    now
  );

  const announcement = db.prepare('SELECT * FROM announcements WHERE id = ?').get(result.lastInsertRowid) as Announcement;
  res.status(201).json(formatAnnouncement(announcement));
});

router.get('/', (req: Request, res: Response) => {
  const now = Date.now();
  const urgentCutoff = now - DAY_7;
  const normalCutoff = now - DAY_3;

  const rows = db.prepare(`
    SELECT * FROM announcements
    ORDER BY 
      CASE WHEN type = '紧急' AND created_at > ? THEN 0 ELSE 1 END ASC,
      CASE WHEN type != '紧急' AND created_at > ? THEN 0 ELSE 1 END ASC,
      created_at DESC
  `).all(urgentCutoff, normalCutoff) as Announcement[];

  res.json(rows.map(formatAnnouncement));
});

router.get('/:id', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(404).json({ error: '公告不存在' });
  }

  const announcement = db.prepare('SELECT * FROM announcements WHERE id = ?').get(id) as Announcement | undefined;
  if (!announcement) {
    return res.status(404).json({ error: '公告不存在' });
  }

  res.json(formatAnnouncement(announcement));
});

router.put('/:id', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(404).json({ error: '公告不存在' });
  }

  const announcement = db.prepare('SELECT * FROM announcements WHERE id = ?').get(id) as Announcement | undefined;
  if (!announcement) {
    return res.status(404).json({ error: '公告不存在' });
  }

  const now = Date.now();
  if (now - announcement.created_at > HOUR_24) {
    return res.status(400).json({ error: '公告发布超过24小时，不可修改' });
  }

  if (announcement.edit_count >= 3) {
    return res.status(400).json({ error: '公告修改次数已达上限（3次）' });
  }

  const { title, content, summary } = req.body;
  if (!title || title.trim() === '' || !content || content.trim() === '') {
    return res.status(400).json({ error: '标题和正文不能为空' });
  }

  const newTitle = title.trim();
  const newContent = content.trim();
  const editSummary = summary || '修改公告内容';

  const tx = db.transaction(() => {
    db.prepare(`
      INSERT INTO announcement_edits (announcement_id, old_title, old_content, new_title, new_content, summary, edited_at)
      VALUES (?, ?, ?, ?, ?, ?, ?)
    `).run(id, announcement.title, announcement.content, newTitle, newContent, editSummary, now);

    db.prepare(`
      UPDATE announcements 
      SET title = ?, content = ?, updated_at = ?, edit_count = edit_count + 1, edit_summary = ?
      WHERE id = ?
    `).run(newTitle, newContent, now, editSummary, id);
  });
  tx();

  const updated = db.prepare('SELECT * FROM announcements WHERE id = ?').get(id) as Announcement;
  res.json(formatAnnouncement(updated));
});

router.post('/:id/withdraw', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(404).json({ error: '公告不存在' });
  }

  const announcement = db.prepare('SELECT * FROM announcements WHERE id = ?').get(id) as Announcement | undefined;
  if (!announcement) {
    return res.status(404).json({ error: '公告不存在' });
  }

  if (announcement.is_withdrawn) {
    return res.status(409).json({ error: '公告已撤回，无法再次撤回' });
  }

  db.prepare('UPDATE announcements SET is_withdrawn = 1, updated_at = ? WHERE id = ?').run(Date.now(), id);
  const updated = db.prepare('SELECT * FROM announcements WHERE id = ?').get(id) as Announcement;
  res.json(formatAnnouncement(updated));
});

function formatAnnouncement(ann: Announcement) {
  return {
    id: ann.id,
    title: ann.is_withdrawn ? '已被撤回' : ann.title,
    content: ann.is_withdrawn ? '已被撤回' : ann.content,
    type: ann.type,
    scope_type: ann.scope_type,
    scope_buildings: ann.scope_buildings,
    is_withdrawn: !!ann.is_withdrawn,
    created_at: ann.created_at,
    updated_at: ann.updated_at,
    edit_count: ann.edit_count,
    edit_summary: ann.edit_summary
  };
}

export default router;
