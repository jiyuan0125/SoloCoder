package com.example.scheduler.model;

import java.time.Instant;
import java.util.UUID;

public class CallbackLog {
    private String id;
    private String taskId;
    private String executionHistoryId;
    private String requestUrl;
    private String requestBody;
    private String requestHeaders;
    private int responseStatus;
    private String responseBody;
    private String responseHeaders;
    private String errorReason;
    private Instant requestTime;
    private Instant responseTime;

    public CallbackLog() {
        this.id = UUID.randomUUID().toString();
        this.requestTime = Instant.now();
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

    public String getExecutionHistoryId() {
        return executionHistoryId;
    }

    public void setExecutionHistoryId(String executionHistoryId) {
        this.executionHistoryId = executionHistoryId;
    }

    public String getRequestUrl() {
        return requestUrl;
    }

    public void setRequestUrl(String requestUrl) {
        this.requestUrl = requestUrl;
    }

    public String getRequestBody() {
        return requestBody;
    }

    public void setRequestBody(String requestBody) {
        this.requestBody = requestBody;
    }

    public String getRequestHeaders() {
        return requestHeaders;
    }

    public void setRequestHeaders(String requestHeaders) {
        this.requestHeaders = requestHeaders;
    }

    public int getResponseStatus() {
        return responseStatus;
    }

    public void setResponseStatus(int responseStatus) {
        this.responseStatus = responseStatus;
    }

    public String getResponseBody() {
        return responseBody;
    }

    public void setResponseBody(String responseBody) {
        this.responseBody = responseBody;
    }

    public String getResponseHeaders() {
        return responseHeaders;
    }

    public void setResponseHeaders(String responseHeaders) {
        this.responseHeaders = responseHeaders;
    }

    public String getErrorReason() {
        return errorReason;
    }

    public void setErrorReason(String errorReason) {
        this.errorReason = errorReason;
    }

    public Instant getRequestTime() {
        return requestTime;
    }

    public void setRequestTime(Instant requestTime) {
        this.requestTime = requestTime;
    }

    public Instant getResponseTime() {
        return responseTime;
    }

    public void setResponseTime(Instant responseTime) {
        this.responseTime = responseTime;
    }
}
