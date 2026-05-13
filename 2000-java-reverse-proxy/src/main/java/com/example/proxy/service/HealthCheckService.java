package com.example.proxy.service;

import com.example.proxy.config.ProxyProperties;
import com.example.proxy.model.Backend;
import com.example.proxy.model.Route;
import lombok.extern.slf4j.Slf4j;
import org.apache.hc.client5.http.classic.methods.HttpGet;
import org.apache.hc.client5.http.classic.methods.HttpHead;
import org.apache.hc.client5.http.impl.classic.CloseableHttpClient;
import org.apache.hc.core5.http.ClassicHttpResponse;
import org.apache.hc.core5.http.HttpStatus;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import java.time.LocalDateTime;
import java.util.List;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

@Slf4j
@Service
public class HealthCheckService {
    
    private final ProxyProperties properties;
    private final RouteService routeService;
    private final CloseableHttpClient httpClient;
    private ExecutorService executorService;
    
    public HealthCheckService(ProxyProperties properties, RouteService routeService, 
                              CloseableHttpClient httpClient) {
        this.properties = properties;
        this.routeService = routeService;
        this.httpClient = httpClient;
    }
    
    @PostConstruct
    public void init() {
        this.executorService = Executors.newCachedThreadPool();
        log.info("HealthCheckService initialized with interval: {}ms", 
                properties.getHealthCheck().getInterval());
    }
    
    @Scheduled(fixedRateString = "${proxy.health-check.interval:10000}")
    public void checkAllBackends() {
        log.debug("Starting health check cycle...");
        List<Route> routes = routeService.getAllRoutes();
        
        for (Route route : routes) {
            for (Backend backend : route.getBackends()) {
                executorService.submit(() -> checkBackend(backend));
            }
        }
    }
    
    private void checkBackend(Backend backend) {
        try {
            boolean needProbeCheck = backend.getStatus() == Backend.BackendStatus.UNHEALTHY && 
                    shouldStartProbing(backend);
            
            if (needProbeCheck) {
                log.info("Starting probe check for backend: {}", backend.getUrl());
                backend.setStatus(Backend.BackendStatus.PROBING);
            }
            
            if (backend.getStatus() == Backend.BackendStatus.HEALTHY || 
                    backend.getStatus() == Backend.BackendStatus.PROBING) {
                performHealthCheck(backend);
            }
        } catch (Exception e) {
            log.error("Error during health check for backend {}: {}", backend.getUrl(), e.getMessage());
        }
    }
    
    private boolean shouldStartProbing(Backend backend) {
        if (backend.getLastFailureTime() == null) {
            return false;
        }
        long recoveryTimeMs = properties.getFailureRecoveryTime();
        return LocalDateTime.now().isAfter(
                backend.getLastFailureTime().plusNanos(recoveryTimeMs * 1_000_000));
    }
    
    private void performHealthCheck(Backend backend) {
        String healthUrl = buildHealthCheckUrl(backend.getUrl());
        backend.setLastProbeTime(LocalDateTime.now());
        
        try {
            boolean success;
            if (backend.getStatus() == Backend.BackendStatus.PROBING) {
                success = performProbeCheck(backend);
            } else {
                success = performRegularCheck(healthUrl);
            }
            
            handleCheckResult(backend, success);
            
        } catch (Exception e) {
            log.warn("Health check failed for backend {}: {}", backend.getUrl(), e.getMessage());
            handleCheckFailure(backend);
        }
    }
    
    private boolean performRegularCheck(String healthUrl) throws Exception {
        HttpGet request = new HttpGet(healthUrl);
        
        try (ClassicHttpResponse response = httpClient.execute(request)) {
            int statusCode = response.getCode();
            return statusCode >= HttpStatus.SC_OK && statusCode < HttpStatus.SC_MULTIPLE_CHOICES;
        }
    }
    
    private boolean performProbeCheck(Backend backend) throws Exception {
        String probeUrl = backend.getUrl();
        HttpHead request = new HttpHead(probeUrl);
        
        try (ClassicHttpResponse response = httpClient.execute(request)) {
            int statusCode = response.getCode();
            return statusCode >= HttpStatus.SC_OK && statusCode < HttpStatus.SC_MULTIPLE_CHOICES;
        }
    }
    
    private void handleCheckResult(Backend backend, boolean success) {
        if (success) {
            backend.setConsecutiveSuccesses(backend.getConsecutiveSuccesses() + 1);
            backend.setConsecutiveFailures(0);
            backend.setLastSuccessTime(LocalDateTime.now());
            
            if (backend.getStatus() == Backend.BackendStatus.PROBING) {
                if (backend.getConsecutiveSuccesses() >= properties.getHealthCheck().getConsecutiveSuccessesRecovery()) {
                    backend.setStatus(Backend.BackendStatus.HEALTHY);
                    backend.setConsecutive5xxCount(0);
                    log.info("Backend {} recovered to HEALTHY status", backend.getUrl());
                }
            } else if (backend.getStatus() == Backend.BackendStatus.HEALTHY) {
                log.debug("Health check passed for backend: {}", backend.getUrl());
            }
        } else {
            handleCheckFailure(backend);
        }
    }
    
    private void handleCheckFailure(Backend backend) {
        backend.setConsecutiveFailures(backend.getConsecutiveFailures() + 1);
        backend.setConsecutiveSuccesses(0);
        backend.setLastFailureTime(LocalDateTime.now());
        
        if (backend.getStatus() == Backend.BackendStatus.HEALTHY && 
                backend.getConsecutiveFailures() >= properties.getHealthCheck().getConsecutiveFailuresThreshold()) {
            backend.setStatus(Backend.BackendStatus.UNHEALTHY);
            log.warn("Backend {} marked as UNHEALTHY after {} consecutive failures", 
                    backend.getUrl(), backend.getConsecutiveFailures());
        } else if (backend.getStatus() == Backend.BackendStatus.PROBING) {
            backend.setStatus(Backend.BackendStatus.UNHEALTHY);
            log.info("Probe check failed for backend {}, returning to UNHEALTHY: {}", 
                    backend.getUrl(), backend.getConsecutiveFailures());
        }
        
        log.warn("Health check failed for backend {} (consecutive failures: {})", 
                backend.getUrl(), backend.getConsecutiveFailures());
    }
    
    private String buildHealthCheckUrl(String baseUrl) {
        String healthPath = properties.getHealthCheck().getPath();
        if (baseUrl.endsWith("/") && healthPath.startsWith("/")) {
            return baseUrl + healthPath.substring(1);
        }
        if (!baseUrl.endsWith("/") && !healthPath.startsWith("/")) {
            return baseUrl + "/" + healthPath;
        }
        return baseUrl + healthPath;
    }
    
    public void markBackendUnhealthy(Backend backend, String reason) {
        backend.setConsecutive5xxCount(backend.getConsecutive5xxCount() + 1);
        backend.setLastFailureTime(LocalDateTime.now());
        
        if (backend.getConsecutive5xxCount() >= properties.getConsecutive5xxThreshold()) {
            backend.setStatus(Backend.BackendStatus.UNHEALTHY);
            log.warn("Backend {} marked as UNHEALTHY due to {} consecutive 5xx responses. Reason: {}", 
                    backend.getUrl(), backend.getConsecutive5xxCount(), reason);
        }
    }
    
    public void markBackendTimeout(Backend backend) {
        backend.setLastFailureTime(LocalDateTime.now());
        backend.setStatus(Backend.BackendStatus.UNHEALTHY);
        log.warn("Backend {} marked as UNHEALTHY due to connection timeout", backend.getUrl());
    }
    
    public void resetBackendStatus(Backend backend) {
        backend.setConsecutiveFailures(0);
        backend.setConsecutiveSuccesses(0);
        backend.setConsecutive5xxCount(0);
    }
}
