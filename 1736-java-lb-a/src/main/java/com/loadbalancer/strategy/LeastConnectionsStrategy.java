package com.loadbalancer.strategy;

import com.loadbalancer.model.BackendInstance;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.Random;

@Component
public class LeastConnectionsStrategy implements LoadBalancerStrategy {

    private final Random random = new Random();

    @Override
    public BackendInstance selectInstance(List<BackendInstance> instances) {
        if (instances == null || instances.isEmpty()) {
            return null;
        }

        List<BackendInstance> healthyInstances = instances.stream()
                .filter(BackendInstance::isHealthy)
                .toList();

        if (healthyInstances.isEmpty()) {
            return null;
        }

        int minConnections = healthyInstances.stream()
                .mapToInt(BackendInstance::getActiveConnections)
                .min()
                .orElse(0);

        List<BackendInstance> candidates = healthyInstances.stream()
                .filter(instance -> instance.getActiveConnections() == minConnections)
                .toList();

        if (candidates.isEmpty()) {
            return null;
        }

        if (candidates.size() == 1) {
            return candidates.get(0);
        }

        return candidates.get(random.nextInt(candidates.size()));
    }
}
