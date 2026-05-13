package com.loadbalancer.manager;

import com.loadbalancer.model.Node;
import com.loadbalancer.model.NodeStatus;
import com.loadbalancer.strategy.ConsistentHashRing;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicInteger;

@Component
public class NodeManager {

    private final ConcurrentHashMap<String, Node> nodes = new ConcurrentHashMap<>();
    private final ConsistentHashRing hashRing;

    @Autowired
    public NodeManager(ConsistentHashRing hashRing) {
        this.hashRing = hashRing;
    }

    public synchronized Optional<Node> registerNode(String ip, int port) {
        String key = ip + ":" + port;
        if (nodes.containsKey(key)) {
            return Optional.empty();
        }
        Node node = new Node(ip, port);
        nodes.put(key, node);
        return Optional.of(node);
    }

    public synchronized boolean activateNode(String ip, int port) {
        String key = ip + ":" + port;
        Node node = nodes.get(key);
        if (node == null) {
            return false;
        }
        if (node.getStatus() != NodeStatus.REGISTERING) {
            return false;
        }
        node.setStatus(NodeStatus.ACTIVE);
        hashRing.addNode(node);
        return true;
    }

    public synchronized boolean startDraining(String ip, int port) {
        String key = ip + ":" + port;
        Node node = nodes.get(key);
        if (node == null) {
            return false;
        }
        if (node.getStatus() != NodeStatus.ACTIVE) {
            return false;
        }
        node.setStatus(NodeStatus.DRAINING);
        hashRing.removeNode(key);
        return true;
    }

    public synchronized boolean markOffline(String ip, int port) {
        String key = ip + ":" + port;
        Node node = nodes.get(key);
        if (node == null) {
            return false;
        }
        node.setStatus(NodeStatus.OFFLINE);
        hashRing.removeNode(key);
        return true;
    }

    public synchronized boolean removeNode(String ip, int port) {
        String key = ip + ":" + port;
        Node node = nodes.remove(key);
        if (node == null) {
            return false;
        }
        hashRing.removeNode(key);
        return true;
    }

    public Optional<Node> getNode(String ip, int port) {
        return Optional.ofNullable(nodes.get(ip + ":" + port));
    }

    public List<Node> getAllNodes() {
        return new ArrayList<>(nodes.values());
    }

    public Node getTargetNode(String sessionId) {
        Node node = hashRing.getNodeForKey(sessionId);
        if (node != null && node.canReceiveTraffic()) {
            return node;
        }
        return null;
    }

    public Node getNodeForExistingSession(String nodeKey) {
        Node node = nodes.get(nodeKey);
        if (node != null && node.canReceiveExistingSessions()) {
            return node;
        }
        return null;
    }

    public AtomicInteger getActiveNodeCount() {
        AtomicInteger count = new AtomicInteger(0);
        nodes.values().forEach(n -> {
            if (n.getStatus() == NodeStatus.ACTIVE) {
                count.incrementAndGet();
            }
        });
        return count;
    }

    public boolean isEmpty() {
        return hashRing.isEmpty();
    }
}
