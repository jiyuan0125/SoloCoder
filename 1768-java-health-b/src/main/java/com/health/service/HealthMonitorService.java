package com.health.service;

import com.health.checker.HealthChecker;
import com.health.model.CheckHistory;
import com.health.model.CheckResult;
import com.health.model.HealthStatus;
import com.health.model.ServiceRegistration;
import com.health.store.HistoryStore;
import com.health.store.ServiceRegistry;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;

@Slf4j
@Service
public class HealthMonitorService {

    private final ServiceRegistry serviceRegistry;
    private final HistoryStore historyStore;
    private final NotificationService notificationService;
    private final List<HealthChecker> healthCheckers;

    public HealthMonitorService(ServiceRegistry serviceRegistry,
                                HistoryStore historyStore,
                                NotificationService notificationService,
                                List<HealthChecker> healthCheckers) {
        this.serviceRegistry = serviceRegistry;
        this.historyStore = historyStore;
        this.notificationService = notificationService;
        this.healthCheckers = healthCheckers;
    }

    @Scheduled(fixedRate = 5000)
    public void runChecks() {
        List<ServiceRegistration> dueServices = serviceRegistry.findAllDueForCheck();
        
        for (ServiceRegistration service : dueServices) {
            checkService(service);
        }
    }

    public void checkService(ServiceRegistration service) {
        HealthChecker checker = findChecker(service);
        if (checker == null) {
            log.warn("No health checker found for service: {} type: {}", 
                    service.getName(), service.getCheckType());
            return;
        }

        HealthStatus previousStatus = service.getStatus();
        
        CheckResult result = checker.check(service);
        updateServiceStatus(service, result);
        
        saveHistory(service, result);
        
        checkAndNotify(service, previousStatus);
    }

    private HealthChecker findChecker(ServiceRegistration service) {
        return healthCheckers.stream()
                .filter(checker -> checker.supports(service))
                .findFirst()
                .orElse(null);
    }

    private void updateServiceStatus(ServiceRegistration service, CheckResult result) {
        service.setLastCheckedAt(LocalDateTime.now());
        service.setLastMessage(result.getMessage());

        if (result.isHealthy()) {
            service.setConsecutiveFailures(0);
            if (service.getStatus() != HealthStatus.HEALTHY) {
                service.setStatus(HealthStatus.HEALTHY);
                log.info("Service {} is now HEALTHY", service.getName());
            }
        } else {
            int currentFailures = service.getConsecutiveFailures() + 1;
            service.setConsecutiveFailures(currentFailures);
            
            int threshold = service.getUnhealthyThreshold() != null ? 
                    service.getUnhealthyThreshold() : 3;
            
            if (currentFailures >= threshold && service.getStatus() != HealthStatus.UNHEALTHY) {
                service.setStatus(HealthStatus.UNHEALTHY);
                log.warn("Service {} is now UNHEALTHY after {} consecutive failures", 
                        service.getName(), currentFailures);
            }
        }
        
        serviceRegistry.update(service);
    }

    private void saveHistory(ServiceRegistration service, CheckResult result) {
        CheckHistory history = CheckHistory.builder()
                .serviceId(service.getId())
                .serviceName(service.getName())
                .checkedAt(result.getCheckedAt())
                .status(result.isHealthy() ? HealthStatus.HEALTHY : HealthStatus.UNHEALTHY)
                .responseTimeMs(result.getResponseTimeMs())
                .message(result.getMessage())
                .build();
        
        historyStore.add(history);
    }

    private void checkAndNotify(ServiceRegistration service, HealthStatus previousStatus) {
        HealthStatus currentStatus = service.getStatus();
        
        if (previousStatus != currentStatus) {
            if (currentStatus == HealthStatus.UNHEALTHY) {
                notificationService.notifyUnhealthy(service);
            } else if (currentStatus == HealthStatus.HEALTHY && 
                    previousStatus == HealthStatus.UNHEALTHY) {
                notificationService.notifyRecovered(service);
            }
        }
    }
}