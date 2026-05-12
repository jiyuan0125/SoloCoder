package com.example.circuitbreaker.stats;

import com.example.circuitbreaker.breaker.CircuitBreakerState;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.ArrayList;
import java.util.Collections;
import java.util.LinkedList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class StatsService {

    private static final int MAX_CALL_RECORDS = 10;
    private static final int MAX_STATE_CHANGE_RECORDS = 100;

    private final Map<String, LinkedList<CallRecord>> callRecords = new ConcurrentHashMap<>();
    private final Map<String, LinkedList<StateChangeEvent>> stateChangeEvents = new ConcurrentHashMap<>();

    public void recordCall(String serviceName, String requestId, boolean success) {
        CallRecord record = CallRecord.builder()
            .serviceName(serviceName)
            .requestId(requestId)
            .success(success)
            .timestamp(Instant.now())
            .build();

        callRecords.computeIfAbsent(serviceName, k -> new LinkedList<>()).add(record);
        trimList(callRecords.get(serviceName), MAX_CALL_RECORDS);
    }

    public void recordStateChange(String serviceName, CircuitBreakerState oldState,
                                  CircuitBreakerState newState, ChangeReason reason, String requestId) {
        StateChangeEvent event = StateChangeEvent.builder()
            .serviceName(serviceName)
            .oldState(oldState.name())
            .newState(newState.name())
            .reason(reason)
            .requestId(requestId)
            .timestamp(Instant.now())
            .build();

        stateChangeEvents.computeIfAbsent(serviceName, k -> new LinkedList<>()).add(event);
        trimList(stateChangeEvents.get(serviceName), MAX_STATE_CHANGE_RECORDS);
    }

    public List<CallRecord> getRecentCalls(String serviceName) {
        LinkedList<CallRecord> records = callRecords.get(serviceName);
        if (records == null || records.isEmpty()) {
            return Collections.emptyList();
        }
        return new ArrayList<>(records);
    }

    public List<StateChangeEvent> getStateChangeEvents(String serviceName) {
        LinkedList<StateChangeEvent> events = stateChangeEvents.get(serviceName);
        if (events == null || events.isEmpty()) {
            return Collections.emptyList();
        }
        return new ArrayList<>(events);
    }

    public void removeServiceRecords(String serviceName) {
        callRecords.remove(serviceName);
        stateChangeEvents.remove(serviceName);
    }

    private <T> void trimList(LinkedList<T> list, int maxSize) {
        while (list.size() > maxSize) {
            list.removeFirst();
        }
    }
}
