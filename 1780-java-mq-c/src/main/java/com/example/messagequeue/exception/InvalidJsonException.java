package com.example.messagequeue.exception;

import lombok.Getter;

@Getter
public class InvalidJsonException extends RuntimeException {

    private final int errorLine;
    private final int errorColumn;
    private final String errorMessage;

    public InvalidJsonException(int errorLine, int errorColumn, String errorMessage) {
        super(String.format("Invalid JSON at line %d, column %d: %s", errorLine, errorColumn, errorMessage));
        this.errorLine = errorLine;
        this.errorColumn = errorColumn;
        this.errorMessage = errorMessage;
    }

    public InvalidJsonException(String errorMessage) {
        super(errorMessage);
        this.errorLine = -1;
        this.errorColumn = -1;
        this.errorMessage = errorMessage;
    }
}
