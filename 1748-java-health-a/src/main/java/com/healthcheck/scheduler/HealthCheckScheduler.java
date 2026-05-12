package com.healthcheck.scheduler;

import com.healthcheck.model.ServiceConfig;
import com.healthcheck.service.HealthCheckService;
import com.healthcheck.service.ServiceConfigStore;
import jakarta.annotation.PostConstruct;
import org.springframework.scheduling.TaskScheduler;
import org.springframework.scheduling.concurrent.ThreadPoolTaskScheduler;
import org.springframework.stereotype.Component;

import java.time.Duration;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ScheduledFuture;

@Component
public class HealthCheckScheduler {

    private final ServiceConfigStore configStore;
    private final HealthCheckService healthCheckService;
    private final TaskScheduler taskScheduler;

    private final Map<String, ScheduledFuture<?>> scheduledTasks = new ConcurrentHashMap<>();

    public HealthCheckScheduler(ServiceConfigStore configStore,
                                HealthCheckService healthCheckService) {
        this.configStore = configStore;
        this.healthCheckService = healthCheckService;
        
        ThreadPoolTaskScheduler scheduler = new ThreadPoolTaskScheduler();
        scheduler.setPoolSize(10);
        scheduler.setThreadNamePrefix("health-check-");
        scheduler.initialize();
        this.taskScheduler = scheduler;
    }

    @PostConstruct
    public void init() {
        configStore.getAllServices().forEach(this::scheduleService);
    }

    public void scheduleService(ServiceConfig config) {
        unscheduleService(config.getServiceId());

        healthCheckService.initializeServiceState(config);

        int interval = config.getCheckInterval() != null ? config.getCheckInterval() : 30;
        Duration period = Duration.ofSeconds(interval);

        ScheduledFuture<?> future = taskScheduler.scheduleAtFixedRate(
                () -> healthCheckService.checkService(config.getServiceId()),
                period);

        scheduledTasks.put(config.getServiceId(), future);
    }

    public void unscheduleService(String serviceId) {
        ScheduledFuture<?> future = scheduledTasks.remove(serviceId);
        if (future != null) {
            future.cancel(false);
        }
    }

    public void rescheduleAll() {
        configStore.getAllServices().forEach(this::scheduleService);
    }
}
