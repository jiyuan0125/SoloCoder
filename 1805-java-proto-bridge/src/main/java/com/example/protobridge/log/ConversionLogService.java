package com.example.protobridge.log;

import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.*;
import java.util.concurrent.ConcurrentLinkedDeque;

@Service
public class ConversionLogService {

    private static final int MAX_LOGS = 1000;
    private final Deque<ConversionLog> logs = new ConcurrentLinkedDeque<>();

    public ConversionLog recordSuccess(String direction, String path, String method, Long durationMs, String ruleId) {
        ConversionLog log = ConversionLog.builder()
                .id(UUID.randomUUID().toString())
                .direction(direction)
                .path(path)
                .method(method)
                .durationMs(durationMs)
                .success(true)
                .errorReason(null)
                .timestamp(LocalDateTime.now())
                .ruleId(ruleId)
                .build();
        addLog(log);
        return log;
    }

    public ConversionLog recordFailure(String direction, String path, String method, Long durationMs, String errorReason, String ruleId) {
        ConversionLog log = ConversionLog.builder()
                .id(UUID.randomUUID().toString())
                .direction(direction)
                .path(path)
                .method(method)
                .durationMs(durationMs)
                .success(false)
                .errorReason(errorReason)
                .timestamp(LocalDateTime.now())
                .ruleId(ruleId)
                .build();
        addLog(log);
        return log;
    }

    private void addLog(ConversionLog log) {
        logs.addLast(log);
        while (logs.size() > MAX_LOGS) {
            logs.pollFirst();
        }
    }

    public List<ConversionLog> getLogs(Integer limit) {
        int actualLimit = limit != null && limit > 0 ? Math.min(limit, MAX_LOGS) : MAX_LOGS;
        List<ConversionLog> result = new ArrayList<>(logs);
        Collections.reverse(result);
        return result.subList(0, Math.min(actualLimit, result.size()));
    }

    public List<ConversionLog> getLogsByDirection(String direction, Integer limit) {
        int actualLimit = limit != null && limit > 0 ? Math.min(limit, MAX_LOGS) : MAX_LOGS;
        List<ConversionLog> result = new ArrayList<>();
        Iterator<ConversionLog> it = logs.descendingIterator();
        while (it.hasNext() && result.size() < actualLimit) {
            ConversionLog log = it.next();
            if (direction == null || direction.equalsIgnoreCase(log.getDirection())) {
                result.add(log);
            }
        }
        return result;
    }

    public void clearLogs() {
        logs.clear();
    }

    public int getLogCount() {
        return logs.size();
    }
}
