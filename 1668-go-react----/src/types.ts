export enum InvoiceType {
  SPECIAL = 'special',
  NORMAL = 'normal',
  ELECTRONIC = 'electronic'
}

export enum InvoiceStatus {
  DRAFT = 'draft',
  ISSUED = 'issued',
  RED_INVOICED = 'red_invoiced',
  VOIDED = 'voided'
}

export enum ContractStatus {
  PENDING = 'pending',
  APPROVED = 'approved',
  REJECTED = 'rejected'
}

export interface PartyInfo {
  name: string;
  taxId: string;
}

export interface Contract {
  id: string;
  code: string;
  name: string;
  status: ContractStatus;
  createdAt: string;
}

export interface Invoice {
  id: string;
  type: InvoiceType;
  code: string;
  number: string;
  issueDate: string;
  buyer: PartyInfo;
  seller: PartyInfo;
  amount: number;
  taxRate: number;
  taxAmount: number;
  totalAmount: number;
  status: InvoiceStatus;
  contractId?: string;
  originalInvoiceId?: string;
  isRedInvoice: boolean;
  createdAt: string;
}

export interface CreateInvoiceRequest {
  type: InvoiceType;
  code: string;
  number: string;
  issueDate: string;
  buyer: PartyInfo;
  seller: PartyInfo;
  amount: number;
  taxRate: number;
  contractId?: string;
}

export interface RedInvoiceRequest {
  originalInvoiceId: string;
}

export interface VerifyInvoiceRequest {
  code: string;
  number: string;
  amount: number;
  taxAmount: number;
}

export interface VerifyInvoiceResponse {
  valid: boolean;
  message?: string;
  invoice?: Invoice;
}
