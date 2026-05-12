import { InvoiceType, CreateInvoiceRequest } from '../types';

const VALID_TAX_RATES = [0.13, 0.09, 0.06, 0.03, 0.01];
const VALID_INVOICE_TYPES = Object.values(InvoiceType);

export interface ValidationError {
  field: string;
  message: string;
}

export function roundToCents(value: number): number {
  return Math.round(value * 100);
}

export function centsToDisplay(cents: number): number {
  return cents / 100;
}

export function isValidTaxRate(rate: number): boolean {
  return VALID_TAX_RATES.includes(rate);
}

export function calculateTaxAmount(amount: number, taxRate: number): number {
  const rawTax = amount * taxRate;
  return roundToCents(rawTax / 100);
}

export function calculateTotalAmount(amount: number, taxAmount: number): number {
  return amount + taxAmount;
}

export function validateCreateInvoiceRequest(req: CreateInvoiceRequest): ValidationError[] {
  const errors: ValidationError[] = [];

  if (!req.code || req.code.trim() === '') {
    errors.push({ field: 'code', message: '发票代码不能为空' });
  }

  if (!req.number || req.number.trim() === '') {
    errors.push({ field: 'number', message: '发票号码不能为空' });
  }

  if (!VALID_INVOICE_TYPES.includes(req.type)) {
    errors.push({ field: 'type', message: '发票类型无效' });
  }

  if (!req.issueDate) {
    errors.push({ field: 'issueDate', message: '开票日期不能为空' });
  }

  if (!req.buyer || !req.buyer.name || !req.buyer.taxId) {
    errors.push({ field: 'buyer', message: '购买方信息不完整' });
  }

  if (!req.seller || !req.seller.name || !req.seller.taxId) {
    errors.push({ field: 'seller', message: '销售方信息不完整' });
  }

  if (typeof req.amount !== 'number' || isNaN(req.amount)) {
    errors.push({ field: 'amount', message: '金额无效' });
  }

  if (!isValidTaxRate(req.taxRate)) {
    errors.push({ field: 'taxRate', message: '税率无效，支持13%、9%、6%、3%、1%' });
  }

  return errors;
}

export function validateVerifyRequest(
  code: string,
  number: string,
  amount: number,
  taxAmount: number
): ValidationError[] {
  const errors: ValidationError[] = [];

  if (!code || code.trim() === '') {
    errors.push({ field: 'code', message: '发票代码不能为空' });
  }

  if (!number || number.trim() === '') {
    errors.push({ field: 'number', message: '发票号码不能为空' });
  }

  if (typeof amount !== 'number' || isNaN(amount)) {
    errors.push({ field: 'amount', message: '金额无效' });
  }

  if (typeof taxAmount !== 'number' || isNaN(taxAmount)) {
    errors.push({ field: 'taxAmount', message: '税额无效' });
  }

  return errors;
}
