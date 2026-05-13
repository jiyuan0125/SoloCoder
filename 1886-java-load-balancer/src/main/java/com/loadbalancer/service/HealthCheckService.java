package com.loadbalancer.service;

import com.loadbalancer.config.LoadBalancerConfig;
import com.loadbalancer.model.Node;
import com.loadbalancer.model.NodeStatus;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.time.Instant;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

@Slf4j
@Service
@RequiredArgsConstructor
public class HealthCheckService {
    private final NodeRegistryService nodeRegistryService;
    private final LoadBalancerConfig loadBalancerConfig;
    private final HttpClient httpClient = HttpClient.newHttpClient();
    private final ScheduledExecutorService scheduler = Executors.newScheduledThreadPool(10);
    private final Map<String, Instant> lastCheckTimes = new ConcurrentHashMap<>();

    @Scheduled(fixedRate = 1000)
    public void triggerHealthChecks() {
        for (Node node : nodeRegistryService.getAllNodes()) {
            if (node.getStatus() == NodeStatus.OFFLINE) {
                continue;
            }
            
            Instant now = Instant.now();
            Instant lastCheck = lastCheckTimes.get(node.getId());
            int intervalSeconds = node.getHealthCheckIntervalSeconds();
            
            if (lastCheck == null || 
                Duration.between(lastCheck, now).getSeconds() >= intervalSeconds) {
                scheduler.schedule(() -> performHealthCheck(node), 0, TimeUnit.MILLISECONDS);
            }
        }
    }

    private void performHealthCheck(Node node) {
        String healthUrl = buildHealthUrl(node);
        int timeoutSeconds = node.getHealthCheckTimeoutSeconds();
        
        try {
            HttpRequest request = HttpRequest.newBuilder()
                    .uri(URI.create(healthUrl))
                    .timeout(Duration.ofSeconds(timeoutSeconds))
                    .GET()
                    .build();
            
            HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
            boolean success = response.statusCode() >= 200 && response.statusCode() < 400;
            
            log.debug("Health check for {}: status={}", node.getAddress(), response.statusCode());
            nodeRegistryService.processHealthCheckResult(node, success);
            lastCheckTimes.put(node.getId(), Instant.now());
            
        } catch (Exception e) {
            log.warn("Health check failed for {}: {}", node.getAddress(), e.getMessage());
            nodeRegistryService.processHealthCheckResult(node, false);
            lastCheckTimes.put(node.getId(), Instant.now());
        }
    }

    private String buildHealthUrl(Node node) {
        String address = node.getAddress();
        String path = loadBalancerConfig.getPath();
        if (path.startsWith("/")) {
            return "http://" + address + path;
        }
        return "http://" + address + "/" + path;
    }
}
