package com.retail.memberpoints.exception;

import lombok.Getter;

@Getter
public class BusinessException extends RuntimeException {

    private final int code;

    public BusinessException(String message) {
        super(message);
        this.code = 400;
    }

    public BusinessException(int code, String message) {
        super(message);
        this.code = code;
    }

    public static BusinessException memberNotFound() {
        return new BusinessException(404, "会员不存在");
    }

    public static BusinessException memberAlreadyExists() {
        return new BusinessException(409, "会员已存在");
    }

    public static BusinessException insufficientPoints() {
        return new BusinessException(400, "积分余额不足");
    }

    public static BusinessException invalidDeductionAmount() {
        return new BusinessException(400, "抵扣金额不能超过订单金额");
    }

    public static BusinessException orderAlreadyRefunded() {
        return new BusinessException(400, "该订单已退款");
    }
}
