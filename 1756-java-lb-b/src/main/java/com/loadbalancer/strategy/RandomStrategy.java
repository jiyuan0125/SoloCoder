package com.loadbalancer.strategy;

import com.loadbalancer.model.Instance;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.Random;

@Component
public class RandomStrategy implements LoadBalanceStrategy {
    private final Random random = new Random();

    @Override
    public Instance select(List<Instance> availableInstances) {
        if (availableInstances == null || availableInstances.isEmpty()) {
            return null;
        }
        int index = random.nextInt(availableInstances.size());
        return availableInstances.get(index);
    }

    @Override
    public String getName() {
        return "RANDOM";
    }
}
