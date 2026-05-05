import { ErrorCode, ErrorCodes } from './errors';

export const ErrorMessages: Record<ErrorCode, string> = {
  [ErrorCodes.INVALID_INPUT]: '输入参数无效',
  [ErrorCodes.SALESPERSON_NOT_FOUND]: '销售不存在',
  [ErrorCodes.SALESPERSON_ALREADY_EXISTS]: '销售已存在',
  [ErrorCodes.SALESPERSON_ALREADY_RESIGNED]: '销售已离职',
  [ErrorCodes.ORDER_NOT_FOUND]: '订单不存在',
  [ErrorCodes.ORDER_ALREADY_REFUNDED]: '订单已退款',
  [ErrorCodes.ORDER_LOCKED]: '订单已锁定，无法修改',
  [ErrorCodes.SETTLEMENT_NOT_FOUND]: '结算记录不存在',
  [ErrorCodes.INVALID_DATE]: '日期格式无效',
  [ErrorCodes.INVALID_AMOUNT]: '金额无效',
  [ErrorCodes.INVALID_SALES_CONTRIBUTION]: '销售贡献分配无效',
  [ErrorCodes.NO_PRIMARY_SALESPERSON]: '团队订单必须指定主销售',
  [ErrorCodes.MULTIPLE_PRIMARY_SALESPERSONS]: '团队订单只能有一个主销售',
  [ErrorCodes.INSUFFICIENT_COMMISSION_BALANCE]: '佣金余额不足',
  [ErrorCodes.INVALID_RANKING_DIMENSION]: '排名维度无效',
  [ErrorCodes.INVALID_QUARTER]: '季度无效，必须是1-4',
  [ErrorCodes.INVALID_MONTH]: '月份无效，必须是1-12',
  [ErrorCodes.INTERNAL_ERROR]: '内部服务器错误',
};

export function getErrorMessage(code: ErrorCode): string {
  return ErrorMessages[code] || '未知错误';
}

export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: Record<string, unknown>;
  };
}

export function successResponse<T>(data: T): ApiResponse<T> {
  return {
    success: true,
    data,
  };
}

export function errorResponse(
  code: string,
  message: string,
  details?: Record<string, unknown>
): ApiResponse<never> {
  return {
    success: false,
    error: {
      code,
      message,
      details,
    },
  };
}
