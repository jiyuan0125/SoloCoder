export type InvoiceId = string;
export type TaxNumber = string;
export type AmountFen = number;

export enum InvoiceType {
  SPECIAL = "special",
  NORMAL = "normal",
}

export enum InvoiceStatus {
  NORMAL = "normal",
  VOIDED = "voided",
  RED_INVOICED = "red_invoiced",
}

export interface Party {
  name: string;
  taxNumber: TaxNumber;
}

export interface Invoice {
  id: InvoiceId;
  invoiceNumber: string;
  invoiceType: InvoiceType;
  buyer: Party;
  seller: Party;
  amountExcludingTax: AmountFen;
  taxAmount: AmountFen;
  totalAmount: AmountFen;
  status: InvoiceStatus;
  isDeducted: boolean;
  reimbursementId?: string;
  redInvoiceId?: InvoiceId;
  originalInvoiceId?: InvoiceId;
  issuedAt: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateInvoiceRequest {
  invoiceNumber: string;
  invoiceType: InvoiceType;
  buyer: Party;
  seller: Party;
  amountExcludingTax: AmountFen;
  taxAmount: AmountFen;
  totalAmount: AmountFen;
  isDeducted: boolean;
  issuedAt: string;
}

export interface UpdateInvoiceRequest {
  id: InvoiceId;
  reimbursementId?: string;
  isDeducted?: boolean;
}

export interface InvoiceFilter {
  taxAmountMin?: AmountFen;
  taxAmountMax?: AmountFen;
  invoiceType?: InvoiceType;
  status?: InvoiceStatus;
  startDate?: string;
  endDate?: string;
}

export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: unknown;
  };
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
}

export interface BatchImportResult {
  success: number;
  failed: number;
  successItems: Invoice[];
  failedItems: Array<{
    index: number;
    data: CreateInvoiceRequest;
    reason: string;
  }>;
}

export interface MonthlySummary {
  year: number;
  month: number;
  specialInvoice: {
    count: number;
    totalAmountExcludingTax: AmountFen;
    totalTaxAmount: AmountFen;
    totalAmount: AmountFen;
  };
  normalInvoice: {
    count: number;
    totalAmountExcludingTax: AmountFen;
    totalTaxAmount: AmountFen;
    totalAmount: AmountFen;
  };
}
