package com.loadbalancer.service;

import com.loadbalancer.model.BackendNode;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicInteger;

@Service
public class WeightedRoundRobinBalancer {
    private final ConcurrentHashMap<String, AtomicInteger> currentWeights = new ConcurrentHashMap<>();

    public BackendNode select(List<BackendNode> availableNodes) {
        if (availableNodes == null || availableNodes.isEmpty()) {
            return null;
        }
        if (availableNodes.size() == 1) {
            return availableNodes.get(0);
        }

        int totalWeight = availableNodes.stream().mapToInt(BackendNode::getWeight).sum();

        while (true) {
            for (BackendNode node : availableNodes) {
                AtomicInteger current = currentWeights.computeIfAbsent(node.getId(), k -> new AtomicInteger(0));
                int newWeight = current.addAndGet(node.getWeight());

                if (newWeight >= totalWeight) {
                    current.set(newWeight - totalWeight);
                    return node;
                }
            }
        }
    }

    public void resetState() {
        currentWeights.clear();
    }
}
