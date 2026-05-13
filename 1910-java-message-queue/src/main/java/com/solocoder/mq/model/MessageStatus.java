package com.solocoder.mq.model;

public enum MessageStatus {
    PENDING,
    READY,
    CONSUMING,
    ACKED,
    FAILED,
    DLQ
}
