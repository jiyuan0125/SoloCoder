package com.loadbalancer.service;

import com.loadbalancer.model.BackendNode;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Service
public class NodeRegistryService {
    private final Map<String, BackendNode> nodes = new ConcurrentHashMap<>();

    public void registerNode(BackendNode node) {
        if (node.getId() == null || node.getId().isEmpty()) {
            throw new IllegalArgumentException("Node id cannot be empty");
        }
        if (node.getHost() == null || node.getHost().isEmpty()) {
            throw new IllegalArgumentException("Node host cannot be empty");
        }
        if (node.getPort() <= 0) {
            throw new IllegalArgumentException("Invalid port");
        }
        if (node.getWeight() <= 0) {
            node.setWeight(1);
        }
        nodes.put(node.getId(), node);
    }

    public void unregisterNode(String nodeId) {
        nodes.remove(nodeId);
    }

    public Optional<BackendNode> getNode(String nodeId) {
        return Optional.ofNullable(nodes.get(nodeId));
    }

    public List<BackendNode> getAllNodes() {
        return new ArrayList<>(nodes.values());
    }

    public List<BackendNode> getHealthyNodes() {
        return nodes.values().stream()
                .filter(BackendNode::isHealthy)
                .collect(Collectors.toList());
    }

    public List<BackendNode> getHealthyNodesWithTags(Set<String> tags) {
        return nodes.values().stream()
                .filter(BackendNode::isHealthy)
                .filter(node -> node.hasAllTags(tags))
                .collect(Collectors.toList());
    }

    public void updateWeight(String nodeId, int weight) {
        BackendNode node = nodes.get(nodeId);
        if (node == null) {
            throw new IllegalArgumentException("Node not found: " + nodeId);
        }
        if (weight <= 0) {
            throw new IllegalArgumentException("Weight must be positive");
        }
        node.setWeight(weight);
    }

    public void updateTags(String nodeId, Set<String> tags) {
        BackendNode node = nodes.get(nodeId);
        if (node == null) {
            throw new IllegalArgumentException("Node not found: " + nodeId);
        }
        node.setTags(tags != null ? tags : new HashSet<>());
    }

    public void markNodeUnhealthy(String nodeId) {
        BackendNode node = nodes.get(nodeId);
        if (node != null) {
            node.setHealthy(false);
            node.setCurrentWeight(0);
        }
    }

    public void markNodeHealthy(String nodeId) {
        BackendNode node = nodes.get(nodeId);
        if (node != null) {
            node.setHealthy(true);
        }
    }

    public void updateNodeHealth(String nodeId, boolean healthy) {
        if (healthy) {
            markNodeHealthy(nodeId);
        } else {
            markNodeUnhealthy(nodeId);
        }
    }
}
