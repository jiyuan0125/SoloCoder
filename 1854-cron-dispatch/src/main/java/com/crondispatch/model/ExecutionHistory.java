package com.crondispatch.model;

import com.crondispatch.enums.ExecutionResult;

import java.time.Instant;
import java.util.UUID;

public class ExecutionHistory {
    private String id;
    private String taskId;
    private Instant startTime;
    private Instant endTime;
    private ExecutionResult result;
    private String errorMessage;
    private int retryCount;

    public ExecutionHistory() {
        this.id = UUID.randomUUID().toString().replace("-", "").substring(0, 16);
        this.startTime = Instant.now();
        this.retryCount = 0;
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

    public Instant getStartTime() {
        return startTime;
    }

    public void setStartTime(Instant startTime) {
        this.startTime = startTime;
    }

    public Instant getEndTime() {
        return endTime;
    }

    public void setEndTime(Instant endTime) {
        this.endTime = endTime;
    }

    public ExecutionResult getResult() {
        return result;
    }

    public void setResult(ExecutionResult result) {
        this.result = result;
    }

    public String getErrorMessage() {
        return errorMessage;
    }

    public void setErrorMessage(String errorMessage) {
        this.errorMessage = errorMessage;
    }

    public int getRetryCount() {
        return retryCount;
    }

    public void setRetryCount(int retryCount) {
        this.retryCount = retryCount;
    }
}
