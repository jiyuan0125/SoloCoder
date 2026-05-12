package com.example.circuitbreaker.config;

import com.example.circuitbreaker.breaker.CircuitBreakerRegistry;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;

@Slf4j
@Service
@RequiredArgsConstructor
public class ConfigService {

    private final Map<String, CircuitBreakerConfig> configs = new ConcurrentHashMap<>();
    private final CircuitBreakerRegistry circuitBreakerRegistry;

    public CircuitBreakerConfig registerService(String serviceName) {
        return registerService(serviceName, null);
    }

    public CircuitBreakerConfig registerService(String serviceName, CircuitBreakerConfig config) {
        CircuitBreakerConfig effectiveConfig = (config != null) ? config : new CircuitBreakerConfig();
        configs.put(serviceName, effectiveConfig);
        log.info("Registered config for service: {} with threshold={}, duration={}s",
            serviceName, effectiveConfig.getFailureThreshold(), effectiveConfig.getOpenDurationSeconds());
        return effectiveConfig;
    }

    public CircuitBreakerConfig updateConfig(String serviceName, CircuitBreakerConfig config) {
        configs.put(serviceName, config);
        log.info("Updated config for service: {} with threshold={}, duration={}s",
            serviceName, config.getFailureThreshold(), config.getOpenDurationSeconds());
        return config;
    }

    public Optional<CircuitBreakerConfig> getConfig(String serviceName) {
        return Optional.ofNullable(configs.get(serviceName));
    }

    public boolean removeConfig(String serviceName) {
        configs.remove(serviceName);
        circuitBreakerRegistry.remove(serviceName);
        log.info("Removed config and breaker for service: {}", serviceName);
        return true;
    }

    public boolean exists(String serviceName) {
        return configs.containsKey(serviceName);
    }

    public Map<String, CircuitBreakerConfig> getAllConfigs() {
        return new ConcurrentHashMap<>(configs);
    }
}
