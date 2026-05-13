package com.example.proxy.service;

import com.example.proxy.config.ProxyProperties;
import com.example.proxy.dto.AccessLogStats;
import com.example.proxy.model.AccessLog;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentLinkedDeque;
import java.util.stream.Collectors;

@Slf4j
@Service
public class AccessLogService {
    
    private final ProxyProperties properties;
    private final Deque<AccessLog> logs;
    private final Map<String, Long> backendRequestCount = new ConcurrentHashMap<>();
    private final Map<String, Long> routeRequestCount = new ConcurrentHashMap<>();
    private long totalRequests = 0;
    private long successCount = 0;
    private long errorCount = 0;
    private long total4xxCount = 0;
    private long total5xxCount = 0;
    private long totalResponseTime = 0;
    
    public AccessLogService(ProxyProperties properties) {
        this.properties = properties;
        this.logs = new ConcurrentLinkedDeque<>();
    }
    
    @PostConstruct
    public void init() {
        log.info("AccessLogService initialized with max entries: {}", 
                properties.getAccessLog().getMaxEntries());
    }
    
    public void record(AccessLog accessLog) {
        logs.addLast(accessLog);
        
        while (logs.size() > properties.getAccessLog().getMaxEntries()) {
            logs.removeFirst();
        }
        
        totalRequests++;
        
        int status = accessLog.getResponseStatus();
        if (status >= 200 && status < 400) {
            successCount++;
        } else {
            errorCount++;
            if (status >= 400 && status < 500) {
                total4xxCount++;
            } else if (status >= 500) {
                total5xxCount++;
            }
        }
        
        totalResponseTime += accessLog.getDurationMs();
        
        if (accessLog.getTargetBackend() != null) {
            backendRequestCount.merge(accessLog.getTargetBackend(), 1L, Long::sum);
        }
        
        if (accessLog.getRequestPath() != null) {
            String routeKey = extractRouteKey(accessLog.getRequestPath());
            routeRequestCount.merge(routeKey, 1L, Long::sum);
        }
    }
    
    private String extractRouteKey(String path) {
        if (path == null || path.isEmpty()) {
            return "/";
        }
        
        String[] parts = path.split("/");
        if (parts.length >= 3) {
            return "/" + parts[1] + "/" + parts[2];
        }
        return path;
    }
    
    public List<AccessLog> getRecentLogs(int limit) {
        List<AccessLog> result = new ArrayList<>(logs);
        Collections.reverse(result);
        return result.stream().limit(limit).collect(Collectors.toList());
    }
    
    public AccessLogStats getStats() {
        List<Long> responseTimes = logs.stream()
                .map(AccessLog::getDurationMs)
                .sorted()
                .collect(Collectors.toList());
        
        double p95ResponseTime = 0;
        if (!responseTimes.isEmpty()) {
            int p95Index = (int) Math.ceil(responseTimes.size() * 0.95) - 1;
            p95ResponseTime = responseTimes.get(Math.max(0, p95Index));
        }
        
        return AccessLogStats.builder()
                .totalRequests(totalRequests)
                .successCount(successCount)
                .errorCount(errorCount)
                .total4xxCount(total4xxCount)
                .total5xxCount(total5xxCount)
                .avgResponseTime(totalRequests > 0 ? (double) totalResponseTime / totalRequests : 0)
                .p95ResponseTime(p95ResponseTime)
                .requestsByRoute(new HashMap<>(routeRequestCount))
                .requestsByBackend(new HashMap<>(backendRequestCount))
                .build();
    }
    
    public void clearStats() {
        logs.clear();
        backendRequestCount.clear();
        routeRequestCount.clear();
        totalRequests = 0;
        successCount = 0;
        errorCount = 0;
        total4xxCount = 0;
        total5xxCount = 0;
        totalResponseTime = 0;
        log.info("AccessLog stats cleared");
    }
    
    public int getLogCount() {
        return logs.size();
    }
}
