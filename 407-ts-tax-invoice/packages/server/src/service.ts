import {
  Invoice,
  InvoiceId,
  InvoiceStatus,
  CreateInvoiceRequest,
  UpdateInvoiceRequest,
  InvoiceFilter,
  BatchImportResult,
  MonthlySummary,
  ErrorCode,
  BusinessError,
  createError,
  AmountFen,
  InvoiceType,
} from "@tax-invoice/shared";
import {
  generateInvoiceId,
  getCurrentTimestamp,
  MAX_BATCH_SIZE,
} from "@tax-invoice/shared";
import { store } from "./storage.js";
import {
  validateCreateInvoiceRequest,
  validateUpdateInvoiceRequest,
  throwValidationError,
} from "./validator.js";

export class InvoiceService {
  createInvoice(request: CreateInvoiceRequest): Invoice {
    const validation = validateCreateInvoiceRequest(request);
    if (!validation.isValid) {
      throwValidationError(validation.errors);
    }

    const now = getCurrentTimestamp();
    const invoice: Invoice = {
      id: generateInvoiceId(),
      invoiceNumber: request.invoiceNumber,
      invoiceType: request.invoiceType,
      buyer: {
        name: request.buyer.name,
        taxNumber: request.buyer.taxNumber,
      },
      seller: {
        name: request.seller.name,
        taxNumber: request.seller.taxNumber,
      },
      amountExcludingTax: request.amountExcludingTax,
      taxAmount: request.taxAmount,
      totalAmount: request.totalAmount,
      status: InvoiceStatus.NORMAL,
      isDeducted: request.isDeducted,
      issuedAt: request.issuedAt,
      createdAt: now,
      updatedAt: now,
    };

    store.save(invoice);
    return invoice;
  }

  getInvoiceById(id: InvoiceId): Invoice {
    const invoice = store.getById(id);
    if (!invoice) {
      throw createError(ErrorCode.INVOICE_NOT_FOUND);
    }
    return invoice;
  }

  updateInvoice(request: UpdateInvoiceRequest): Invoice {
    const validation = validateUpdateInvoiceRequest(request);
    if (!validation.isValid) {
      throwValidationError(validation.errors);
    }

    const invoice = this.getInvoiceById(request.id);

    if (request.reimbursementId !== undefined) {
      invoice.reimbursementId = request.reimbursementId || undefined;
    }

    if (request.isDeducted !== undefined) {
      invoice.isDeducted = request.isDeducted;
    }

    invoice.updatedAt = getCurrentTimestamp();
    store.update(invoice);

    return invoice;
  }

  voidInvoice(id: InvoiceId): Invoice {
    const invoice = this.getInvoiceById(id);

    if (invoice.status !== InvoiceStatus.NORMAL) {
      if (invoice.status === InvoiceStatus.VOIDED) {
        throw createError(ErrorCode.INVOICE_ALREADY_VOIDED);
      }
      if (invoice.status === InvoiceStatus.RED_INVOICED) {
        throw createError(ErrorCode.INVOICE_ALREADY_RED_INVOICED);
      }
    }

    if (invoice.reimbursementId) {
      throw createError(ErrorCode.INVOICE_ASSOCIATED_WITH_REIMBURSEMENT);
    }

    if (invoice.isDeducted) {
      throw createError(ErrorCode.INVOICE_DEDUCTED_CANNOT_VOID);
    }

    invoice.status = InvoiceStatus.VOIDED;
    invoice.updatedAt = getCurrentTimestamp();
    store.update(invoice);

    return invoice;
  }

  redInvoice(originalId: InvoiceId, redInvoiceData: {
    invoiceNumber: string;
    buyer: { name: string; taxNumber: string };
    seller: { name: string; taxNumber: string };
    issuedAt: string;
  }): { originalInvoice: Invoice; redInvoice: Invoice } {
    const originalInvoice = this.getInvoiceById(originalId);

    if (originalInvoice.status !== InvoiceStatus.NORMAL) {
      if (originalInvoice.status === InvoiceStatus.VOIDED) {
        throw createError(ErrorCode.INVOICE_ALREADY_VOIDED, "已作废发票不能红冲");
      }
      if (originalInvoice.status === InvoiceStatus.RED_INVOICED) {
        throw createError(ErrorCode.INVOICE_ALREADY_RED_INVOICED);
      }
    }

    if (!originalInvoice.isDeducted) {
      throw createError(ErrorCode.INVOICE_NOT_DEDUCTED_CANNOT_RED);
    }

    const now = getCurrentTimestamp();
    const redInvoice: Invoice = {
      id: generateInvoiceId(),
      invoiceNumber: redInvoiceData.invoiceNumber,
      invoiceType: originalInvoice.invoiceType,
      buyer: redInvoiceData.buyer,
      seller: redInvoiceData.seller,
      amountExcludingTax: -originalInvoice.amountExcludingTax,
      taxAmount: -originalInvoice.taxAmount,
      totalAmount: -originalInvoice.totalAmount,
      status: InvoiceStatus.RED_INVOICED,
      isDeducted: true,
      originalInvoiceId: originalInvoice.id,
      issuedAt: redInvoiceData.issuedAt,
      createdAt: now,
      updatedAt: now,
    };

    originalInvoice.status = InvoiceStatus.RED_INVOICED;
    originalInvoice.redInvoiceId = redInvoice.id;
    originalInvoice.updatedAt = now;

    store.save(redInvoice);
    store.update(originalInvoice);

    return {
      originalInvoice,
      redInvoice,
    };
  }

  queryInvoices(filter: InvoiceFilter): Invoice[] {
    let invoices = store.getAll();

    const { invoiceType, status, taxAmountMin, taxAmountMax, startDate, endDate } = filter;

    if (invoiceType) {
      invoices = invoices.filter((inv) => inv.invoiceType === invoiceType);
    }

    if (status) {
      invoices = invoices.filter((inv) => inv.status === status);
    }

    if (taxAmountMin !== undefined) {
      const min = taxAmountMin;
      invoices = invoices.filter((inv) => inv.taxAmount >= min);
    }

    if (taxAmountMax !== undefined) {
      const max = taxAmountMax;
      invoices = invoices.filter((inv) => inv.taxAmount <= max);
    }

    if (startDate) {
      const start = startDate;
      invoices = invoices.filter((inv) => inv.issuedAt >= start);
    }

    if (endDate) {
      const end = endDate;
      invoices = invoices.filter((inv) => inv.issuedAt <= end);
    }

    invoices.sort((a, b) => new Date(b.issuedAt).getTime() - new Date(a.issuedAt).getTime());

    return invoices;
  }

  batchImport(invoiceList: CreateInvoiceRequest[]): BatchImportResult {
    if (invoiceList.length > MAX_BATCH_SIZE) {
      throw createError(ErrorCode.BATCH_IMPORT_TOO_MANY);
    }

    const successItems: Invoice[] = [];
    const failedItems: BatchImportResult["failedItems"] = [];

    for (let i = 0; i < invoiceList.length; i++) {
      const request = invoiceList[i];
      if (!request) {
        failedItems.push({
          index: i,
          data: {} as CreateInvoiceRequest,
          reason: "无效的发票数据",
        });
        continue;
      }

      try {
        const validation = validateCreateInvoiceRequest(request);
        if (!validation.isValid) {
          failedItems.push({
            index: i,
            data: request,
            reason: validation.errors.join("; "),
          });
          continue;
        }

        const invoice = this.createInvoice(request);
        successItems.push(invoice);
      } catch (error) {
        let reason = "未知错误";
        if (error instanceof BusinessError) {
          reason = error.message;
        } else if (error instanceof Error) {
          reason = error.message;
        }

        failedItems.push({
          index: i,
          data: request,
          reason,
        });
      }
    }

    return {
      success: successItems.length,
      failed: failedItems.length,
      successItems,
      failedItems,
    };
  }

  getMonthlySummary(year: number, month: number): MonthlySummary {
    const startDate = `${year}-${month.toString().padStart(2, "0")}-01T00:00:00.000Z`;
    const nextMonth = month === 12 ? 1 : month + 1;
    const nextYear = month === 12 ? year + 1 : year;
    const endDate = `${nextYear}-${nextMonth.toString().padStart(2, "0")}-01T00:00:00.000Z`;

    const invoices = store.getAll().filter((inv) => {
      return inv.issuedAt >= startDate && inv.issuedAt < endDate && inv.status === InvoiceStatus.NORMAL;
    });

    const specialInvoiceStats = {
      count: 0,
      totalAmountExcludingTax: 0 as AmountFen,
      totalTaxAmount: 0 as AmountFen,
      totalAmount: 0 as AmountFen,
    };

    const normalInvoiceStats = {
      count: 0,
      totalAmountExcludingTax: 0 as AmountFen,
      totalTaxAmount: 0 as AmountFen,
      totalAmount: 0 as AmountFen,
    };

    for (const invoice of invoices) {
      if (invoice.invoiceType === InvoiceType.SPECIAL) {
        specialInvoiceStats.count++;
        specialInvoiceStats.totalAmountExcludingTax += invoice.amountExcludingTax;
        specialInvoiceStats.totalTaxAmount += invoice.taxAmount;
        specialInvoiceStats.totalAmount += invoice.totalAmount;
      } else {
        normalInvoiceStats.count++;
        normalInvoiceStats.totalAmountExcludingTax += invoice.amountExcludingTax;
        normalInvoiceStats.totalTaxAmount += invoice.taxAmount;
        normalInvoiceStats.totalAmount += invoice.totalAmount;
      }
    }

    return {
      year,
      month,
      specialInvoice: specialInvoiceStats,
      normalInvoice: normalInvoiceStats,
    };
  }
}

export const invoiceService = new InvoiceService();
