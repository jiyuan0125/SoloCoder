package com.example.healthchecker.repository;

import com.example.healthchecker.model.ServiceRegistry;
import org.springframework.stereotype.Repository;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

@Repository
public class ServiceRegistryRepository {

    private final Map<String, ServiceRegistry> services = new ConcurrentHashMap<>();

    public ServiceRegistry getOrCreate(String name) {
        return services.computeIfAbsent(name, n -> {
            ServiceRegistry registry = new ServiceRegistry();
            registry.setName(n);
            return registry;
        });
    }

    public Optional<ServiceRegistry> findByName(String name) {
        return Optional.ofNullable(services.get(name));
    }

    public Collection<ServiceRegistry> findAll() {
        return new ArrayList<>(services.values());
    }

    public Set<String> findDependentServices(String dependencyName) {
        Set<String> dependents = new HashSet<>();
        for (ServiceRegistry service : services.values()) {
            if (service.getDependencies().contains(dependencyName)) {
                dependents.add(service.getName());
            }
        }
        return dependents;
    }
}
