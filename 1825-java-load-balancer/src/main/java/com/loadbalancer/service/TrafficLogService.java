package com.loadbalancer.service;

import com.loadbalancer.model.TrafficLog;
import com.loadbalancer.model.TrafficLog.SelectionReason;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.*;
import java.util.concurrent.ConcurrentLinkedDeque;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.stream.Collectors;

@Service
public class TrafficLogService {
    @Value("${loadbalancer.traffic-log.max-entries:1000}")
    private int maxEntries;

    private final ConcurrentLinkedDeque<TrafficLog> logs = new ConcurrentLinkedDeque<>();
    private final AtomicInteger logIdGenerator = new AtomicInteger(0);

    public void record(String nodeId, SelectionReason reason, String requestPath) {
        TrafficLog log = TrafficLog.builder()
                .id("log-" + logIdGenerator.incrementAndGet())
                .nodeId(nodeId)
                .reason(reason)
                .requestPath(requestPath)
                .timestamp(LocalDateTime.now())
                .build();
        logs.addFirst(log);

        while (logs.size() > maxEntries) {
            logs.pollLast();
        }
    }

    public List<TrafficLog> getRecentLogs(int limit) {
        int actualLimit = Math.min(limit, logs.size());
        return logs.stream().limit(actualLimit).collect(Collectors.toList());
    }

    public List<TrafficLog> getLogsByNode(String nodeId, int limit) {
        return logs.stream()
                .filter(log -> nodeId.equals(log.getNodeId()))
                .limit(limit)
                .collect(Collectors.toList());
    }

    public List<TrafficLog> getLogsByTimeRange(LocalDateTime start, LocalDateTime end, int limit) {
        return logs.stream()
                .filter(log -> {
                    LocalDateTime ts = log.getTimestamp();
                    return !ts.isBefore(start) && !ts.isAfter(end);
                })
                .limit(limit)
                .collect(Collectors.toList());
    }

    public List<TrafficLog> getLogs(String nodeId, LocalDateTime start, LocalDateTime end, int limit) {
        return logs.stream()
                .filter(log -> {
                    if (nodeId != null && !nodeId.equals(log.getNodeId())) {
                        return false;
                    }
                    LocalDateTime ts = log.getTimestamp();
                    if (start != null && ts.isBefore(start)) {
                        return false;
                    }
                    if (end != null && ts.isAfter(end)) {
                        return false;
                    }
                    return true;
                })
                .limit(limit)
                .collect(Collectors.toList());
    }
}
