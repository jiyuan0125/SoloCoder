package com.canary.gateway.service;

import com.canary.gateway.config.CanaryConfig;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

import java.time.LocalDateTime;

@Service
public class HealthCheckService {
    private static final Logger log = LoggerFactory.getLogger(HealthCheckService.class);

    @Autowired
    private CanaryConfig canaryConfig;

    private final RestTemplate restTemplate = new RestTemplate();

    private volatile boolean healthy = true;
    private int successCount = 0;
    private int failCount = 0;
    private LocalDateTime lastStatusChange = LocalDateTime.now();

    @Scheduled(fixedRate = 10000)
    public void checkHealth() {
        String grayBackend = canaryConfig.getGrayBackend();
        if (grayBackend == null || grayBackend.isEmpty()) {
            return;
        }

        boolean currentHealthy = performHealthCheck(grayBackend);

        if (currentHealthy) {
            successCount++;
            failCount = 0;
            if (!healthy && successCount >= 3) {
                healthy = true;
                lastStatusChange = LocalDateTime.now();
                log.info("Gray backend health status changed to HEALTHY at {}", lastStatusChange);
            }
        } else {
            failCount++;
            successCount = 0;
            if (healthy && failCount >= 2) {
                healthy = false;
                lastStatusChange = LocalDateTime.now();
                log.warn("Gray backend health status changed to UNHEALTHY at {}", lastStatusChange);
            }
        }
    }

    private boolean performHealthCheck(String backend) {
        try {
            String healthUrl = backend.endsWith("/") ? backend + "actuator/health" : backend + "/actuator/health";
            ResponseEntity<String> response = restTemplate.getForEntity(healthUrl, String.class);
            return response.getStatusCode().is2xxSuccessful();
        } catch (Exception e) {
            log.debug("Health check failed for {}: {}", backend, e.getMessage());
            return false;
        }
    }

    public boolean isHealthy() {
        return healthy;
    }

    public LocalDateTime getLastStatusChange() {
        return lastStatusChange;
    }

    public void setHealthy(boolean healthy) {
        this.healthy = healthy;
        this.lastStatusChange = LocalDateTime.now();
    }
}
