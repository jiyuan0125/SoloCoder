package com.example.healthchecker.scheduler;

import com.example.healthchecker.service.HealthCheckService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

@Component
public class HealthCheckScheduler {

    private static final Logger logger = LoggerFactory.getLogger(HealthCheckScheduler.class);

    private final HealthCheckService healthCheckService;

    public HealthCheckScheduler(HealthCheckService healthCheckService) {
        this.healthCheckService = healthCheckService;
    }

    @Scheduled(fixedRate = 10000)
    public void recalculateHealthStatus() {
        logger.debug("Scheduled health status recalculation started");
        healthCheckService.recalculateAll("scheduled recalculation");
        logger.debug("Scheduled health status recalculation completed");
    }
}
