package com.logaggregate.service.retention;

import com.logaggregate.model.RetentionConfig;
import com.logaggregate.service.storage.LogStorageService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.HashMap;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.TimeUnit;

@Service
public class RetentionService {

    private static final Logger logger = LoggerFactory.getLogger(RetentionService.class);
    private static final int DEFAULT_RETENTION_DAYS = 7;

    private final Map<String, Integer> retentionDaysByService = new ConcurrentHashMap<>();
    private final LogStorageService logStorageService;

    public RetentionService(LogStorageService logStorageService) {
        this.logStorageService = logStorageService;
    }

    public Map<String, RetentionConfig> getAllRetentionConfigs() {
        Map<String, RetentionConfig> result = new HashMap<>();
        Set<String> allServices = logStorageService.getAllServices();

        for (String service : allServices) {
            int days = retentionDaysByService.getOrDefault(service, DEFAULT_RETENTION_DAYS);
            result.put(service, new RetentionConfig(service, days));
        }

        for (Map.Entry<String, Integer> entry : retentionDaysByService.entrySet()) {
            result.putIfAbsent(entry.getKey(), new RetentionConfig(entry.getKey(), entry.getValue()));
        }

        return result;
    }

    public void setRetentionDays(String service, int days) {
        if (days <= 0) {
            throw new IllegalArgumentException("Retention days must be greater than 0");
        }
        retentionDaysByService.put(service, days);
        logger.info("Set retention days for {} to {}", service, days);
        cleanupServiceLogs(service, days, System.currentTimeMillis());
    }

    public RetentionConfig getRetentionConfig(String service) {
        int days = retentionDaysByService.getOrDefault(service, DEFAULT_RETENTION_DAYS);
        return new RetentionConfig(service, days);
    }

    @Scheduled(fixedRate = 1, timeUnit = TimeUnit.MINUTES)
    public void scheduledCleanup() {
        long now = System.currentTimeMillis();
        Set<String> services = logStorageService.getAllServices();
        logger.debug("Starting scheduled retention cleanup at {}", Instant.ofEpochMilli(now));

        for (String service : services) {
            int days = retentionDaysByService.getOrDefault(service, DEFAULT_RETENTION_DAYS);
            cleanupServiceLogs(service, days, now);
        }
    }

    private void cleanupServiceLogs(String service, int days, long now) {
        long cutoffTime = now - TimeUnit.DAYS.toMillis(days);
        try {
            logStorageService.deleteLogsByServiceBefore(service, cutoffTime);
        } catch (Exception e) {
            logger.error("Failed to cleanup logs for service {}: {}", service, e.getMessage(), e);
        }
    }
}
