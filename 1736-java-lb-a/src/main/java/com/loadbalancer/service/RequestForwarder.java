package com.loadbalancer.service;

import com.loadbalancer.model.BackendInstance;
import org.springframework.http.*;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

import java.net.URI;

@Service
public class RequestForwarder {

    private final RestTemplate restTemplate = new RestTemplate();
    private final InstanceRegistry instanceRegistry;
    private final StrategyManager strategyManager;

    public RequestForwarder(InstanceRegistry instanceRegistry, StrategyManager strategyManager) {
        this.instanceRegistry = instanceRegistry;
        this.strategyManager = strategyManager;
    }

    public ResponseEntity<byte[]> forwardRequest(String path, HttpMethod method, HttpHeaders headers, byte[] body) {
        if (instanceRegistry.isEmpty()) {
            return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE)
                    .body("No backend instances available".getBytes());
        }

        BackendInstance instance = strategyManager.getCurrentStrategyImplementation()
                .selectInstance(instanceRegistry.getHealthyInstances());

        if (instance == null) {
            return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE)
                    .body("No available backend instances".getBytes());
        }

        String instanceId = instance.getId();
        try {
            instanceRegistry.incrementConnection(instanceId);

            String targetUrl = String.format("http://%s:%d%s",
                    instance.getHost(),
                    instance.getPort(),
                    path);

            HttpEntity<byte[]> requestEntity = new HttpEntity<>(body, headers);

            return restTemplate.exchange(
                    URI.create(targetUrl),
                    method,
                    requestEntity,
                    byte[].class
            );
        } finally {
            instanceRegistry.decrementConnection(instanceId);
        }
    }
}
