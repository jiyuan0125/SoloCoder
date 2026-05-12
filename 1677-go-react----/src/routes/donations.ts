import express from 'express';
import { v4 as uuidv4 } from 'uuid';
import { db } from '../database';
import { Donation, Project, Certificate, CreateDonationRequest } from '../types';
import { generateCertificateNumber } from '../services/certificate-number-generator';
import dayjs from 'dayjs';

const router = express.Router();

router.post('/', (req, res) => {
  const body = req.body as CreateDonationRequest;
  
  if (!body.project_id || !body.donor_name || !body.id_card_last4 || body.amount === undefined) {
    return res.status(400).json({ error: '缺少必要字段' });
  }
  
  if (body.amount <= 0) {
    return res.status(400).json({ error: '捐赠金额必须大于0' });
  }
  
  if (body.id_card_last4.length !== 4 || !/^\d{4}$/.test(body.id_card_last4)) {
    return res.status(400).json({ error: '身份证号后四位必须是4位数字' });
  }
  
  const project = db.prepare(
    'SELECT * FROM projects WHERE id = ?'
  ).get(body.project_id) as Project | undefined;
  
  if (!project) {
    return res.status(400).json({ error: '项目不存在' });
  }
  
  const now = dayjs();
  const donatedAt = body.donated_at ? dayjs(body.donated_at).format('YYYY-MM-DD HH:mm:ss') : now.format('YYYY-MM-DD HH:mm:ss');
  const createdAt = now.format('YYYY-MM-DD HH:mm:ss');
  
  const transaction = db.transaction(() => {
    const donationId = uuidv4();
    const certificateNumber = generateCertificateNumber();
    
    db.prepare(
      `INSERT INTO donations (id, project_id, donor_name, id_card_last4, amount, certificate_number, donated_at, created_at)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
    ).run(donationId, body.project_id, body.donor_name, body.id_card_last4, body.amount, certificateNumber, donatedAt, createdAt);
    
    const certificateId = uuidv4();
    db.prepare(
      `INSERT INTO certificates (id, certificate_number, donation_id, donor_name, id_card_last4, project_id, amount, issued_at, created_at)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
    ).run(certificateId, certificateNumber, donationId, body.donor_name, body.id_card_last4, body.project_id, body.amount, createdAt, createdAt);
    
    const donation = db.prepare('SELECT * FROM donations WHERE id = ?').get(donationId) as Donation;
    const certificate = db.prepare('SELECT * FROM certificates WHERE id = ?').get(certificateId) as Certificate;
    
    return { donation, certificate };
  });
  
  try {
    const result = transaction();
    return res.status(201).json({
      donation: result.donation,
      certificate: result.certificate
    });
  } catch (err) {
    return res.status(500).json({ error: '创建捐赠记录失败' });
  }
});

router.get('/:id', (req, res) => {
  const donation = db.prepare('SELECT * FROM donations WHERE id = ?').get(req.params.id) as Donation | undefined;
  
  if (!donation) {
    return res.status(404).json({ error: '捐赠记录不存在' });
  }
  
  const certificate = db.prepare(
    'SELECT * FROM certificates WHERE donation_id = ?'
  ).get(req.params.id) as Certificate | undefined;
  
  const voucher = db.prepare(
    `SELECT * FROM tax_deduction_vouchers 
     WHERE donation_id = ? AND status = 'active'
     ORDER BY created_at DESC`
  ).get(req.params.id);
  
  return res.json({
    donation,
    certificate,
    voucher: voucher || null
  });
});

router.get('/', (_req, res) => {
  const donations = db.prepare(
    'SELECT * FROM donations ORDER BY created_at DESC'
  ).all() as Donation[];
  
  return res.json(donations);
});

router.get('/by-project/:projectId', (req, res) => {
  const project = db.prepare(
    'SELECT id FROM projects WHERE id = ?'
  ).get(req.params.projectId) as Project | undefined;
  
  if (!project) {
    return res.status(404).json({ error: '项目不存在' });
  }
  
  const donations = db.prepare(
    'SELECT * FROM donations WHERE project_id = ? ORDER BY created_at DESC'
  ).all(req.params.projectId) as Donation[];
  
  return res.json(donations);
});

export default router;
