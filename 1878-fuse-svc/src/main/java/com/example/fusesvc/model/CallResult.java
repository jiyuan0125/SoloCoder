package com.example.fusesvc.model;

import java.time.LocalDateTime;

public class CallResult {
    private LocalDateTime timestamp;
    private boolean success;

    public CallResult(LocalDateTime timestamp, boolean success) {
        this.timestamp = timestamp;
        this.success = success;
    }

    public LocalDateTime getTimestamp() {
        return timestamp;
    }

    public void setTimestamp(LocalDateTime timestamp) {
        this.timestamp = timestamp;
    }

    public boolean isSuccess() {
        return success;
    }

    public void setSuccess(boolean success) {
        this.success = success;
    }
}
