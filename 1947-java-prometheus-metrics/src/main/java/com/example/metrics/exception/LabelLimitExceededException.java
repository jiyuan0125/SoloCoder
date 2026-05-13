package com.example.metrics.exception;

public class LabelLimitExceededException extends RuntimeException {
    public LabelLimitExceededException(String message) {
        super(message);
    }
}
