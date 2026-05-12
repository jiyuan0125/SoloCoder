package com.loadbalancer.service;

import com.loadbalancer.config.LoadBalancerConfig;
import com.loadbalancer.model.Instance;
import com.loadbalancer.registry.InstanceRegistry;
import jakarta.annotation.PostConstruct;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestClient;
import org.springframework.web.client.RestClientException;

@Service
public class HealthCheckService {
    private static final Logger logger = LoggerFactory.getLogger(HealthCheckService.class);

    private final InstanceRegistry instanceRegistry;
    private final LoadBalancerConfig config;
    private final RestClient restClient;

    public HealthCheckService(InstanceRegistry instanceRegistry, LoadBalancerConfig config) {
        this.instanceRegistry = instanceRegistry;
        this.config = config;
        this.restClient = RestClient.create();
    }

    @PostConstruct
    public void init() {
        logger.info("HealthCheckService initialized with interval: {}ms", config.getHealthCheck().getInterval());
    }

    @Scheduled(fixedRateString = "${loadbalancer.health-check.interval:10000}")
    public void checkAllInstances() {
        for (Instance instance : instanceRegistry.getAllInstances()) {
            checkInstanceHealth(instance);
        }
    }

    public void checkInstanceHealth(Instance instance) {
        String healthUrl = instance.getUrl() + config.getHealthCheck().getPath();
        long timeout = config.getHealthCheck().getTimeout();
        boolean isHealthy = false;

        try {
            ResponseEntity<Void> response = restClient.get()
                    .uri(healthUrl)
                    .retrieve()
                    .toBodilessEntity();
            
            isHealthy = response.getStatusCode().is2xxSuccessful();
            logger.debug("Health check for {}: {} - {}", instance.getId(), 
                    response.getStatusCode(), isHealthy ? "HEALTHY" : "UNHEALTHY");
        } catch (RestClientException e) {
            logger.warn("Health check failed for {}: {}", instance.getId(), e.getMessage());
            isHealthy = false;
        }

        instance.setHealthy(isHealthy);
    }

    public boolean isInstanceHealthy(Instance instance) {
        return instance.isHealthy();
    }
}
