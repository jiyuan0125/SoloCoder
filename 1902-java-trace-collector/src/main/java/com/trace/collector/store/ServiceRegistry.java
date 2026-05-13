package com.trace.collector.store;

import com.trace.collector.model.ServiceRegistration;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

public class ServiceRegistry {

    private static final Logger logger = LoggerFactory.getLogger(ServiceRegistry.class);

    private final Map<String, ServiceRegistration> services = new ConcurrentHashMap<>();

    public void register(ServiceRegistration registration) {
        services.put(registration.getName(), registration);
        logger.info("Service registered: name={}, address={}",
                registration.getName(), registration.getAddress());
    }

    public boolean isRegistered(String serviceName) {
        return services.containsKey(serviceName);
    }

    public ServiceRegistration getService(String name) {
        return services.get(name);
    }

    public Map<String, ServiceRegistration> getAllServices() {
        return new ConcurrentHashMap<>(services);
    }
}
