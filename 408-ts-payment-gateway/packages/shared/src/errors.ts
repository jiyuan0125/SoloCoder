export enum ErrorCode {
  SUCCESS = 0,
  INVALID_PARAMS = 10001,
  INVALID_AMOUNT = 10002,
  ORDER_NOT_FOUND = 10003,
  ORDER_STATUS_ERROR = 10004,
  CHANNEL_NOT_SUPPORTED = 10005,
  SIGNATURE_INVALID = 10006,
  CALLBACK_DUPLICATE = 10007,
  REFUND_AMOUNT_EXCEEDED = 10008,
  ORDER_CLOSED = 10009,
  REFUND_NOT_FOUND = 10010,
  INTERNAL_ERROR = 20001,
  CHANNEL_ERROR = 20002,
}

export const ErrorMessages: Record<ErrorCode, string> = {
  [ErrorCode.SUCCESS]: "成功",
  [ErrorCode.INVALID_PARAMS]: "参数错误",
  [ErrorCode.INVALID_AMOUNT]: "金额无效，必须大于0且不超过50000000分(50万)",
  [ErrorCode.ORDER_NOT_FOUND]: "订单不存在",
  [ErrorCode.ORDER_STATUS_ERROR]: "订单状态错误",
  [ErrorCode.CHANNEL_NOT_SUPPORTED]: "不支持的支付渠道",
  [ErrorCode.SIGNATURE_INVALID]: "签名验证失败",
  [ErrorCode.CALLBACK_DUPLICATE]: "回调已处理(幂等)",
  [ErrorCode.REFUND_AMOUNT_EXCEEDED]: "退款金额超过可退款金额",
  [ErrorCode.ORDER_CLOSED]: "订单已关闭，需创建新订单",
  [ErrorCode.REFUND_NOT_FOUND]: "退款记录不存在",
  [ErrorCode.INTERNAL_ERROR]: "内部错误",
  [ErrorCode.CHANNEL_ERROR]: "渠道错误",
};

export class PaymentGatewayError extends Error {
  public readonly code: ErrorCode;
  public readonly message: string;

  constructor(code: ErrorCode, message?: string) {
    const errorMessage = message ?? ErrorMessages[code] ?? "未知错误";
    super(errorMessage);
    this.code = code;
    this.message = errorMessage;
    this.name = "PaymentGatewayError";
  }

  public toJSON(): { code: ErrorCode; message: string } {
    return {
      code: this.code,
      message: this.message,
    };
  }
}
