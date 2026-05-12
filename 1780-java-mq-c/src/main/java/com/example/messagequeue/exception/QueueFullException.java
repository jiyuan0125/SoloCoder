package com.example.messagequeue.exception;

import lombok.Getter;

@Getter
public class QueueFullException extends RuntimeException {

    private final String topic;
    private final int currentSize;
    private final int maxSize;
    private final int waitTimeoutSeconds;

    public QueueFullException(String topic, int currentSize, int maxSize, int waitTimeoutSeconds) {
        super(String.format("Topic '%s' queue is full (current: %d, max: %d). Waited %d seconds but no space available.",
                topic, currentSize, maxSize, waitTimeoutSeconds));
        this.topic = topic;
        this.currentSize = currentSize;
        this.maxSize = maxSize;
        this.waitTimeoutSeconds = waitTimeoutSeconds;
    }
}
