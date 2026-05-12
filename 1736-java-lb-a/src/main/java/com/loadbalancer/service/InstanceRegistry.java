package com.loadbalancer.service;

import com.loadbalancer.model.BackendInstance;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.time.Duration;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Component
public class InstanceRegistry {

    @Value("${loadbalancer.warmup-seconds:30}")
    private int warmupSeconds;

    private final Map<String, BackendInstance> instances = new ConcurrentHashMap<>();

    public void register(BackendInstance instance) {
        instances.put(instance.getId(), instance);
    }

    public void deregister(String instanceId) {
        instances.remove(instanceId);
    }

    public List<BackendInstance> getAllInstances() {
        return List.copyOf(instances.values());
    }

    public List<BackendInstance> getHealthyInstances() {
        return instances.values().stream()
                .filter(BackendInstance::isHealthy)
                .collect(Collectors.toList());
    }

    public BackendInstance getInstance(String instanceId) {
        return instances.get(instanceId);
    }

    public int getEffectiveWeight(BackendInstance instance) {
        if (instance.getWeight() <= 0) {
            return 0;
        }

        LocalDateTime registeredAt = instance.getRegisteredAt();
        if (registeredAt == null) {
            return instance.getWeight();
        }

        Duration elapsed = Duration.between(registeredAt, LocalDateTime.now());
        long elapsedSeconds = elapsed.getSeconds();

        if (elapsedSeconds >= warmupSeconds) {
            return instance.getWeight();
        }

        double ratio = (double) elapsedSeconds / warmupSeconds;
        return (int) Math.round(instance.getWeight() * ratio);
    }

    public void incrementConnection(String instanceId) {
        BackendInstance instance = instances.get(instanceId);
        if (instance != null) {
            instance.setActiveConnections(instance.getActiveConnections() + 1);
        }
    }

    public void decrementConnection(String instanceId) {
        BackendInstance instance = instances.get(instanceId);
        if (instance != null && instance.getActiveConnections() > 0) {
            instance.setActiveConnections(instance.getActiveConnections() - 1);
        }
    }

    public boolean isEmpty() {
        return instances.isEmpty();
    }
}
