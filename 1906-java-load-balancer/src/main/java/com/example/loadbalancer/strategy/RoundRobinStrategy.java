package com.example.loadbalancer.strategy;

import com.example.loadbalancer.model.Node;
import com.example.loadbalancer.registry.NodeRegistry;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.concurrent.atomic.AtomicInteger;

@Component
public class RoundRobinStrategy implements LoadBalanceStrategy {

    private final NodeRegistry nodeRegistry;
    private final AtomicInteger counter = new AtomicInteger(0);

    public RoundRobinStrategy(NodeRegistry nodeRegistry) {
        this.nodeRegistry = nodeRegistry;
    }

    @Override
    public String getName() {
        return "round_robin";
    }

    @Override
    public Node select() {
        List<Node> available = nodeRegistry.getAvailableNodes();
        if (available.isEmpty()) {
            return null;
        }
        int index = Math.abs(counter.getAndIncrement() % available.size());
        return available.get(index);
    }
}
