package com.gateway.auth.service;

import com.gateway.auth.config.AuditConfig;
import com.gateway.auth.model.AuditLog;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.concurrent.locks.ReadWriteLock;
import java.util.concurrent.locks.ReentrantReadWriteLock;
import java.util.stream.Collectors;

@Service
public class AuditLogService {

    private final AuditConfig auditConfig;
    private final List<AuditLog> logs;
    private final ReadWriteLock lock = new ReentrantReadWriteLock();

    public AuditLogService(AuditConfig auditConfig) {
        this.auditConfig = auditConfig;
        this.logs = new ArrayList<>(auditConfig.getMaxRecords());
    }

    public void log(AuditLog auditLog) {
        lock.writeLock().lock();
        try {
            while (logs.size() >= auditConfig.getMaxRecords()) {
                logs.remove(0);
            }
            logs.add(auditLog);
        } finally {
            lock.writeLock().unlock();
        }
    }

    public List<AuditLog> getLogs(String clientId, Instant startTime, Instant endTime) {
        lock.readLock().lock();
        try {
            List<AuditLog> result = new ArrayList<>(logs);

            if (clientId != null && !clientId.trim().isEmpty()) {
                result = result.stream()
                        .filter(log -> clientId.equals(log.getClientId()))
                        .collect(Collectors.toList());
            }

            if (startTime != null) {
                result = result.stream()
                        .filter(log -> !log.getTimestamp().isBefore(startTime))
                        .collect(Collectors.toList());
            }

            if (endTime != null) {
                result = result.stream()
                        .filter(log -> !log.getTimestamp().isAfter(endTime))
                        .collect(Collectors.toList());
            }

            return result;
        } finally {
            lock.readLock().unlock();
        }
    }

    public List<AuditLog> getLogs(String clientId) {
        return getLogs(clientId, null, null);
    }

    public List<AuditLog> getAllLogs() {
        return getLogs(null, null, null);
    }
}
