package com.example.loadbalancer.strategy;

import com.example.loadbalancer.model.Node;
import com.example.loadbalancer.registry.NodeRegistry;
import org.springframework.stereotype.Component;

import java.util.Comparator;
import java.util.List;

@Component
public class LeastConnectionStrategy implements LoadBalanceStrategy {

    private final NodeRegistry nodeRegistry;

    public LeastConnectionStrategy(NodeRegistry nodeRegistry) {
        this.nodeRegistry = nodeRegistry;
    }

    @Override
    public String getName() {
        return "least_conn";
    }

    @Override
    public Node select() {
        List<Node> available = nodeRegistry.getAvailableNodes();
        if (available.isEmpty()) {
            return null;
        }

        return available.stream()
                .min(Comparator.comparingInt(Node::getActiveConnectionsCount)
                        .thenComparingLong(Node::getRegisteredAt))
                .orElse(null);
    }
}
