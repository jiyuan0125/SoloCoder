import { db } from '../database';
import { Invoice, InvoiceStatus, InvoiceType } from '../types';

export interface InvoiceRow {
  id: string;
  type: string;
  code: string;
  number: string;
  issue_date: string;
  buyer_name: string;
  buyer_tax_id: string;
  seller_name: string;
  seller_tax_id: string;
  amount: number;
  tax_rate: number;
  tax_amount: number;
  total_amount: number;
  status: string;
  contract_id: string | null;
  original_invoice_id: string | null;
  is_red_invoice: number;
  created_at: string;
}

export function mapInvoiceRow(row: InvoiceRow): Invoice {
  return {
    id: row.id,
    type: row.type as InvoiceType,
    code: row.code,
    number: row.number,
    issueDate: row.issue_date,
    buyer: {
      name: row.buyer_name,
      taxId: row.buyer_tax_id
    },
    seller: {
      name: row.seller_name,
      taxId: row.seller_tax_id
    },
    amount: row.amount,
    taxRate: row.tax_rate,
    taxAmount: row.tax_amount,
    totalAmount: row.total_amount,
    status: row.status as InvoiceStatus,
    contractId: row.contract_id ?? undefined,
    originalInvoiceId: row.original_invoice_id ?? undefined,
    isRedInvoice: row.is_red_invoice === 1,
    createdAt: row.created_at
  };
}

export function findByCodeAndNumber(code: string, number: string): Invoice | undefined {
  const stmt = db.prepare(`
    SELECT * FROM invoices WHERE code = ? AND number = ?
  `);
  const row = stmt.get(code, number) as InvoiceRow | undefined;
  return row ? mapInvoiceRow(row) : undefined;
}

export function findById(id: string): Invoice | undefined {
  const stmt = db.prepare(`
    SELECT * FROM invoices WHERE id = ?
  `);
  const row = stmt.get(id) as InvoiceRow | undefined;
  return row ? mapInvoiceRow(row) : undefined;
}

export function findByOriginalId(originalInvoiceId: string): Invoice | undefined {
  const stmt = db.prepare(`
    SELECT * FROM invoices WHERE original_invoice_id = ?
  `);
  const row = stmt.get(originalInvoiceId) as InvoiceRow | undefined;
  return row ? mapInvoiceRow(row) : undefined;
}

export function insertInvoice(invoice: Invoice): void {
  const stmt = db.prepare(`
    INSERT INTO invoices (
      id, type, code, number, issue_date,
      buyer_name, buyer_tax_id, seller_name, seller_tax_id,
      amount, tax_rate, tax_amount, total_amount, status,
      contract_id, original_invoice_id, is_red_invoice, created_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `);
  stmt.run(
    invoice.id,
    invoice.type,
    invoice.code,
    invoice.number,
    invoice.issueDate,
    invoice.buyer.name,
    invoice.buyer.taxId,
    invoice.seller.name,
    invoice.seller.taxId,
    invoice.amount,
    invoice.taxRate,
    invoice.taxAmount,
    invoice.totalAmount,
    invoice.status,
    invoice.contractId ?? null,
    invoice.originalInvoiceId ?? null,
    invoice.isRedInvoice ? 1 : 0,
    invoice.createdAt
  );
}

export function updateInvoiceStatus(id: string, status: InvoiceStatus): void {
  const stmt = db.prepare(`
    UPDATE invoices SET status = ? WHERE id = ?
  `);
  stmt.run(status, id);
}

export function getAllInvoices(): Invoice[] {
  const stmt = db.prepare(`SELECT * FROM invoices ORDER BY created_at DESC`);
  const rows = stmt.all() as InvoiceRow[];
  return rows.map(mapInvoiceRow);
}
