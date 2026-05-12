package com.health.store;

import com.health.config.HealthMonitorProperties;
import com.health.model.ServiceRegistration;
import org.springframework.stereotype.Component;

import java.time.LocalDateTime;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Component
public class ServiceRegistry {

    private final Map<String, ServiceRegistration> services = new ConcurrentHashMap<>();
    private final HealthMonitorProperties properties;

    public ServiceRegistry(HealthMonitorProperties properties) {
        this.properties = properties;
    }

    public ServiceRegistration register(ServiceRegistration service) {
        String id = service.getId();
        if (id == null || id.isEmpty()) {
            id = UUID.randomUUID().toString();
            service.setId(id);
        }

        applyDefaults(service);

        if (service.getCreatedAt() == null) {
            service.setCreatedAt(LocalDateTime.now());
        }

        services.put(id, service);
        return service;
    }

    private void applyDefaults(ServiceRegistration service) {
        if (service.getIntervalSeconds() == null) {
            service.setIntervalSeconds(properties.getDefaultIntervalSeconds());
        }
        if (service.getTimeoutMilliseconds() == null) {
            service.setTimeoutMilliseconds(properties.getDefaultTimeoutMilliseconds());
        }
        if (service.getUnhealthyThreshold() == null) {
            service.setUnhealthyThreshold(properties.getUnhealthyThreshold());
        }
        if (service.getConsecutiveFailures() == null) {
            service.setConsecutiveFailures(0);
        }
    }

    public ServiceRegistration getById(String id) {
        return services.get(id);
    }

    public List<ServiceRegistration> getAll() {
        return new ArrayList<>(services.values());
    }

    public boolean remove(String id) {
        return services.remove(id) != null;
    }

    public ServiceRegistration update(ServiceRegistration service) {
        if (services.containsKey(service.getId())) {
            return services.put(service.getId(), service);
        }
        return null;
    }

    public List<ServiceRegistration> findAllDueForCheck() {
        LocalDateTime now = LocalDateTime.now();
        return services.values().stream()
                .filter(service -> {
                    LocalDateTime lastChecked = service.getLastCheckedAt();
                    if (lastChecked == null) {
                        return true;
                    }
                    int interval = service.getIntervalSeconds() != null ? 
                            service.getIntervalSeconds() : properties.getDefaultIntervalSeconds();
                    return lastChecked.plusSeconds(interval).isBefore(now);
                })
                .collect(Collectors.toList());
    }
}