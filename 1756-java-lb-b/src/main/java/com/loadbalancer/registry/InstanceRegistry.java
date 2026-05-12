package com.loadbalancer.registry;

import com.loadbalancer.config.LoadBalancerConfig;
import com.loadbalancer.model.Instance;
import org.springframework.stereotype.Component;

import java.time.Duration;
import java.time.LocalDateTime;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Component
public class InstanceRegistry {
    private final Map<String, Instance> instances = new ConcurrentHashMap<>();
    private final LoadBalancerConfig config;

    public InstanceRegistry(LoadBalancerConfig config) {
        this.config = config;
    }

    public void registerInstance(String id, String host, int port) {
        Instance instance = new Instance(id, host, port);
        instances.put(id, instance);
    }

    public void deregisterInstance(String id) {
        instances.remove(id);
    }

    public Instance getInstance(String id) {
        return instances.get(id);
    }

    public List<Instance> getAllInstances() {
        return new ArrayList<>(instances.values());
    }

    public List<Instance> getAvailableInstances() {
        LocalDateTime now = LocalDateTime.now();
        int warmupSeconds = config.getWarmupSeconds();

        return instances.values().stream()
                .filter(Instance::isOnline)
                .filter(Instance::isHealthy)
                .filter(instance -> {
                    if (warmupSeconds <= 0) {
                        return true;
                    }
                    return Duration.between(instance.getRegisteredAt(), now).getSeconds() >= warmupSeconds;
                })
                .collect(Collectors.toList());
    }

    public void markInstanceOffline(String id) {
        Instance instance = instances.get(id);
        if (instance != null) {
            instance.setOnline(false);
        }
    }

    public void markInstanceOnline(String id) {
        Instance instance = instances.get(id);
        if (instance != null) {
            instance.setOnline(true);
            instance.setRegisteredAt(LocalDateTime.now());
        }
    }

    public boolean isInstanceAvailable(String id) {
        Instance instance = instances.get(id);
        if (instance == null || !instance.isOnline() || !instance.isHealthy()) {
            return false;
        }

        LocalDateTime now = LocalDateTime.now();
        int warmupSeconds = config.getWarmupSeconds();
        if (warmupSeconds > 0) {
            return Duration.between(instance.getRegisteredAt(), now).getSeconds() >= warmupSeconds;
        }
        return true;
    }
}
