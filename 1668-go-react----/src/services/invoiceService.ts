import {
  Invoice,
  InvoiceStatus,
  CreateInvoiceRequest,
  VerifyInvoiceResponse,
  ContractStatus,
  InvoiceType
} from '../types';
import {
  findByCodeAndNumber,
  findById,
  findByOriginalId,
  insertInvoice,
  updateInvoiceStatus,
  getAllInvoices
} from '../repositories/invoiceRepository';
import { findContractById } from '../repositories/contractRepository';
import {
  validateCreateInvoiceRequest,
  validateVerifyRequest,
  roundToCents,
  calculateTaxAmount,
  calculateTotalAmount,
  centsToDisplay
} from '../utils/validation';
import { acquireLock } from '../utils/lock';
import { generateId } from '../utils/id';
import { db } from '../database';

export class InvoiceError extends Error {
  public statusCode: number;

  constructor(message: string, statusCode: number) {
    super(message);
    this.name = 'InvoiceError';
    this.statusCode = statusCode;
  }
}

export function createInvoice(req: CreateInvoiceRequest): Invoice {
  const validationErrors = validateCreateInvoiceRequest(req);
  if (validationErrors.length > 0) {
    const firstError = validationErrors[0];
    throw new InvoiceError(
      firstError.field === 'code' || firstError.field === 'number'
        ? firstError.message
        : firstError.message,
      400
    );
  }

  if (!req.code.trim() || !req.number.trim()) {
    throw new InvoiceError('发票代码或号码不能为空', 400);
  }

  const existing = findByCodeAndNumber(req.code, req.number);
  if (existing) {
    throw new InvoiceError('发票代码和号码已存在', 409);
  }

  if (req.contractId) {
    const contract = findContractById(req.contractId);
    if (!contract) {
      throw new InvoiceError('合同不存在', 404);
    }
    if (contract.status !== ContractStatus.APPROVED) {
      throw new InvoiceError('合同未审批', 400);
    }
  }

  const amountCents = roundToCents(req.amount);
  const taxAmountCents = calculateTaxAmount(amountCents, req.taxRate);
  const totalAmountCents = calculateTotalAmount(amountCents, taxAmountCents);

  const invoice: Invoice = {
    id: generateId(),
    type: req.type,
    code: req.code,
    number: req.number,
    issueDate: req.issueDate,
    buyer: req.buyer,
    seller: req.seller,
    amount: amountCents,
    taxRate: req.taxRate,
    taxAmount: taxAmountCents,
    totalAmount: totalAmountCents,
    status: InvoiceStatus.ISSUED,
    contractId: req.contractId,
    originalInvoiceId: undefined,
    isRedInvoice: false,
    createdAt: new Date().toISOString()
  };

  insertInvoice(invoice);
  return invoice;
}

export async function createRedInvoice(
  originalInvoiceId: string
): Promise<Invoice> {
  if (!originalInvoiceId) {
    throw new InvoiceError('原发票ID不能为空', 400);
  }

  const original = findById(originalInvoiceId);
  if (!original) {
    throw new InvoiceError('原发票不存在', 404);
  }

  if (original.isRedInvoice) {
    throw new InvoiceError('红字发票不能再次红冲', 400);
  }

  const existingRed = findByOriginalId(originalInvoiceId);
  if (existingRed || original.status === InvoiceStatus.RED_INVOICED) {
    throw new InvoiceError('该发票已被红冲', 409);
  }

  if (original.status !== InvoiceStatus.ISSUED) {
    throw new InvoiceError('只有已开具的发票才能红冲', 400);
  }

  let releaseLock: (() => void) | null = null;

  try {
    releaseLock = await acquireLock(`invoice:${originalInvoiceId}`, 5000);
  } catch (error) {
    throw new InvoiceError(
      (error as Error).message || '原发票正在被操作请稍后重试',
      409
    );
  }

  try {
    const freshOriginal = findById(originalInvoiceId);
    if (!freshOriginal || freshOriginal.status !== InvoiceStatus.ISSUED) {
      releaseLock();
      throw new InvoiceError('该发票已被红冲或状态已变更', 409);
    }

    const stillRed = findByOriginalId(originalInvoiceId);
    if (stillRed) {
      releaseLock();
      throw new InvoiceError('该发票已被红冲', 409);
    }

    const redCode = `RED-${freshOriginal.code}`;
    const redNumber = `RED-${freshOriginal.number}`;

    const existingRedByCode = findByCodeAndNumber(redCode, redNumber);
    if (existingRedByCode) {
      releaseLock();
      throw new InvoiceError('红字发票代码和号码已存在', 409);
    }

    const redAmount = -freshOriginal.amount;
    const redTaxAmount = -freshOriginal.taxAmount;
    const redTotalAmount = -freshOriginal.totalAmount;

    const redInvoice: Invoice = {
      id: generateId(),
      type: freshOriginal.type,
      code: redCode,
      number: redNumber,
      issueDate: new Date().toISOString().split('T')[0],
      buyer: freshOriginal.buyer,
      seller: freshOriginal.seller,
      amount: redAmount,
      taxRate: freshOriginal.taxRate,
      taxAmount: redTaxAmount,
      totalAmount: redTotalAmount,
      status: InvoiceStatus.ISSUED,
      contractId: freshOriginal.contractId,
      originalInvoiceId: freshOriginal.id,
      isRedInvoice: true,
      createdAt: new Date().toISOString()
    };

    const transaction = db.transaction(() => {
      insertInvoice(redInvoice);
      updateInvoiceStatus(freshOriginal.id, InvoiceStatus.RED_INVOICED);
    });

    transaction();
    releaseLock();

    return redInvoice;
  } catch (error) {
    if (releaseLock) {
      releaseLock();
    }
    throw error;
  }
}

export function verifyInvoice(
  code: string,
  number: string,
  amount: number,
  taxAmount: number
): VerifyInvoiceResponse {
  const validationErrors = validateVerifyRequest(code, number, amount, taxAmount);
  if (validationErrors.length > 0) {
    throw new InvoiceError(validationErrors[0].message, 400);
  }

  const invoice = findByCodeAndNumber(code, number);
  if (!invoice) {
    return {
      valid: false,
      message: '发票不存在'
    };
  }

  const requestAmountCents = roundToCents(amount);
  const requestTaxAmountCents = roundToCents(taxAmount);

  const amountDiff = requestAmountCents - invoice.amount;
  const taxDiff = requestTaxAmountCents - invoice.taxAmount;

  if (amountDiff !== 0 || taxDiff !== 0) {
    const messages: string[] = [];
    if (amountDiff !== 0) {
      messages.push(
        `金额不匹配：期望 ${centsToDisplay(invoice.amount)}，实际 ${amount}`
      );
    }
    if (taxDiff !== 0) {
      messages.push(
        `税额不匹配：期望 ${centsToDisplay(invoice.taxAmount)}，实际 ${taxAmount}`
      );
    }
    return {
      valid: false,
      message: messages.join('；'),
      invoice
    };
  }

  return {
    valid: true,
    message: '发票信息匹配',
    invoice
  };
}

export function getInvoiceById(id: string): Invoice | undefined {
  return findById(id);
}

export function getInvoiceByCodeAndNumber(
  code: string,
  number: string
): Invoice | undefined {
  return findByCodeAndNumber(code, number);
}

export function listAllInvoices(): Invoice[] {
  return getAllInvoices();
}

export function toDisplayInvoice(invoice: Invoice): any {
  return {
    ...invoice,
    amount: centsToDisplay(invoice.amount),
    taxAmount: centsToDisplay(invoice.taxAmount),
    totalAmount: centsToDisplay(invoice.totalAmount)
  };
}

export function toDisplayInvoices(invoices: Invoice[]): any[] {
  return invoices.map(toDisplayInvoice);
}
