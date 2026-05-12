package com.loadbalancer.strategy;

import com.loadbalancer.model.BackendInstance;
import com.loadbalancer.service.InstanceRegistry;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.concurrent.atomic.AtomicInteger;

@Component
public class WeightedRoundRobinStrategy implements LoadBalancerStrategy {

    private final InstanceRegistry instanceRegistry;
    private final AtomicInteger currentIndex = new AtomicInteger(0);
    private final AtomicInteger currentWeight = new AtomicInteger(0);

    public WeightedRoundRobinStrategy(InstanceRegistry instanceRegistry) {
        this.instanceRegistry = instanceRegistry;
    }

    @Override
    public BackendInstance selectInstance(List<BackendInstance> instances) {
        if (instances == null || instances.isEmpty()) {
            return null;
        }

        List<BackendInstance> eligibleInstances = instances.stream()
                .filter(instance -> instanceRegistry.getEffectiveWeight(instance) > 0)
                .toList();

        if (eligibleInstances.isEmpty()) {
            return null;
        }

        int maxWeight = eligibleInstances.stream()
                .mapToInt(instanceRegistry::getEffectiveWeight)
                .max()
                .orElse(1);

        int gcdWeight = computeGCD(eligibleInstances);

        while (true) {
            int index = currentIndex.get() % eligibleInstances.size();
            BackendInstance instance = eligibleInstances.get(index);
            int effectiveWeight = instanceRegistry.getEffectiveWeight(instance);

            if (currentWeight.addAndGet(-gcdWeight) <= 0) {
                currentWeight.set(maxWeight);
            }

            if (effectiveWeight >= currentWeight.get()) {
                currentIndex.set((index + 1) % eligibleInstances.size());
                return instance;
            }

            currentIndex.set((index + 1) % eligibleInstances.size());
        }
    }

    private int computeGCD(List<BackendInstance> instances) {
        int gcd = 0;
        for (BackendInstance instance : instances) {
            int weight = instanceRegistry.getEffectiveWeight(instance);
            if (weight > 0) {
                gcd = gcd(gcd, weight);
            }
        }
        return gcd > 0 ? gcd : 1;
    }

    private int gcd(int a, int b) {
        while (b != 0) {
            int temp = b;
            b = a % b;
            a = temp;
        }
        return a;
    }
}
