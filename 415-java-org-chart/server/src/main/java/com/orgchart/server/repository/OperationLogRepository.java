package com.orgchart.server.repository;

import com.orgchart.server.entity.OperationLog;
import org.springframework.stereotype.Repository;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Repository
public class OperationLogRepository {

    private final Map<String, OperationLog> logs = new ConcurrentHashMap<>();

    public OperationLog save(OperationLog log) {
        logs.put(log.getId(), log);
        return log;
    }

    public Optional<OperationLog> findById(String id) {
        return Optional.ofNullable(logs.get(id));
    }

    public List<OperationLog> findAll() {
        return logs.values().stream()
                .sorted(Comparator.comparing(OperationLog::getOperationTime).reversed())
                .collect(Collectors.toList());
    }

    public List<OperationLog> findByTargetTypeAndTargetId(String targetType, String targetId) {
        return logs.values().stream()
                .filter(log -> targetType.equals(log.getTargetType()) && targetId.equals(log.getTargetId()))
                .sorted(Comparator.comparing(OperationLog::getOperationTime).reversed())
                .collect(Collectors.toList());
    }

    public List<OperationLog> findByOperationType(String operationType) {
        return logs.values().stream()
                .filter(log -> operationType.equals(log.getOperationType()))
                .sorted(Comparator.comparing(OperationLog::getOperationTime).reversed())
                .collect(Collectors.toList());
    }
}
