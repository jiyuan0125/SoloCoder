package com.loadbalancer.service;

import com.loadbalancer.model.DistributionHistory;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.Deque;
import java.util.LinkedList;
import java.util.List;
import java.util.concurrent.locks.ReentrantReadWriteLock;
import java.util.stream.Collectors;

@Service
public class DistributionHistoryService {
    private static final int MAX_HISTORY_SIZE = 1000;
    private final Deque<DistributionHistory> historyDeque = new LinkedList<>();
    private final ReentrantReadWriteLock lock = new ReentrantReadWriteLock();

    public void record(String nodeId, String reason, String requestPath, String routeRuleId) {
        DistributionHistory record = new DistributionHistory(
                nodeId, reason, requestPath, Instant.now(), routeRuleId);
        lock.writeLock().lock();
        try {
            historyDeque.addFirst(record);
            if (historyDeque.size() > MAX_HISTORY_SIZE) {
                historyDeque.removeLast();
            }
        } finally {
            lock.writeLock().unlock();
        }
    }

    public List<DistributionHistory> getRecentHistory(int limit) {
        lock.readLock().lock();
        try {
            return historyDeque.stream()
                    .limit(limit)
                    .collect(Collectors.toList());
        } finally {
            lock.readLock().unlock();
        }
    }

    public List<DistributionHistory> getHistoryByTimeRange(Instant startTime, Instant endTime) {
        lock.readLock().lock();
        try {
            return historyDeque.stream()
                    .filter(h -> !h.getTimestamp().isBefore(startTime) && !h.getTimestamp().isAfter(endTime))
                    .collect(Collectors.toList());
        } finally {
            lock.readLock().unlock();
        }
    }

    public List<DistributionHistory> getHistoryByNode(String nodeId, int limit) {
        lock.readLock().lock();
        try {
            return historyDeque.stream()
                    .filter(h -> nodeId.equals(h.getNodeId()))
                    .limit(limit)
                    .collect(Collectors.toList());
        } finally {
            lock.readLock().unlock();
        }
    }
}
