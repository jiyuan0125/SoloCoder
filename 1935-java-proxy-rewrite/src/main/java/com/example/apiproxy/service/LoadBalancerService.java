package com.example.apiproxy.service;

import com.example.apiproxy.config.ProxyProperties;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicInteger;

@Service
public class LoadBalancerService {

    private static final Logger logger = LoggerFactory.getLogger(LoadBalancerService.class);

    private final ProxyProperties properties;
    private final ConcurrentHashMap<String, AtomicInteger> counters = new ConcurrentHashMap<>();

    public LoadBalancerService(ProxyProperties properties) {
        this.properties = properties;
    }

    public String selectBackend(String version) {
        List<String> backends = properties.getBackends(version);
        if (backends.isEmpty()) {
            backends = properties.getBackends(properties.getDefaultVersion());
            logger.warn("No backends configured for version {}, using default version {}", version, properties.getDefaultVersion());
        }
        if (backends.isEmpty()) {
            throw new IllegalStateException("No backend servers configured");
        }
        AtomicInteger counter = counters.computeIfAbsent(version, k -> new AtomicInteger(0));
        int index = Math.abs(counter.getAndIncrement()) % backends.size();
        String selected = backends.get(index);
        logger.debug("Selected backend {} for version {} (round-robin index {})", selected, version, index);
        return selected;
    }

    public String selectBackendExcluding(String version, String excludeUrl) {
        List<String> backends = properties.getBackends(version);
        if (backends.isEmpty()) {
            backends = properties.getBackends(properties.getDefaultVersion());
        }
        if (backends.size() <= 1) {
            return backends.isEmpty() ? null : backends.get(0);
        }
        AtomicInteger counter = counters.computeIfAbsent(version, k -> new AtomicInteger(0));
        for (int i = 0; i < backends.size(); i++) {
            int index = Math.abs(counter.getAndIncrement()) % backends.size();
            String selected = backends.get(index);
            if (!selected.equals(excludeUrl)) {
                logger.debug("Selected alternative backend {} for version {} (excluded: {})", selected, version, excludeUrl);
                return selected;
            }
        }
        return backends.get(0);
    }

    public List<String> getBackends(String version) {
        List<String> backends = properties.getBackends(version);
        if (backends.isEmpty()) {
            backends = properties.getBackends(properties.getDefaultVersion());
        }
        return backends;
    }
}
