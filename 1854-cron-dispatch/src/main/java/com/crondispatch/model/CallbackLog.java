package com.crondispatch.model;

import java.time.Instant;
import java.util.UUID;

public class CallbackLog {
    private String id;
    private String executionHistoryId;
    private String taskId;
    private Instant timestamp;
    private String requestBody;
    private String responseBody;
    private int responseStatus;
    private String errorType;

    public CallbackLog() {
        this.id = UUID.randomUUID().toString().replace("-", "").substring(0, 16);
        this.timestamp = Instant.now();
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getExecutionHistoryId() {
        return executionHistoryId;
    }

    public void setExecutionHistoryId(String executionHistoryId) {
        this.executionHistoryId = executionHistoryId;
    }

    public String getTaskId() {
        return taskId;
    }

    public void setTaskId(String taskId) {
        this.taskId = taskId;
    }

    public Instant getTimestamp() {
        return timestamp;
    }

    public void setTimestamp(Instant timestamp) {
        this.timestamp = timestamp;
    }

    public String getRequestBody() {
        return requestBody;
    }

    public void setRequestBody(String requestBody) {
        this.requestBody = requestBody;
    }

    public String getResponseBody() {
        return responseBody;
    }

    public void setResponseBody(String responseBody) {
        this.responseBody = responseBody;
    }

    public int getResponseStatus() {
        return responseStatus;
    }

    public void setResponseStatus(int responseStatus) {
        this.responseStatus = responseStatus;
    }

    public String getErrorType() {
        return errorType;
    }

    public void setErrorType(String errorType) {
        this.errorType = errorType;
    }
}
