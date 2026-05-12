import { Router, Request, Response } from 'express';
import db from '../db';
import { nowISO, isInInspectionPeriod, isInspectionReminderDue, shouldAutoCancel } from '../utils';
import type { InspectionResult, OrganizationStatus } from '../types';

const router = Router();

router.post('/submit', (req: Request, res: Response) => {
  const { organizationId, year, annualReport, financialAudit, activityList } = req.body;

  if (!organizationId || !year || !annualReport || !financialAudit || !activityList) {
    return res.status(400).json({ error: '缺少必填字段' });
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
    return res.status(400).json({ error: '只有已注册组织才能参加年检' });
  }

  if (!isInInspectionPeriod(new Date(), year)) {
    return res.status(400).json({ error: '不在年检申报期内' });
  }

  const existing = db.prepare('SELECT id FROM annual_inspections WHERE organization_id = ? AND year = ?').get(organizationId, year);
  if (existing) {
    return res.status(409).json({ error: '该年度年检已提交' });
  }

  const now = nowISO();
  const insert = db.prepare(`
    INSERT INTO annual_inspections (organization_id, year, annual_report, financial_audit, activity_list, submit_time)
    VALUES (?, ?, ?, ?, ?, ?)
  `);
  const result = insert.run(organizationId, year, annualReport, financialAudit, activityList, now);

  res.status(201).json({ id: result.lastInsertRowid, message: '年检材料已提交' });
});

router.post('/:id/review', (req: Request, res: Response) => {
  const { result } = req.body as { result: InspectionResult };

  if (!['合格', '基本合格', '不合格'].includes(result)) {
    return res.status(400).json({ error: '年检结果无效' });
  }

  const inspection = db.prepare('SELECT * FROM annual_inspections WHERE id = ?').get(req.params.id);
  if (!inspection) {
    return res.status(404).json({ error: '年检记录不存在' });
  }

  const inspectionData = inspection as any;
  const orgId = inspectionData.organization_id;

  const org = db.prepare('SELECT * FROM organizations WHERE id = ?').get(orgId);
  if ((org as any).status === '已注销') {
    return res.status(405).json({ error: '组织已注销，不可操作' });
  }

  const now = nowISO();
  db.prepare('UPDATE annual_inspections SET result = ?, review_time = ? WHERE id = ?').run(result, now, req.params.id);

  const allInspections = db.prepare('SELECT result FROM annual_inspections WHERE organization_id = ? AND result IS NOT NULL ORDER BY year ASC').all(orgId);
  const results = allInspections.map((r: any) => r.result as InspectionResult);

  if (shouldAutoCancel(results)) {
    db.prepare('UPDATE organizations SET status = ?, updated_at = ? WHERE id = ?').run('已注销', now, orgId);
    res.json({ message: '年检结果已录入，该组织连续两年不合格，已自动注销' });
  } else {
    res.json({ message: '年检结果已录入' });
  }
});

router.post('/reminders/generate', (req: Request, res: Response) => {
  const { year } = req.body;
  if (!year) {
    return res.status(400).json({ error: '请指定年份' });
  }

  if (!isInspectionReminderDue(new Date(), year)) {
    return res.status(200).json({ message: '不在提醒期内，无需生成提醒' });
  }

  const orgs = db.prepare('SELECT id FROM organizations WHERE status = ?').all('已注册');

  const now = nowISO();
  const insert = db.prepare('INSERT INTO reminders (organization_id, type, content, created_at) VALUES (?, ?, ?, ?)');

  let count = 0;
  const reminderType = '年检_' + year;
  const reminderContent = '请在 ' + year + ' 年 3 月 1 日至 5 月 31 日期间提交年检材料';

  for (const org of orgs as any[]) {
    const existing = db.prepare('SELECT id FROM reminders WHERE organization_id = ? AND type = ?').get(org.id, reminderType);
    if (!existing) {
      insert.run(org.id, reminderType, reminderContent, now);
      count++;
    }
  }

  res.json({ message: '已为 ' + count + ' 个组织生成提醒' });
});

router.get('/reminders/:orgId', (req: Request, res: Response) => {
  const reminders = db.prepare('SELECT * FROM reminders WHERE organization_id = ? ORDER BY created_at DESC').all(req.params.orgId);
  res.json(reminders);
});

router.post('/:year/mark-missing', (req: Request, res: Response) => {
  const { year } = req.params;
  const now = nowISO();

  const orgs = db.prepare('SELECT id FROM organizations WHERE status = ?').all('已注册') as any[];
  let count = 0;

  for (const org of orgs) {
    const hasInspection = db.prepare('SELECT id FROM annual_inspections WHERE organization_id = ? AND year = ?').get(org.id, parseInt(year));
    if (!hasInspection) {
      db.prepare(`
        INSERT INTO annual_inspections (organization_id, year, annual_report, financial_audit, activity_list, result, submit_time, review_time)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
      `).run(org.id, parseInt(year), '未参检', '未参检', '未参检', '不合格', now, now);
      count++;
    }
  }

  res.json({ message: '已将 ' + count + ' 个未参检组织标记为不合格' });
});

export default router;
