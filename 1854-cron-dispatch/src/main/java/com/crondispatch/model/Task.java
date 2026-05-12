package com.crondispatch.model;

import com.crondispatch.enums.MisfireStrategy;
import com.crondispatch.enums.TaskStatus;
import com.crondispatch.enums.TaskType;

import java.time.Instant;
import java.util.Map;
import java.util.UUID;

public class Task {
    private String id;
    private String name;
    private TaskType type;
    private String cronExpression;
    private Long delaySeconds;
    private String callbackUrl;
    private Map<String, Object> params;
    private MisfireStrategy misfireStrategy;
    private boolean retryEnabled;
    private int maxRetryCount;
    private TaskStatus status;
    private Instant createdAt;
    private Instant nextExecutionTime;
    private int currentRetryCount;

    public Task() {
        this.id = UUID.randomUUID().toString().replace("-", "").substring(0, 16);
        this.createdAt = Instant.now();
        this.status = TaskStatus.PENDING;
        this.currentRetryCount = 0;
        this.retryEnabled = false;
        this.maxRetryCount = 3;
        this.misfireStrategy = MisfireStrategy.SKIPPED;
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

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

    public String getCronExpression() {
        return cronExpression;
    }

    public void setCronExpression(String cronExpression) {
        this.cronExpression = cronExpression;
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

    public Map<String, Object> getParams() {
        return params;
    }

    public void setParams(Map<String, Object> params) {
        this.params = params;
    }

    public MisfireStrategy getMisfireStrategy() {
        return misfireStrategy;
    }

    public void setMisfireStrategy(MisfireStrategy misfireStrategy) {
        this.misfireStrategy = misfireStrategy;
    }

    public boolean isRetryEnabled() {
        return retryEnabled;
    }

    public void setRetryEnabled(boolean retryEnabled) {
        this.retryEnabled = retryEnabled;
    }

    public int getMaxRetryCount() {
        return maxRetryCount;
    }

    public void setMaxRetryCount(int maxRetryCount) {
        this.maxRetryCount = maxRetryCount;
    }

    public TaskStatus getStatus() {
        return status;
    }

    public void setStatus(TaskStatus status) {
        this.status = status;
    }

    public Instant getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(Instant createdAt) {
        this.createdAt = createdAt;
    }

    public Instant getNextExecutionTime() {
        return nextExecutionTime;
    }

    public void setNextExecutionTime(Instant nextExecutionTime) {
        this.nextExecutionTime = nextExecutionTime;
    }

    public int getCurrentRetryCount() {
        return currentRetryCount;
    }

    public void setCurrentRetryCount(int currentRetryCount) {
        this.currentRetryCount = currentRetryCount;
    }
}
