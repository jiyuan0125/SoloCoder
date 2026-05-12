package com.example.protobridge.protocol;

import lombok.Getter;
import org.springframework.http.HttpStatus;

@Getter
public class ProtocolException extends RuntimeException {

    private final HttpStatus status;

    public ProtocolException(String message) {
        super(message);
        this.status = HttpStatus.BAD_REQUEST;
    }

    public ProtocolException(String message, HttpStatus status) {
        super(message);
        this.status = status;
    }

    public ProtocolException(String message, Throwable cause) {
        super(message, cause);
        this.status = HttpStatus.BAD_REQUEST;
    }
}
