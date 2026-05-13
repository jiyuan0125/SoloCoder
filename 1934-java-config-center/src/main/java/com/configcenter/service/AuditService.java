package com.configcenter.service;

import com.configcenter.model.AuditAction;
import com.configcenter.model.AuditLog;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.Comparator;
import java.util.List;
import java.util.UUID;
import java.util.concurrent.ConcurrentLinkedDeque;
import java.util.stream.Collectors;
import java.util.stream.Stream;

@Service
public class AuditService {

    private static final int MAX_LOGS = 1000;
    private final ConcurrentLinkedDeque<AuditLog> logs = new ConcurrentLinkedDeque<>();

    public AuditLog createLog(String userId, String appName, AuditAction action,
                              String oldValue, String newValue, String configKey) {
        AuditLog log = new AuditLog();
        log.setId(UUID.randomUUID().toString());
        log.setUserId(userId);
        log.setAppName(appName);
        log.setAction(action);
        log.setOldValue(oldValue);
        log.setNewValue(newValue);
        log.setConfigKey(configKey);
        log.setTimestamp(Instant.now());
        return log;
    }

    public void saveLog(AuditLog log) {
        logs.addLast(log);
        while (logs.size() > MAX_LOGS) {
            logs.pollFirst();
        }
    }

    public List<AuditLog> queryLogs(String userId, String appName, 
                                     Instant startTime, Instant endTime) {
        Stream<AuditLog> stream = logs.stream();
        
        if (userId != null && !userId.isEmpty()) {
            stream = stream.filter(log -> userId.equals(log.getUserId()));
        }
        
        if (appName != null && !appName.isEmpty()) {
            stream = stream.filter(log -> appName.equals(log.getAppName()));
        }
        
        if (startTime != null) {
            stream = stream.filter(log -> !log.getTimestamp().isBefore(startTime));
        }
        
        if (endTime != null) {
            stream = stream.filter(log -> !log.getTimestamp().isAfter(endTime));
        }
        
        return stream
                .sorted(Comparator.comparing(AuditLog::getTimestamp).reversed())
                .collect(Collectors.toList());
    }
}
