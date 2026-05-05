package com.employee.server.repository;

import com.employee.server.entity.OperationLog;
import org.springframework.stereotype.Repository;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;

@Repository
public class OperationLogRepository {
    private final Map<String, OperationLog> logMap = new ConcurrentHashMap<>();
    private final AtomicLong idCounter = new AtomicLong(0);

    public OperationLog save(OperationLog log) {
        if (log.getLogId() == null) {
            log.setLogId("LOG" + idCounter.incrementAndGet());
        }
        logMap.put(log.getLogId(), log);
        return log;
    }

    public List<OperationLog> findAll() {
        List<OperationLog> logs = new ArrayList<>(logMap.values());
        logs.sort(Comparator.comparing(OperationLog::getOperationTime).reversed());
        return logs;
    }

    public List<OperationLog> findByTargetEmployeeId(String employeeId) {
        return logMap.values().stream()
                .filter(log -> employeeId.equals(log.getTargetEmployeeId()))
                .sorted(Comparator.comparing(OperationLog::getOperationTime).reversed())
                .toList();
    }

    public List<OperationLog> findByOperatorId(String operatorId) {
        return logMap.values().stream()
                .filter(log -> operatorId.equals(log.getOperatorId()))
                .sorted(Comparator.comparing(OperationLog::getOperationTime).reversed())
                .toList();
    }
}
