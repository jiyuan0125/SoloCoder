package com.breaker.core;

import com.breaker.model.BreakerConfig;
import com.breaker.model.BreakerDetail;
import com.breaker.model.BreakerState;
import com.breaker.model.BreakerSummary;
import com.breaker.model.CallRecord;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.function.Consumer;

@Component
public class BreakerManager {
    private final Map<String, CircuitBreaker> breakers = new ConcurrentHashMap<>();
    private Consumer<CircuitBreaker> stateChangeListener;

    public void setStateChangeListener(Consumer<CircuitBreaker> listener) {
        this.stateChangeListener = listener;
        for (CircuitBreaker breaker : breakers.values()) {
            breaker.setStateChangeListener(listener);
        }
    }

    public void createOrUpdateBreaker(String serviceName, BreakerConfig config) {
        breakers.compute(serviceName, (name, existing) -> {
            if (existing == null) {
                CircuitBreaker newBreaker = new CircuitBreaker(serviceName, config);
                if (stateChangeListener != null) {
                    newBreaker.setStateChangeListener(stateChangeListener);
                }
                return newBreaker;
            } else {
                existing.updateConfig(config);
                return existing;
            }
        });
    }

    public Optional<CircuitBreaker> getBreaker(String serviceName) {
        return Optional.ofNullable(breakers.get(serviceName));
    }

    public List<BreakerSummary> getAllBreakers() {
        List<BreakerSummary> summaries = new ArrayList<>();
        for (Map.Entry<String, CircuitBreaker> entry : breakers.entrySet()) {
            CircuitBreaker breaker = entry.getValue();
            summaries.add(new BreakerSummary(
                    breaker.getServiceName(),
                    breaker.getState(),
                    breaker.getConfig()
            ));
        }
        return summaries;
    }

    public Optional<BreakerDetail> getBreakerDetail(String serviceName) {
        CircuitBreaker breaker = breakers.get(serviceName);
        if (breaker == null) {
            return Optional.empty();
        }
        BreakerDetail detail = new BreakerDetail();
        detail.setService(breaker.getServiceName());
        detail.setState(breaker.getState());
        detail.setConfig(breaker.getConfig());
        detail.setRecentCalls(new ArrayList<>(breaker.getRecentCalls()));
        detail.setLastStateChange(breaker.getLastStateChange());
        return Optional.of(detail);
    }
}
