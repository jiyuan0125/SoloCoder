package com.example.apiproxymirror.service;

import com.example.apiproxymirror.config.RecordingConfig;
import com.example.apiproxymirror.model.RecordedExchange;
import com.example.apiproxymirror.model.RecordingSettings;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentLinkedDeque;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
public class RecordingService {

    private final RecordingConfig recordingConfig;

    private final Map<String, RecordedExchange> recordsById = new ConcurrentHashMap<>();
    private final Deque<String> recordOrder = new ConcurrentLinkedDeque<>();

    private volatile RecordingSettings settings = new RecordingSettings();

    public void startRecording(RecordingSettings newSettings) {
        this.settings = newSettings;
        this.settings.setEnabled(true);
    }

    public void stopRecording() {
        this.settings.setEnabled(false);
    }

    public RecordingSettings getSettings() {
        return settings;
    }

    public boolean isRecordingEnabled() {
        return settings.isEnabled();
    }

    public boolean shouldRecord(String path) {
        if (!settings.isEnabled()) {
            return false;
        }

        List<String> excludePaths = settings.getExcludePaths();
        if (excludePaths != null && !excludePaths.isEmpty()) {
            for (String exclude : excludePaths) {
                if (pathMatches(path, exclude)) {
                    return false;
                }
            }
        }

        List<String> includePaths = settings.getIncludePaths();
        if (includePaths == null || includePaths.isEmpty()) {
            return true;
        }

        for (String include : includePaths) {
            if (pathMatches(path, include)) {
                return true;
            }
        }

        return false;
    }

    private boolean pathMatches(String path, String pattern) {
        if (pattern.endsWith("*")) {
            return path.startsWith(pattern.substring(0, pattern.length() - 1));
        }
        return path.equals(pattern);
    }

    public void record(RecordedExchange exchange) {
        String id = UUID.randomUUID().toString();
        exchange.setId(id);
        exchange.setTimestamp(Instant.now());

        recordsById.put(id, exchange);
        recordOrder.addLast(id);

        while (recordOrder.size() > recordingConfig.getMaxRecords()) {
            String oldestId = recordOrder.pollFirst();
            if (oldestId != null) {
                recordsById.remove(oldestId);
            }
        }
    }

    public Optional<RecordedExchange> getById(String id) {
        return Optional.ofNullable(recordsById.get(id));
    }

    public List<RecordedExchange> list(String path, Integer status) {
        return recordOrder.stream()
                .map(recordsById::get)
                .filter(Objects::nonNull)
                .filter(record -> path == null || record.getPath().equals(path))
                .filter(record -> status == null || record.getResponseStatus() == status)
                .collect(Collectors.toList());
    }

    public Optional<RecordedExchange> getLatest(String method, String path) {
        Iterator<String> descendingIterator = recordOrder.descendingIterator();
        while (descendingIterator.hasNext()) {
            String id = descendingIterator.next();
            RecordedExchange record = recordsById.get(id);
            if (record != null) {
                boolean methodMatch = method == null || record.getMethod().equalsIgnoreCase(method);
                boolean pathMatch = path == null || record.getPath().equals(path);
                if (methodMatch && pathMatch) {
                    return Optional.of(record);
                }
            }
        }
        return Optional.empty();
    }

    public int getRecordCount() {
        return recordsById.size();
    }

    public void clearAllRecords() {
        recordsById.clear();
        recordOrder.clear();
    }
}
