package com.breaker.model;

import java.time.LocalDateTime;

public class CallRecord {
    private boolean success;
    private LocalDateTime timestamp;

    public CallRecord(boolean success, LocalDateTime timestamp) {
        this.success = success;
        this.timestamp = timestamp;
    }

    public boolean isSuccess() {
        return success;
    }

    public void setSuccess(boolean success) {
        this.success = success;
    }

    public LocalDateTime getTimestamp() {
        return timestamp;
    }

    public void setTimestamp(LocalDateTime timestamp) {
        this.timestamp = timestamp;
    }
}
