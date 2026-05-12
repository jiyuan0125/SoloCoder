package com.health.store;

import com.health.config.HealthMonitorProperties;
import com.health.model.CheckHistory;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

import java.time.LocalDateTime;
import java.util.*;
import java.util.concurrent.ConcurrentLinkedDeque;
import java.util.stream.Collectors;

@Component
public class HistoryStore {

    private final Deque<CheckHistory> history = new ConcurrentLinkedDeque<>();
    private final HealthMonitorProperties properties;

    public HistoryStore(HealthMonitorProperties properties) {
        this.properties = properties;
    }

    public void add(CheckHistory record) {
        if (record.getId() == null) {
            record.setId(UUID.randomUUID().toString());
        }
        history.addFirst(record);
    }

    public List<CheckHistory> getByServiceId(String serviceId, int limit) {
        return history.stream()
                .filter(record -> serviceId.equals(record.getServiceId()))
                .limit(limit)
                .collect(Collectors.toList());
    }

    public List<CheckHistory> getByServiceId(String serviceId) {
        return getByServiceId(serviceId, 100);
    }

    public List<CheckHistory> getAll(int limit) {
        return history.stream()
                .limit(limit)
                .collect(Collectors.toList());
    }

    public List<CheckHistory> getAll() {
        return getAll(1000);
    }

    @Scheduled(fixedRate = 60000)
    public void cleanup() {
        LocalDateTime cutoff = LocalDateTime.now()
                .minusMinutes(properties.getHistoryRetentionMinutes());
        
        while (!history.isEmpty() && history.getLast().getCheckedAt().isBefore(cutoff)) {
            history.removeLast();
        }
    }
}