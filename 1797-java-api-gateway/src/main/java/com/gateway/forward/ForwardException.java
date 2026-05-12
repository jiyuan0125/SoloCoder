package com.gateway.forward;

import lombok.Getter;

@Getter
public class ForwardException extends RuntimeException {

    private final String errorCode;
    private final int httpStatus;
    private final String backendUrl;

    public ForwardException(String errorCode, int httpStatus, String message, String backendUrl) {
        super(message);
        this.errorCode = errorCode;
        this.httpStatus = httpStatus;
        this.backendUrl = backendUrl;
    }

    public ForwardException(String errorCode, int httpStatus, String message, String backendUrl, Throwable cause) {
        super(message, cause);
        this.errorCode = errorCode;
        this.httpStatus = httpStatus;
        this.backendUrl = backendUrl;
    }
}
