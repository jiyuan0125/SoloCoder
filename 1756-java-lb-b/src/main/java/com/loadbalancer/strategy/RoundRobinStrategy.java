package com.loadbalancer.strategy;

import com.loadbalancer.model.Instance;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.concurrent.atomic.AtomicInteger;

@Component
public class RoundRobinStrategy implements LoadBalanceStrategy {
    private final AtomicInteger counter = new AtomicInteger(0);

    @Override
    public Instance select(List<Instance> availableInstances) {
        if (availableInstances == null || availableInstances.isEmpty()) {
            return null;
        }
        int index = Math.abs(counter.getAndIncrement()) % availableInstances.size();
        return availableInstances.get(index);
    }

    @Override
    public String getName() {
        return "ROUND_ROBIN";
    }
}
