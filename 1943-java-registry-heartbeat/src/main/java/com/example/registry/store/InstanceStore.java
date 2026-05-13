package com.example.registry.store;

import com.example.registry.model.ServiceInstance;
import org.springframework.stereotype.Component;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Component
public class InstanceStore {
    private final Map<String, ServiceInstance> instancesById = new ConcurrentHashMap<>();
    private final Map<String, Set<String>> instanceIdsByService = new ConcurrentHashMap<>();

    public ServiceInstance save(ServiceInstance instance) {
        instancesById.put(instance.getId(), instance);
        instanceIdsByService.computeIfAbsent(instance.getServiceName(), k -> ConcurrentHashMap.newKeySet())
                .add(instance.getId());
        return instance;
    }

    public Optional<ServiceInstance> findById(String id) {
        return Optional.ofNullable(instancesById.get(id));
    }

    public List<ServiceInstance> findByServiceName(String serviceName) {
        Set<String> ids = instanceIdsByService.get(serviceName);
        if (ids == null || ids.isEmpty()) {
            return Collections.emptyList();
        }
        return ids.stream()
                .map(instancesById::get)
                .filter(Objects::nonNull)
                .collect(Collectors.toList());
    }

    public List<ServiceInstance> findAll() {
        return new ArrayList<>(instancesById.values());
    }

    public Set<String> getAllServiceNames() {
        return new HashSet<>(instanceIdsByService.keySet());
    }

    public int countByServiceName(String serviceName) {
        Set<String> ids = instanceIdsByService.get(serviceName);
        return ids == null ? 0 : ids.size();
    }

    public void remove(String id) {
        ServiceInstance instance = instancesById.remove(id);
        if (instance != null) {
            Set<String> ids = instanceIdsByService.get(instance.getServiceName());
            if (ids != null) {
                ids.remove(id);
                if (ids.isEmpty()) {
                    instanceIdsByService.remove(instance.getServiceName());
                }
            }
        }
    }
}
