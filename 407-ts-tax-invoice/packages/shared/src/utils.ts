import { AmountFen } from "./types.js";

export const TAX_NUMBER_REGEX = /^[A-Z0-9]{18}$/;
export const MAX_BATCH_SIZE = 100;

export function isValidTaxNumber(taxNumber: string): boolean {
  return TAX_NUMBER_REGEX.test(taxNumber);
}

export function calculateTaxRate(
  amountExcludingTax: AmountFen,
  taxAmount: AmountFen
): number | null {
  if (amountExcludingTax === 0) {
    return taxAmount === 0 ? 0 : null;
  }
  
  const rate = taxAmount / amountExcludingTax;
  return rate;
}

export function validateAmountRelation(
  amountExcludingTax: AmountFen,
  taxAmount: AmountFen,
  totalAmount: AmountFen
): boolean {
  if (amountExcludingTax < 0 || taxAmount < 0 || totalAmount < 0) {
    return false;
  }

  if (amountExcludingTax + taxAmount !== totalAmount) {
    return false;
  }

  if (amountExcludingTax === 0) {
    return taxAmount === 0;
  }

  const rate = taxAmount / amountExcludingTax;
  const calculatedTaxAmount = Math.round(amountExcludingTax * rate);
  
  return calculatedTaxAmount === taxAmount;
}

export function generateInvoiceId(): string {
  const timestamp = Date.now().toString(36);
  const random = Math.random().toString(36).substring(2, 8);
  return `INV-${timestamp}-${random}`.toUpperCase();
}

export function formatAmountToYuan(amount: AmountFen): string {
  const yuan = Math.floor(amount / 100);
  const fen = amount % 100;
  return `${yuan}.${fen.toString().padStart(2, "0")}`;
}

export function parseYuanToFen(yuanStr: string): AmountFen | null {
  const match = yuanStr.match(/^(-?\d+)(\.(\d{1,2}))?$/);
  if (!match) {
    return null;
  }

  const yuan = parseInt(match[1] || "0", 10);
  const decimal = match[3] || "";
  const fen = decimal.padEnd(2, "0").substring(0, 2);
  
  return yuan * 100 + parseInt(fen, 10);
}

export function isValidDate(dateStr: string): boolean {
  const date = new Date(dateStr);
  return !isNaN(date.getTime());
}

export function getCurrentTimestamp(): string {
  return new Date().toISOString();
}

export function asyncWrapper<T>(
  fn: () => Promise<T>,
  errorHandler: (error: Error) => T
): Promise<T> {
  return fn().catch((error: unknown) => {
    if (error instanceof Error) {
      return errorHandler(error);
    }
    return errorHandler(new Error(String(error)));
  });
}
