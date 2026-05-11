package com.delivery.enums;

import lombok.Getter;

@Getter
public enum ExceptionReason {
    WRONG_ADDRESS("客户地址错误"),
    CUSTOMER_RESHEDULE("客户主动改约"),
    BAD_WEATHER("恶劣天气影响"),
    CUSTOMER_REJECT("客户拒收"),
    OTHER("其他原因");

    private final String description;

    ExceptionReason(String description) {
        this.description = description;
    }
}
