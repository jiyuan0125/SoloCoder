package com.healthchecker.service;

import com.healthchecker.dto.ServiceStatusSummary;
import com.healthchecker.engine.StateTransitionEngine;
import com.healthchecker.entity.CheckResult;
import com.healthchecker.entity.ServiceConfig;
import com.healthchecker.entity.ServiceStatus;
import com.healthchecker.entity.StatusChangeNotification;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.TaskScheduler;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ScheduledFuture;

@Service
public class HealthCheckService {

    private static final Logger logger = LoggerFactory.getLogger(HealthCheckService.class);

    private final Map<String, ServiceConfig> services = new ConcurrentHashMap<>();
    private final Map<String, ScheduledFuture<?>> schedules = new ConcurrentHashMap<>();

    private final TaskScheduler taskScheduler;
    private final HttpChecker httpChecker;
    private final NotificationService notificationService;
    private final StateTransitionEngine stateEngine;

    public HealthCheckService(TaskScheduler taskScheduler,
                              HttpChecker httpChecker,
                              NotificationService notificationService,
                              StateTransitionEngine stateEngine) {
        this.taskScheduler = taskScheduler;
        this.httpChecker = httpChecker;
        this.notificationService = notificationService;
        this.stateEngine = stateEngine;
    }

    public boolean registerService(ServiceConfig config) {
        String serviceName = config.getServiceName();
        if (services.containsKey(serviceName)) {
            logger.warn("服务已存在: {}", serviceName);
            return false;
        }

        services.put(serviceName, config);
        scheduleService(config);
        logger.info("服务已注册并开始调度: {}", serviceName);
        return true;
    }

    private void scheduleService(ServiceConfig config) {
        String serviceName = config.getServiceName();
        long intervalMs = (long) config.getCheckIntervalSeconds() * 1000;

        ScheduledFuture<?> future = taskScheduler.scheduleWithFixedDelay(
                () -> performCheck(serviceName),
                new Date(System.currentTimeMillis() + 1000),
                intervalMs
        );

        ScheduledFuture<?> oldFuture = schedules.put(serviceName, future);
        if (oldFuture != null) {
            oldFuture.cancel(false);
        }
    }

    public void performCheck(String serviceName) {
        ServiceConfig config = services.get(serviceName);
        if (config == null) {
            logger.warn("服务不存在: {}", serviceName);
            return;
        }

        if (config.isUnderMaintenance()) {
            logger.debug("服务处于维护中，跳过检查: {}", serviceName);
            return;
        }

        if (!config.getMaintenanceLock().writeLock().tryLock()) {
            logger.debug("无法获取维护锁，跳过本次检查: {}", serviceName);
            return;
        }

        try {
            if (config.isUnderMaintenance()) {
                logger.debug("获取锁后发现服务处于维护中，跳过检查: {}", serviceName);
                return;
            }

            config.setInProgress(true);
            logger.debug("开始检查服务: {}", serviceName);

            CheckResult result = httpChecker.check(config);
            config.setLastCheckTime(result.timestamp());
            config.addHistory(result);

            if (config.isUnderMaintenance()) {
                logger.info("服务在检查期间进入维护中，忽略本次检查结果: {}", serviceName);
                return;
            }

            StateTransitionEngine.TransitionResult transition =
                    stateEngine.processCheckResult(config, result.success());

            if (transition.statusChanged()) {
                logger.info("服务 [{}] 状态变更: {} -> {}",
                        serviceName, transition.oldStatus(), transition.newStatus());

                StatusChangeNotification notification = new StatusChangeNotification(
                        serviceName,
                        transition.oldStatus(),
                        transition.newStatus(),
                        LocalDateTime.now(),
                        transition.consecutiveCount()
                );

                notificationService.sendNotification(config.getCallbackUrl(), notification);
            }

        } finally {
            config.setInProgress(false);
            config.getMaintenanceLock().writeLock().unlock();
        }
    }

    public boolean triggerImmediateCheck(String serviceName) {
        ServiceConfig config = services.get(serviceName);
        if (config == null) {
            return false;
        }

        new Thread(() -> performCheck(serviceName)).start();
        return true;
    }

    public List<ServiceStatusSummary> getAllStatusSummary() {
        List<ServiceStatusSummary> summaries = new ArrayList<>();

        for (ServiceConfig config : services.values()) {
            ServiceStatus status = config.isUnderMaintenance()
                    ? ServiceStatus.MAINTENANCE
                    : config.getCurrentStatus();

            summaries.add(new ServiceStatusSummary(
                    config.getServiceName(),
                    status,
                    config.getLastCheckTime(),
                    config.getConsecutiveSuccessCount(),
                    config.getConsecutiveFailureCount()
            ));
        }

        summaries.sort(Comparator.comparing(ServiceStatusSummary::serviceName));
        return summaries;
    }

    public ServiceConfig getService(String serviceName) {
        return services.get(serviceName);
    }

    public List<CheckResult> getServiceHistory(String serviceName) {
        ServiceConfig config = services.get(serviceName);
        if (config == null) {
            return Collections.emptyList();
        }
        return new ArrayList<>(config.getHistorySnapshot());
    }

    public boolean setMaintenance(String serviceName) {
        ServiceConfig config = services.get(serviceName);
        if (config == null) {
            return false;
        }

        config.getMaintenanceLock().writeLock().lock();
        try {
            config.setUnderMaintenance(true);
            config.setCurrentStatus(ServiceStatus.MAINTENANCE);
            logger.info("服务设置为维护中: {}", serviceName);
            return true;
        } finally {
            config.getMaintenanceLock().writeLock().unlock();
        }
    }

    public boolean cancelMaintenance(String serviceName) {
        ServiceConfig config = services.get(serviceName);
        if (config == null) {
            return false;
        }

        config.getMaintenanceLock().writeLock().lock();
        try {
            if (!config.isUnderMaintenance()) {
                return true;
            }

            config.setUnderMaintenance(false);
            stateEngine.resetToUnknown(config);
            logger.info("服务取消维护中，重置为未知状态: {}", serviceName);
            return true;
        } finally {
            config.getMaintenanceLock().writeLock().unlock();
        }
    }
}