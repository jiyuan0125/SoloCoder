package com.example.loadbalancer.registry;

import com.example.loadbalancer.model.Node;
import com.example.loadbalancer.model.NodeStatus;
import org.springframework.stereotype.Component;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Component
public class NodeRegistry {

    private final Map<String, Node> nodesById = new ConcurrentHashMap<>();
    private final Map<String, String> addressToId = new ConcurrentHashMap<>();

    public synchronized Optional<Node> registerNode(String ip, int port, int weight) {
        String address = ip + ":" + port;
        if (addressToId.containsKey(address)) {
            return Optional.empty();
        }
        
        String id = UUID.randomUUID().toString();
        Node node = new Node(id, ip, port, weight);
        nodesById.put(id, node);
        addressToId.put(address, id);
        return Optional.of(node);
    }

    public synchronized Optional<Node> getNodeById(String id) {
        return Optional.ofNullable(nodesById.get(id));
    }

    public synchronized Optional<Node> offlineNode(String id) {
        Node node = nodesById.get(id);
        if (node == null) {
            return Optional.empty();
        }
        node.markOffline();
        return Optional.of(node);
    }

    public synchronized List<Node> getAllNodes() {
        return new ArrayList<>(nodesById.values());
    }

    public List<Node> getAvailableNodes() {
        return nodesById.values().stream()
                .filter(Node::canReceiveTraffic)
                .collect(Collectors.toList());
    }

    public synchronized void updateNodeHealth(String nodeId, boolean healthy) {
        Node node = nodesById.get(nodeId);
        if (node == null) {
            return;
        }
        
        node.setLastHealthCheckAt(System.currentTimeMillis());
        node.setLastHealthCheckPassed(healthy);
        
        if (healthy) {
            if (node.getStatus() == NodeStatus.PENDING_VERIFICATION || node.getStatus() == NodeStatus.OFFLINE) {
                node.markHealthy();
            }
        } else {
            if (node.getStatus() == NodeStatus.HEALTHY || node.getStatus() == NodeStatus.DEGRADED) {
                node.markOffline();
            }
        }
    }

    public synchronized void markNodeForRecheck(String nodeId) {
        Node node = nodesById.get(nodeId);
        if (node != null && node.getStatus() == NodeStatus.OFFLINE) {
            node.markPending();
        }
    }
}
