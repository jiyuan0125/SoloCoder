import {
  CreateInvoiceRequest,
  ErrorCode,
  InvoiceType,
  BusinessError,
  AmountFen,
  UpdateInvoiceRequest,
} from "@tax-invoice/shared";
import {
  isValidTaxNumber,
  validateAmountRelation,
  isValidDate,
} from "@tax-invoice/shared";
import { store } from "./storage.js";

export interface ValidationResult {
  isValid: boolean;
  errors: string[];
}

export function validateTaxNumber(taxNumber: string, fieldName: string): string[] {
  const errors: string[] = [];
  if (!taxNumber) {
    errors.push(`${fieldName}不能为空`);
    return errors;
  }
  if (!isValidTaxNumber(taxNumber)) {
    errors.push(`税号格式错误，必须是18位字母数字的统一社会信用代码: ${taxNumber}`);
  }
  return errors;
}

export function validateParty(party: { name: string; taxNumber: string }, type: "buyer" | "seller"): string[] {
  const errors: string[] = [];
  const typeLabel = type === "buyer" ? "购买方" : "销售方";

  if (!party.name) {
    errors.push(`${typeLabel}名称不能为空`);
  }
  errors.push(...validateTaxNumber(party.taxNumber, `${typeLabel}税号`));

  return errors;
}

export function validateAmounts(
  amountExcludingTax: AmountFen,
  taxAmount: AmountFen,
  totalAmount: AmountFen
): string[] {
  const errors: string[] = [];

  if (amountExcludingTax < 0) {
    errors.push("金额不含税不能为负数");
  }
  if (taxAmount < 0) {
    errors.push("税额不能为负数");
  }
  if (totalAmount < 0) {
    errors.push("价税合计不能为负数");
  }

  if (!validateAmountRelation(amountExcludingTax, taxAmount, totalAmount)) {
    errors.push(
      `金额关系错误：金额不含税(${amountExcludingTax}) + 税额(${taxAmount}) 必须等于价税合计(${totalAmount})，且金额不含税 × 税率 = 税额`
    );
  }

  return errors;
}

export function validateInvoiceType(invoiceType: unknown): invoiceType is InvoiceType {
  return invoiceType === InvoiceType.SPECIAL || invoiceType === InvoiceType.NORMAL;
}

export function validateCreateInvoiceRequest(request: CreateInvoiceRequest): ValidationResult {
  const errors: string[] = [];

  if (!request.invoiceNumber) {
    errors.push("发票号码不能为空");
  } else {
    const existing = store.getByInvoiceNumber(request.invoiceNumber);
    if (existing) {
      errors.push(`发票号码已存在: ${request.invoiceNumber}`);
    }
  }

  if (!request.invoiceType) {
    errors.push("发票类型不能为空");
  } else if (!validateInvoiceType(request.invoiceType)) {
    errors.push(`无效的发票类型: ${request.invoiceType}`);
  }

  if (!request.buyer) {
    errors.push("购买方信息不能为空");
  } else {
    errors.push(...validateParty(request.buyer, "buyer"));
  }

  if (!request.seller) {
    errors.push("销售方信息不能为空");
  } else {
    errors.push(...validateParty(request.seller, "seller"));
  }

  if (request.amountExcludingTax === undefined || request.amountExcludingTax === null) {
    errors.push("金额不含税不能为空");
  }
  if (request.taxAmount === undefined || request.taxAmount === null) {
    errors.push("税额不能为空");
  }
  if (request.totalAmount === undefined || request.totalAmount === null) {
    errors.push("价税合计不能为空");
  }

  if (
    request.amountExcludingTax !== undefined &&
    request.amountExcludingTax !== null &&
    request.taxAmount !== undefined &&
    request.taxAmount !== null &&
    request.totalAmount !== undefined &&
    request.totalAmount !== null
  ) {
    errors.push(...validateAmounts(request.amountExcludingTax, request.taxAmount, request.totalAmount));
  }

  if (!request.issuedAt) {
    errors.push("开票日期不能为空");
  } else if (!isValidDate(request.issuedAt)) {
    errors.push(`无效的日期格式: ${request.issuedAt}`);
  }

  return {
    isValid: errors.length === 0,
    errors,
  };
}

export function validateUpdateInvoiceRequest(request: UpdateInvoiceRequest): ValidationResult {
  const errors: string[] = [];

  if (!request.id) {
    errors.push("发票ID不能为空");
  }

  return {
    isValid: errors.length === 0,
    errors,
  };
}

export function throwValidationError(errors: string[]): never {
  throw new BusinessError(
    ErrorCode.MISSING_REQUIRED_FIELD,
    "请求参数验证失败",
    errors
  );
}
