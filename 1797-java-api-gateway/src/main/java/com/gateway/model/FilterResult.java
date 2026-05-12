package com.gateway.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class FilterResult {

    public enum Status {
        PASS,
        REJECT
    }

    private Status status;
    private String errorCode;
    private String errorMessage;
    private int httpStatus;

    public static FilterResult pass() {
        return new FilterResult(Status.PASS, null, null, 0);
    }

    public static FilterResult reject(String errorCode, String errorMessage, int httpStatus) {
        return new FilterResult(Status.REJECT, errorCode, errorMessage, httpStatus);
    }
}
