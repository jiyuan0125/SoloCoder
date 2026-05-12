package com.loadbalancer.service;

import com.loadbalancer.model.BackendNode;
import com.loadbalancer.model.BackendNode.NodeStatus;
import com.loadbalancer.dto.NodeRegistrationRequest;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.stream.Collectors;

@Service
public class NodeManager {
    private final Map<String, BackendNode> nodes = new ConcurrentHashMap<>();
    private final AtomicInteger nodeIdGenerator = new AtomicInteger(0);

    public BackendNode register(NodeRegistrationRequest request) {
        String nodeId = "node-" + nodeIdGenerator.incrementAndGet();
        BackendNode node = BackendNode.builder()
                .id(nodeId)
                .host(request.getHost())
                .port(request.getPort())
                .weight(request.getWeight() <= 0 ? 1 : request.getWeight())
                .tags(request.getTags() != null ? new ArrayList<>(request.getTags()) : new ArrayList<>())
                .status(NodeStatus.HEALTHY)
                .registeredAt(LocalDateTime.now())
                .statusChangedAt(LocalDateTime.now())
                .build();
        nodes.put(nodeId, node);
        return node;
    }

    public Optional<BackendNode> deregister(String nodeId) {
        BackendNode node = nodes.get(nodeId);
        if (node == null) {
            return Optional.empty();
        }
        node.setStatus(NodeStatus.DRAINING);
        node.setStatusChangedAt(LocalDateTime.now());
        nodes.put(nodeId, node);
        scheduleFinalRemoval(nodeId);
        return Optional.of(node);
    }

    private void scheduleFinalRemoval(String nodeId) {
        new Thread(() -> {
            try {
                Thread.sleep(60000);
                nodes.remove(nodeId);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        }).start();
    }

    public Optional<BackendNode> updateWeight(String nodeId, int weight) {
        BackendNode node = nodes.get(nodeId);
        if (node == null) {
            return Optional.empty();
        }
        node.setWeight(weight <= 0 ? 1 : weight);
        nodes.put(nodeId, node);
        return Optional.of(node);
    }

    public Optional<BackendNode> getNode(String nodeId) {
        return Optional.ofNullable(nodes.get(nodeId));
    }

    public List<BackendNode> getAllNodes() {
        return new ArrayList<>(nodes.values());
    }

    public List<BackendNode> getAvailableNodes() {
        return nodes.values().stream()
                .filter(node -> node.getStatus() == NodeStatus.HEALTHY)
                .collect(Collectors.toList());
    }

    public List<BackendNode> getNodesWithTags(List<String> requiredTags) {
        if (requiredTags == null || requiredTags.isEmpty()) {
            return getAvailableNodes();
        }
        return getAvailableNodes().stream()
                .filter(node -> node.getTags().containsAll(requiredTags))
                .collect(Collectors.toList());
    }

    public void updateStatus(String nodeId, NodeStatus status) {
        BackendNode node = nodes.get(nodeId);
        if (node != null) {
            node.setStatus(status);
            node.setStatusChangedAt(LocalDateTime.now());
            nodes.put(nodeId, node);
        }
    }
}
