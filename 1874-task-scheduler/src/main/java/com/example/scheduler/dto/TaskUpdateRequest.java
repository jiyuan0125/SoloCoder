package com.example.scheduler.dto;

import com.example.scheduler.model.ConcurrencyPolicy;
import com.example.scheduler.model.TaskType;

import java.util.Map;

public class TaskUpdateRequest {
    private String name;
    private TaskType type;
    private String expression;
    private Long delaySeconds;
    private String callbackUrl;
    private Map<String, Object> parameters;
    private Boolean allowRetry;
    private ConcurrencyPolicy concurrencyPolicy;

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public TaskType getType() {
        return type;
    }

    public void setType(TaskType type) {
        this.type = type;
    }

    public String getExpression() {
        return expression;
    }

    public void setExpression(String expression) {
        this.expression = expression;
    }

    public Long getDelaySeconds() {
        return delaySeconds;
    }

    public void setDelaySeconds(Long delaySeconds) {
        this.delaySeconds = delaySeconds;
    }

    public String getCallbackUrl() {
        return callbackUrl;
    }

    public void setCallbackUrl(String callbackUrl) {
        this.callbackUrl = callbackUrl;
    }

    public Map<String, Object> getParameters() {
        return parameters;
    }

    public void setParameters(Map<String, Object> parameters) {
        this.parameters = parameters;
    }

    public Boolean getAllowRetry() {
        return allowRetry;
    }

    public void setAllowRetry(Boolean allowRetry) {
        this.allowRetry = allowRetry;
    }

    public ConcurrencyPolicy getConcurrencyPolicy() {
        return concurrencyPolicy;
    }

    public void setConcurrencyPolicy(ConcurrencyPolicy concurrencyPolicy) {
        this.concurrencyPolicy = concurrencyPolicy;
    }
}
