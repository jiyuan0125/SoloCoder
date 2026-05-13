package com.example.healthchecker.repository;

import com.example.healthchecker.model.ServiceRegistry;
import org.springframework.stereotype.Repository;

import java.util.*;

@Repository
public class ServiceRegistryRepository {

    private final Map<String, ServiceRegistry> services = new HashMap<>();
    private final Object lock = new Object();

    public ServiceRegistry getOrCreate(String name) {
        synchronized (lock) {
            ServiceRegistry registry = services.get(name);
            if (registry == null) {
                registry = new ServiceRegistry();
                registry.setName(name);
                services.put(name, registry);
            }
            return registry;
        }
    }

    public Optional<ServiceRegistry> findByName(String name) {
        synchronized (lock) {
            return Optional.ofNullable(services.get(name));
        }
    }

    public Collection<ServiceRegistry> findAll() {
        synchronized (lock) {
            return new ArrayList<>(services.values());
        }
    }

    public Set<String> findDependentServices(String dependencyName) {
        Set<String> dependents = new HashSet<>();
        synchronized (lock) {
            for (ServiceRegistry service : services.values()) {
                if (service.getDependencies().contains(dependencyName)) {
                    dependents.add(service.getName());
                }
            }
        }
        return dependents;
    }
}
