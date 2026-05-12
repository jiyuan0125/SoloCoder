package com.example.messagequeue.exception;

import lombok.Getter;

@Getter
public class MessageTooLargeException extends RuntimeException {

    private final int actualSize;
    private final int maxSize;

    public MessageTooLargeException(int actualSize, int maxSize) {
        super(String.format("Message size %d bytes exceeds maximum allowed %d bytes", actualSize, maxSize));
        this.actualSize = actualSize;
        this.maxSize = maxSize;
    }
}
