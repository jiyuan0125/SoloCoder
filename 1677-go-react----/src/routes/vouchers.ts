import express from 'express';
import { v4 as uuidv4 } from 'uuid';
import { db } from '../database';
import { Donation, Project, TaxDeductionVoucher, CreateVoucherRequest, DonorAnnualSummary } from '../types';
import { generateVoucherNumber } from '../services/certificate-number-generator';
import dayjs from 'dayjs';

const router = express.Router();

function getOrCreateDonorSummary(donorName: string, idCardLast4: string, year: number, taxIncomeBase: number): DonorAnnualSummary {
  const existing = db.prepare(
    `SELECT * FROM donor_annual_summary 
     WHERE donor_name = ? AND id_card_last4 = ? AND year = ?`
  ).get(donorName, idCardLast4, year) as DonorAnnualSummary | undefined;
  
  if (existing) {
    return existing;
  }
  
  const id = uuidv4();
  const now = dayjs().format('YYYY-MM-DD HH:mm:ss');
  const limit = Math.floor(taxIncomeBase * 0.3);
  
  db.prepare(
    `INSERT INTO donor_annual_summary 
     (id, donor_name, id_card_last4, year, total_donation_amount, total_tax_deductible_amount, tax_income_base, tax_limit, updated_at)
     VALUES (?, ?, ?, ?, 0, 0, ?, ?, ?)`
  ).run(id, donorName, idCardLast4, year, taxIncomeBase, limit, now);
  
  return db.prepare(
    `SELECT * FROM donor_annual_summary WHERE id = ?`
  ).get(id) as DonorAnnualSummary;
}

router.post('/', (req, res) => {
  const body = req.body as CreateVoucherRequest;
  
  if (!body.donation_id) {
    return res.status(400).json({ error: '缺少必要字段 donation_id' });
  }
  
  const donation = db.prepare(
    'SELECT * FROM donations WHERE id = ?'
  ).get(body.donation_id) as Donation | undefined;
  
  if (!donation) {
    return res.status(404).json({ error: '捐赠记录不存在' });
  }
  
  const project = db.prepare(
    'SELECT * FROM projects WHERE id = ?'
  ).get(donation.project_id) as Project | undefined;
  
  if (!project) {
    return res.status(400).json({ error: '项目不存在' });
  }
  
  if (!project.has_tax_deductible_qualification) {
    return res.status(400).json({ error: '该项目不具备税前扣除资格' });
  }
  
  const existingActiveVoucher = db.prepare(
    `SELECT * FROM tax_deduction_vouchers 
     WHERE donation_id = ? AND status = 'active'`
  ).get(body.donation_id);
  
  if (existingActiveVoucher) {
    return res.status(409).json({ error: '该捐赠已有有效凭证，如需修改请先作废原有凭证' });
  }
  
  const taxYear = dayjs(donation.donated_at).year();
  const taxIncomeBase = body.tax_income_base;
  
  if (taxIncomeBase === undefined || taxIncomeBase === null) {
    return res.status(400).json({ error: '缺少必要字段 tax_income_base（应纳税所得额，单位：分）' });
  }
  
  if (taxIncomeBase < 0) {
    return res.status(400).json({ error: '应纳税所得额不能为负数' });
  }
  
  const transaction = db.transaction(() => {
    const summary = getOrCreateDonorSummary(donation.donor_name, donation.id_card_last4, taxYear, taxIncomeBase);
    
    const effectiveIncomeBase = Number(summary.tax_income_base) || 0;
    const newTaxIncomeBase = Number(taxIncomeBase) || 0;
    
    let taxLimit = Number(summary.tax_limit) || 0;
    if (newTaxIncomeBase > effectiveIncomeBase) {
      taxLimit = Math.floor(newTaxIncomeBase * 0.3);
      db.prepare(
        `UPDATE donor_annual_summary 
         SET tax_income_base = ?, tax_limit = ?, updated_at = ?
         WHERE id = ?`
      ).run(newTaxIncomeBase, taxLimit, dayjs().format('YYYY-MM-DD HH:mm:ss'), summary.id);
    } else if (effectiveIncomeBase > 0) {
      taxLimit = Math.floor(effectiveIncomeBase * 0.3);
    }
    
    const currentTotal = Number(summary.total_tax_deductible_amount) || 0;
    const limit = taxLimit;
    
    let deductibleAmount: number;
    if (currentTotal >= limit) {
      deductibleAmount = 0;
    } else {
      const remainingLimit = limit - currentTotal;
      deductibleAmount = Math.min(donation.amount, remainingLimit);
    }
    
    if (deductibleAmount <= 0) {
      const recordId = uuidv4();
      const now = dayjs().format('YYYY-MM-DD HH:mm:ss');
      db.prepare(
        `INSERT INTO donation_records 
         (id, donation_id, donor_name, id_card_last4, project_id, amount, year, is_tax_deductible, created_at)
         VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?)`
      ).run(recordId, donation.id, donation.donor_name, donation.id_card_last4, donation.project_id, donation.amount, taxYear, now);
      
      return {
        voucher: null,
        message: '已达年度30%限额，超出部分记为普通捐赠',
        isFullDonation: true
      };
    }
    
    let isFullDonation = true;
    let nonDeductibleAmount = 0;
    
    if (deductibleAmount < donation.amount) {
      isFullDonation = false;
      nonDeductibleAmount = donation.amount - deductibleAmount;
    }
    
    const voucherId = uuidv4();
    const voucherNumber = generateVoucherNumber();
    const now = dayjs().format('YYYY-MM-DD HH:mm:ss');
    
    db.prepare(
      `INSERT INTO tax_deduction_vouchers 
       (id, voucher_number, donation_id, donor_name, id_card_last4, project_id, deductible_amount, tax_year, status, issued_at, created_at)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'active', ?, ?)`
    ).run(voucherId, voucherNumber, donation.id, donation.donor_name, donation.id_card_last4, donation.project_id, deductibleAmount, taxYear, now, now);
    
    const newTotalDeductible = currentTotal + deductibleAmount;
    const newTotalDonation = summary.total_donation_amount + donation.amount;
    db.prepare(
      `UPDATE donor_annual_summary 
       SET total_donation_amount = ?, total_tax_deductible_amount = ?, updated_at = ?
       WHERE id = ?`
    ).run(newTotalDonation, newTotalDeductible, now, summary.id);
    
    const recordId1 = uuidv4();
    db.prepare(
      `INSERT INTO donation_records 
       (id, donation_id, donor_name, id_card_last4, project_id, amount, year, is_tax_deductible, created_at)
       VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?)`
    ).run(recordId1, donation.id, donation.donor_name, donation.id_card_last4, donation.project_id, deductibleAmount, taxYear, now);
    
    if (!isFullDonation) {
      const recordId2 = uuidv4();
      db.prepare(
        `INSERT INTO donation_records 
         (id, donation_id, donor_name, id_card_last4, project_id, amount, year, is_tax_deductible, created_at)
         VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?)`
      ).run(recordId2, donation.id, donation.donor_name, donation.id_card_last4, donation.project_id, nonDeductibleAmount, taxYear, now);
    }
    
    const voucher = db.prepare(
      'SELECT * FROM tax_deduction_vouchers WHERE id = ?'
    ).get(voucherId) as TaxDeductionVoucher;
    
    return {
      voucher,
      message: isFullDonation ? '凭证生成成功' : '部分金额已达限额，超出部分记为普通捐赠',
      isFullDonation,
      nonDeductibleAmount: nonDeductibleAmount || 0
    };
  });
  
  try {
    const result = transaction();
    return res.status(201).json(result);
  } catch (err) {
    return res.status(500).json({ error: '生成凭证失败' });
  }
});

router.post('/:id/void', (req, res) => {
  const voucher = db.prepare(
    'SELECT * FROM tax_deduction_vouchers WHERE id = ?'
  ).get(req.params.id) as TaxDeductionVoucher | undefined;
  
  if (!voucher) {
    return res.status(404).json({ error: '凭证不存在' });
  }
  
  if (voucher.status === 'voided') {
    return res.status(400).json({ error: '凭证已作废' });
  }
  
  const transaction = db.transaction(() => {
    const now = dayjs().format('YYYY-MM-DD HH:mm:ss');
    
    db.prepare(
      `UPDATE tax_deduction_vouchers 
       SET status = 'voided', voided_at = ?
       WHERE id = ?`
    ).run(now, voucher.id);
    
    const summary = db.prepare(
      `SELECT * FROM donor_annual_summary 
       WHERE donor_name = ? AND id_card_last4 = ? AND year = ?`
    ).get(voucher.donor_name, voucher.id_card_last4, voucher.tax_year) as DonorAnnualSummary | undefined;
    
    if (summary) {
      const newTotalDeductible = Math.max(0, summary.total_tax_deductible_amount - voucher.deductible_amount);
      const newTotalDonation = Math.max(0, summary.total_donation_amount - voucher.deductible_amount);
      db.prepare(
        `UPDATE donor_annual_summary 
         SET total_donation_amount = ?, total_tax_deductible_amount = ?, updated_at = ?
         WHERE id = ?`
      ).run(newTotalDonation, newTotalDeductible, now, summary.id);
    }
    
    return db.prepare('SELECT * FROM tax_deduction_vouchers WHERE id = ?').get(req.params.id);
  });
  
  try {
    const result = transaction();
    return res.json(result);
  } catch (err) {
    return res.status(500).json({ error: '作废凭证失败' });
  }
});

router.put('/:id', (_req, res) => {
  return res.status(405).json({ error: '凭证不能修改，只能作废后重新开具' });
});

router.patch('/:id', (_req, res) => {
  return res.status(405).json({ error: '凭证不能修改，只能作废后重新开具' });
});

router.get('/:id', (req, res) => {
  const voucher = db.prepare(
    'SELECT * FROM tax_deduction_vouchers WHERE id = ?'
  ).get(req.params.id) as TaxDeductionVoucher | undefined;
  
  if (!voucher) {
    return res.status(404).json({ error: '凭证不存在' });
  }
  
  return res.json(voucher);
});

router.get('/', (_req, res) => {
  const vouchers = db.prepare(
    'SELECT * FROM tax_deduction_vouchers ORDER BY created_at DESC'
  ).all() as TaxDeductionVoucher[];
  
  return res.json(vouchers);
});

export default router;
