package com.example.scheduler.model;

import java.time.Instant;
import java.util.UUID;

public class ExecutionHistory {
    private String id;
    private String taskId;
    private ExecutionStatus status;
    private Instant executionTime;
    private String errorReason;
    private int retryCount;
    private String callbackLogId;

    public ExecutionHistory() {
        this.id = UUID.randomUUID().toString();
        this.executionTime = Instant.now();
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getTaskId() {
        return taskId;
    }

    public void setTaskId(String taskId) {
        this.taskId = taskId;
    }

    public ExecutionStatus getStatus() {
        return status;
    }

    public void setStatus(ExecutionStatus status) {
        this.status = status;
    }

    public Instant getExecutionTime() {
        return executionTime;
    }

    public void setExecutionTime(Instant executionTime) {
        this.executionTime = executionTime;
    }

    public String getErrorReason() {
        return errorReason;
    }

    public void setErrorReason(String errorReason) {
        this.errorReason = errorReason;
    }

    public int getRetryCount() {
        return retryCount;
    }

    public void setRetryCount(int retryCount) {
        this.retryCount = retryCount;
    }

    public String getCallbackLogId() {
        return callbackLogId;
    }

    public void setCallbackLogId(String callbackLogId) {
        this.callbackLogId = callbackLogId;
    }
}
