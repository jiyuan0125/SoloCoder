package com.loadbalancer.service;

import com.loadbalancer.model.BackendNode;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Optional;

@Service
public class WeightedRoundRobinService {
    private final NodeRegistryService nodeRegistryService;

    public WeightedRoundRobinService(NodeRegistryService nodeRegistryService) {
        this.nodeRegistryService = nodeRegistryService;
    }

    public Optional<BackendNode> selectNode() {
        List<BackendNode> healthyNodes = nodeRegistryService.getHealthyNodes();
        return selectNodeFromCandidates(healthyNodes);
    }

    public Optional<BackendNode> selectNodeWithTags(java.util.Set<String> tags) {
        List<BackendNode> candidateNodes = nodeRegistryService.getHealthyNodesWithTags(tags);
        return selectNodeFromCandidates(candidateNodes);
    }

    private Optional<BackendNode> selectNodeFromCandidates(List<BackendNode> candidates) {
        if (candidates.isEmpty()) {
            return Optional.empty();
        }

        synchronized (this) {
            int totalWeight = candidates.stream()
                    .mapToInt(BackendNode::getWeight)
                    .sum();

            for (BackendNode node : candidates) {
                node.setCurrentWeight(node.getCurrentWeight() + node.getWeight());
            }

            BackendNode selectedNode = null;
            int maxWeight = Integer.MIN_VALUE;

            for (BackendNode node : candidates) {
                if (node.getCurrentWeight() > maxWeight) {
                    maxWeight = node.getCurrentWeight();
                    selectedNode = node;
                }
            }

            if (selectedNode != null) {
                selectedNode.setCurrentWeight(selectedNode.getCurrentWeight() - totalWeight);
            }

            return Optional.ofNullable(selectedNode);
        }
    }
}
