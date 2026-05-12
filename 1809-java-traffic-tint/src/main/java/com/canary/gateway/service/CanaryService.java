package com.canary.gateway.service;

import com.canary.gateway.config.CanaryConfig;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.nio.charset.StandardCharsets;
import java.time.LocalDateTime;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.Map;

@Service
public class CanaryService {
    private static final Logger log = LoggerFactory.getLogger(CanaryService.class);

    @Autowired
    private CanaryConfig canaryConfig;

    @Autowired
    private HealthCheckService healthCheckService;

    private String lastOperator = "system";
    private int lastRatio = 0;
    private LocalDateTime lastModified = LocalDateTime.now();

    private final Map<Long, Long> grayStats = Collections.synchronizedMap(new LinkedHashMap<>());
    private final Map<Long, Long> prodStats = Collections.synchronizedMap(new LinkedHashMap<>());

    public boolean shouldRouteToGray(String userId) {
        int ratio = canaryConfig.getRatio();
        if (ratio <= 0) {
            return false;
        }
        if (ratio >= 100) {
            return healthCheckService.isHealthy();
        }

        long hashValue = calculateHash(userId);
        boolean shouldRoute = (hashValue % 100) < ratio;

        if (shouldRoute && !healthCheckService.isHealthy()) {
            log.warn("Gray backend is unhealthy, falling back to prod for user: {}", userId);
            return false;
        }

        return shouldRoute;
    }

    private long calculateHash(String userId) {
        try {
            return Long.parseLong(userId);
        } catch (NumberFormatException e) {
            int hash = userId.hashCode();
            return Math.abs(hash);
        }
    }

    public String getTargetBackend(String userId) {
        if (shouldRouteToGray(userId)) {
            return canaryConfig.getGrayBackend();
        }
        return canaryConfig.getProdBackend();
    }

    public void incrementGrayStats() {
        long minute = System.currentTimeMillis() / 60000;
        cleanOldStats();
        grayStats.merge(minute, 1L, Long::sum);
    }

    public void incrementProdStats() {
        long minute = System.currentTimeMillis() / 60000;
        cleanOldStats();
        prodStats.merge(minute, 1L, Long::sum);
    }

    private void cleanOldStats() {
        long cutoff = System.currentTimeMillis() / 60000 - 60;
        grayStats.keySet().removeIf(k -> k < cutoff);
        prodStats.keySet().removeIf(k -> k < cutoff);
    }

    public Map<Long, Long> getGrayStats() {
        cleanOldStats();
        return new LinkedHashMap<>(grayStats);
    }

    public Map<Long, Long> getProdStats() {
        cleanOldStats();
        return new LinkedHashMap<>(prodStats);
    }

    public long getTotalGrayRequests() {
        return grayStats.values().stream().mapToLong(Long::longValue).sum();
    }

    public long getTotalProdRequests() {
        return prodStats.values().stream().mapToLong(Long::longValue).sum();
    }

    public synchronized void setRatio(int ratio, String operator) {
        if (ratio < 0 || ratio > 100) {
            throw new IllegalArgumentException("Ratio must be between 0 and 100");
        }
        int oldRatio = canaryConfig.getRatio();
        canaryConfig.setRatio(ratio);
        this.lastOperator = operator != null ? operator : "system";
        this.lastRatio = oldRatio;
        this.lastModified = LocalDateTime.now();
        log.info("Canary ratio changed from {} to {} by {} at {}", oldRatio, ratio, lastOperator, lastModified);
    }

    public synchronized void setGrayBackend(String backend) {
        canaryConfig.setGrayBackend(backend);
    }

    public synchronized void setProdBackend(String backend) {
        canaryConfig.setProdBackend(backend);
    }

    public String getLastOperator() {
        return lastOperator;
    }

    public int getLastRatio() {
        return lastRatio;
    }

    public LocalDateTime getLastModified() {
        return lastModified;
    }
}
