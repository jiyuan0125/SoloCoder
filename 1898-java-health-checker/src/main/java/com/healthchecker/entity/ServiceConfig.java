package com.healthchecker.entity;

import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

import java.time.LocalDateTime;
import java.util.Deque;
import java.util.concurrent.ConcurrentLinkedDeque;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.locks.ReadWriteLock;
import java.util.concurrent.locks.ReentrantReadWriteLock;

@Data
public class ServiceConfig {

    @NotBlank(message = "服务名称不能为空")
    private String serviceName;

    @NotBlank(message = "检查URL不能为空")
    private String checkUrl;

    @Min(value = 1, message = "检查间隔至少为1秒")
    private int checkIntervalSeconds = 30;

    @Min(value = 1, message = "超时时间至少为1秒")
    private int timeoutSeconds = 5;

    private String callbackUrl;

    private ServiceStatus currentStatus = ServiceStatus.UNKNOWN;

    private LocalDateTime lastCheckTime;

    private final AtomicInteger consecutiveSuccess = new AtomicInteger(0);

    private final AtomicInteger consecutiveFailure = new AtomicInteger(0);

    private final Deque<CheckResult> history = new ConcurrentLinkedDeque<>();

    private final ReadWriteLock maintenanceLock = new ReentrantReadWriteLock();

    private volatile boolean underMaintenance = false;

    private volatile boolean inProgress = false;

    public void incrementConsecutiveSuccess() {
        consecutiveSuccess.incrementAndGet();
        consecutiveFailure.set(0);
    }

    public void incrementConsecutiveFailure() {
        consecutiveFailure.incrementAndGet();
        consecutiveSuccess.set(0);
    }

    public int getConsecutiveSuccessCount() {
        return consecutiveSuccess.get();
    }

    public int getConsecutiveFailureCount() {
        return consecutiveFailure.get();
    }

    public void resetCounters() {
        consecutiveSuccess.set(0);
        consecutiveFailure.set(0);
    }

    public void addHistory(CheckResult result) {
        synchronized (history) {
            if (history.size() >= 100) {
                history.pollFirst();
            }
            history.addLast(result);
        }
    }

    public Deque<CheckResult> getHistorySnapshot() {
        synchronized (history) {
            return new ConcurrentLinkedDeque<>(history);
        }
    }

    public boolean isUnderMaintenance() {
        return underMaintenance;
    }

    public void setUnderMaintenance(boolean underMaintenance) {
        this.underMaintenance = underMaintenance;
    }
}