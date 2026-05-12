package com.logaggregate.service.storage;

import com.logaggregate.model.StoredLogEntry;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.stream.Collectors;

@Service
public class InMemoryLogStorageService implements LogStorageService {

    private static final Logger logger = LoggerFactory.getLogger(InMemoryLogStorageService.class);

    private final ConcurrentMap<String, List<StoredLogEntry>> logsByService = new ConcurrentHashMap<>();
    private final ConcurrentMap<String, Set<String>> levelsByService = new ConcurrentHashMap<>();

    @Override
    public void storeLogs(List<StoredLogEntry> logs) {
        for (StoredLogEntry log : logs) {
            String service = log.getService();
            String level = log.getLevel().toUpperCase();

            logsByService.computeIfAbsent(service, k -> new CopyOnWriteArrayList<>()).add(log);
            levelsByService.computeIfAbsent(service, k -> ConcurrentHashMap.newKeySet()).add(level);
        }
        logger.debug("Stored {} log entries", logs.size());
    }

    @Override
    public List<StoredLogEntry> queryLogs(String service, List<String> levels, Long startTimeUtcMs, Long endTimeUtcMs, String keyword) {
        List<StoredLogEntry> result = new ArrayList<>();

        List<String> normalizedLevels = null;
        if (levels != null && !levels.isEmpty()) {
            normalizedLevels = levels.stream().map(String::toUpperCase).collect(Collectors.toList());
        }

        if (service != null) {
            List<StoredLogEntry> serviceLogs = logsByService.get(service);
            if (serviceLogs != null) {
                filterLogs(serviceLogs, normalizedLevels, startTimeUtcMs, endTimeUtcMs, keyword, result);
            }
        } else {
            for (Map.Entry<String, List<StoredLogEntry>> entry : logsByService.entrySet()) {
                filterLogs(entry.getValue(), normalizedLevels, startTimeUtcMs, endTimeUtcMs, keyword, result);
            }
        }

        result.sort(Comparator.comparingLong(StoredLogEntry::getTimestampUtcMs).reversed());
        return result;
    }

    @Override
    public void deleteLogsByServiceBefore(String service, long beforeUtcMs) {
        List<StoredLogEntry> serviceLogs = logsByService.get(service);
        if (serviceLogs == null) {
            return;
        }

        List<StoredLogEntry> remaining = serviceLogs.stream()
                .filter(log -> log.getTimestampUtcMs() >= beforeUtcMs)
                .collect(Collectors.toList());

        logsByService.put(service, new CopyOnWriteArrayList<>(remaining));
        logger.info("Deleted logs for service {} before timestamp {}, remaining: {}", service, beforeUtcMs, remaining.size());
    }

    @Override
    public Set<String> getAllServices() {
        return new HashSet<>(logsByService.keySet());
    }

    private void filterLogs(List<StoredLogEntry> logs, List<String> levels, Long startTimeUtcMs, Long endTimeUtcMs, String keyword, List<StoredLogEntry> result) {
        for (StoredLogEntry log : logs) {
            if (levels != null && !levels.contains(log.getLevel().toUpperCase())) {
                continue;
            }
            if (startTimeUtcMs != null && log.getTimestampUtcMs() < startTimeUtcMs) {
                continue;
            }
            if (endTimeUtcMs != null && log.getTimestampUtcMs() > endTimeUtcMs) {
                continue;
            }
            if (keyword != null && !keyword.isBlank() && !log.getMessage().toLowerCase().contains(keyword.toLowerCase())) {
                continue;
            }
            result.add(log);
        }
    }
}
