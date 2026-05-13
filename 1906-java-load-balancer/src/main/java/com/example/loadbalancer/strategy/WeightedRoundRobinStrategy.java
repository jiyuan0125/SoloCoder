package com.example.loadbalancer.strategy;

import com.example.loadbalancer.model.Node;
import com.example.loadbalancer.registry.NodeRegistry;
import org.springframework.stereotype.Component;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Component
public class WeightedRoundRobinStrategy implements LoadBalanceStrategy {

    private final NodeRegistry nodeRegistry;
    private final Map<String, Integer> currentWeights = new ConcurrentHashMap<>();

    public WeightedRoundRobinStrategy(NodeRegistry nodeRegistry) {
        this.nodeRegistry = nodeRegistry;
    }

    @Override
    public String getName() {
        return "weighted_round_robin";
    }

    @Override
    public synchronized Node select() {
        List<Node> available = nodeRegistry.getAvailableNodes();
        if (available.isEmpty()) {
            return null;
        }

        int totalWeight = available.stream()
                .mapToInt(Node::getCurrentWeight)
                .sum();

        for (Node node : available) {
            String nodeId = node.getId();
            int curWeight = currentWeights.getOrDefault(nodeId, 0) + node.getCurrentWeight();
            currentWeights.put(nodeId, curWeight);
        }

        Node selected = null;
        int maxWeight = Integer.MIN_VALUE;
        
        for (Node node : available) {
            String nodeId = node.getId();
            int curWeight = currentWeights.getOrDefault(nodeId, 0);
            if (curWeight > maxWeight) {
                maxWeight = curWeight;
                selected = node;
            }
        }

        if (selected != null) {
            String selectedId = selected.getId();
            int newWeight = currentWeights.getOrDefault(selectedId, 0) - totalWeight;
            currentWeights.put(selectedId, newWeight);
        }

        return selected;
    }
}
