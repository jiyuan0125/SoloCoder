package com.healthcheck.service;

import com.healthcheck.model.ServiceConfig;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class ServiceConfigStore {
    
    private final Map<String, ServiceConfig> configs = new ConcurrentHashMap<>();
    
    public void registerService(ServiceConfig config) {
        configs.put(config.getServiceId(), config);
    }
    
    public void unregisterService(String serviceId) {
        configs.remove(serviceId);
    }
    
    public Optional<ServiceConfig> getServiceConfig(String serviceId) {
        return Optional.ofNullable(configs.get(serviceId));
    }
    
    public List<ServiceConfig> getAllServices() {
        return new ArrayList<>(configs.values());
    }
    
    public boolean exists(String serviceId) {
        return configs.containsKey(serviceId);
    }
}
