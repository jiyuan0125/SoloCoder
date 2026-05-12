import { Router, Request, Response } from 'express';
import db from '../db';
import {
  isValidCreditCode, normalizeName, nowISO, VALID_TYPES, generateCertificateNo, getCertificateDates, isWithinReSubmitWindow } from '../utils';
import type { OrganizationStatus } from '../types';

const router = Router();

interface CreateOrgRequest {
  name: string;
  creditCode: string;
  type: string;
  businessSupervisor: string;
  legalRepresentative: string;
  address: string;
  businessScope: string;
}

router.post('/', (req: Request, res: Response) => {
  const body = req.body as Partial<CreateOrgRequest>;
  const { name, creditCode, type, businessSupervisor, legalRepresentative, address, businessScope } = body;

  if (!name || !creditCode || !type || !businessSupervisor || !legalRepresentative || !address || !businessScope) {
    return res.status(400).json({ error: '缺少必填字段' });
  }

  if (!isValidCreditCode(creditCode)) {
    return res.status(400).json({ error: '统一社会信用代码格式错误' });
  }

  if (!VALID_TYPES.includes(type as any)) {
    return res.status(400).json({ error: '组织类型不在允许范围内' });
  }

  const normalized = normalizeName(name);
  const existing = db.prepare('SELECT id FROM organizations WHERE normalized_name = ?').get(normalized);
  if (existing) {
    return res.status(409).json({ error: '组织名称已存在' });
  }

  const existingCode = db.prepare('SELECT id FROM organizations WHERE credit_code = ?').get(creditCode);
  if (existingCode) {
    return res.status(409).json({ error: '统一社会信用代码已存在' });
  }

  const now = nowISO();
  const insert = db.prepare(`
    INSERT INTO organizations (name, normalized_name, credit_code, type, business_supervisor, legal_representative, address, business_scope, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `);
  const result = insert.run(name, normalized, creditCode, type, businessSupervisor, legalRepresentative, address, businessScope, now, now);

  const orgId = result.lastInsertRowid as number;

  db.prepare('INSERT INTO registration_applications (organization_id, status, submit_time) VALUES (?, ?, ?)').run(orgId, '待审核', now);

  res.status(201).json({ id: orgId, message: '注册申请已提交，状态：注册中' });
});

router.get('/:id', (req: Request, res: Response) => {
  const org = db.prepare('SELECT * FROM organizations WHERE id = ?').get(req.params.id);
  if (!org) {
    return res.status(404).json({ error: '组织不存在' });
  }
  res.json(org);
});

router.post('/:id/review', (req: Request, res: Response) => {
  const { approved, rejectReason } = req.body;
  const org = db.prepare('SELECT * FROM organizations WHERE id = ?').get(req.params.id);

  if (!org) {
    return res.status(404).json({ error: '组织不存在' });
  }

  if ((org as any).status === '已注销') {
    return res.status(405).json({ error: '组织已注销，不可操作' });
  }

  const now = nowISO();

  if (approved === true) {
    const { issueDate, expiryDate } = getCertificateDates();
    const certNo = generateCertificateNo();

    db.prepare('UPDATE organizations SET status = ?, certificate_no = ?, certificate_issue_date = ?, certificate_expiry_date = ?, updated_at = ? WHERE id = ?').run('已注册', certNo, issueDate, expiryDate, now, req.params.id);

    db.prepare('UPDATE registration_applications SET status = ? WHERE organization_id = ? AND status = ?').run('已通过', req.params.id, '待审核');

    res.json({
      message: '审核通过，组织已注册',
      certificateNo: certNo,
      issueDate,
      expiryDate
    });
  } else {
    db.prepare('UPDATE organizations SET re_submit_count = re_submit_count + 1, last_reject_time = ?, updated_at = ? WHERE id = ?').run(now, now, req.params.id);

    db.prepare('UPDATE registration_applications SET status = ?, reject_reason = ? WHERE organization_id = ? AND status = ?').run('已驳回', rejectReason || '', req.params.id, '待审核');

    res.json({ message: '审核驳回，请在30天内补材料重新提交' });
  }
});

router.post('/:id/resubmit', (req: Request, res: Response) => {
  const org = db.prepare('SELECT * FROM organizations WHERE id = ?').get(req.params.id);
  if (!org) {
    return res.status(404).json({ error: '组织不存在' });
  }

  const orgData = org as any;

  if (orgData.status === '已注销') {
    return res.status(405).json({ error: '组织已注销，不可操作' });
  }

  if (orgData.re_submit_count >= 3) {
    return res.status(400).json({ error: '重新提交次数已达上限' });
  }

  if (orgData.last_reject_time && !isWithinReSubmitWindow(orgData.last_reject_time)) {
    return res.status(400).json({ error: '已超过30天补材料期限' });
  }

  const now = nowISO();
  db.prepare('INSERT INTO registration_applications (organization_id, status, submit_time) VALUES (?, ?, ?)').run(req.params.id, '待审核', now);
  db.prepare('UPDATE organizations SET updated_at = ? WHERE id = ?').run(now, req.params.id);

  res.json({ message: '已重新提交申请' });
});

router.post('/:id/cancel', (req: Request, res: Response) => {
  const org = db.prepare('SELECT * FROM organizations WHERE id = ?').get(req.params.id);
  if (!org) {
    return res.status(404).json({ error: '组织不存在' });
  }

  const orgData = org as any;

  if (orgData.status === '已注销') {
    return res.status(405).json({ error: '组织已注销，不可操作' });
  }

  if (orgData.status !== '已注册') {
    return res.status(400).json({ error: '只能注销已注册组织' });
  }

  const now = nowISO();
  db.prepare('UPDATE organizations SET status = ?, updated_at = ? WHERE id = ?').run('已注销', now, req.params.id);

  res.json({ message: '组织已注销，记录已保存' });
});

export default router;
