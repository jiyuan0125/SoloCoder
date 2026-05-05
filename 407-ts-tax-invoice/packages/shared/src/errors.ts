export enum ErrorCode {
  INVALID_TAX_NUMBER = "INVALID_TAX_NUMBER",
  INVALID_AMOUNT_RELATION = "INVALID_AMOUNT_RELATION",
  INVOICE_NOT_FOUND = "INVOICE_NOT_FOUND",
  INVOICE_ALREADY_VOIDED = "INVOICE_ALREADY_VOIDED",
  INVOICE_ALREADY_RED_INVOICED = "INVOICE_ALREADY_RED_INVOICED",
  INVOICE_ASSOCIATED_WITH_REIMBURSEMENT = "INVOICE_ASSOCIATED_WITH_REIMBURSEMENT",
  INVOICE_DEDUCTED_CANNOT_VOID = "INVOICE_DEDUCTED_CANNOT_VOID",
  INVOICE_NOT_DEDUCTED_CANNOT_RED = "INVOICE_NOT_DEDUCTED_CANNOT_RED",
  RED_INVOICE_AMOUNT_MISMATCH = "RED_INVOICE_AMOUNT_MISMATCH",
  BATCH_IMPORT_TOO_MANY = "BATCH_IMPORT_TOO_MANY",
  INVALID_INVOICE_TYPE = "INVALID_INVOICE_TYPE",
  INVALID_STATUS = "INVALID_STATUS",
  INVALID_DATE = "INVALID_DATE",
  MISSING_REQUIRED_FIELD = "MISSING_REQUIRED_FIELD",
  INVALID_OPERATION = "INVALID_OPERATION",
  INTERNAL_ERROR = "INTERNAL_ERROR",
}

export class BusinessError extends Error {
  public readonly code: ErrorCode;
  public readonly details?: unknown;

  constructor(code: ErrorCode, message: string, details?: unknown) {
    super(message);
    this.code = code;
    this.details = details;
    this.name = "BusinessError";
  }
}

export const ERROR_MESSAGES: Record<ErrorCode, string> = {
  [ErrorCode.INVALID_TAX_NUMBER]: "税号格式错误，必须是18位字母数字的统一社会信用代码",
  [ErrorCode.INVALID_AMOUNT_RELATION]: "金额关系错误：金额不含税 × 税率 = 税额，且金额不含税 + 税额 = 价税合计",
  [ErrorCode.INVOICE_NOT_FOUND]: "发票不存在",
  [ErrorCode.INVOICE_ALREADY_VOIDED]: "发票已作废",
  [ErrorCode.INVOICE_ALREADY_RED_INVOICED]: "发票已红冲",
  [ErrorCode.INVOICE_ASSOCIATED_WITH_REIMBURSEMENT]: "发票已关联报销单，不能直接作废，请先解除关联",
  [ErrorCode.INVOICE_DEDUCTED_CANNOT_VOID]: "发票已抵扣，不能作废，请使用红冲操作",
  [ErrorCode.INVOICE_NOT_DEDUCTED_CANNOT_RED]: "发票未抵扣，不能红冲，请使用作废操作",
  [ErrorCode.RED_INVOICE_AMOUNT_MISMATCH]: "红字发票金额或税额与原发票不一致",
  [ErrorCode.BATCH_IMPORT_TOO_MANY]: "批量导入数量超过限制，单次最多100张",
  [ErrorCode.INVALID_INVOICE_TYPE]: "无效的发票类型",
  [ErrorCode.INVALID_STATUS]: "无效的发票状态",
  [ErrorCode.INVALID_DATE]: "无效的日期格式",
  [ErrorCode.MISSING_REQUIRED_FIELD]: "缺少必需字段",
  [ErrorCode.INVALID_OPERATION]: "无效的操作",
  [ErrorCode.INTERNAL_ERROR]: "内部服务器错误",
};

export function createError(code: ErrorCode, message?: string, details?: unknown): BusinessError {
  return new BusinessError(code, message || ERROR_MESSAGES[code], details);
}
