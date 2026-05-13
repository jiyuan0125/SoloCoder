package com.loadbalancer.service;

import com.loadbalancer.config.LoadBalancerConfig;
import com.loadbalancer.model.HealthCheckResult;
import com.loadbalancer.model.Node;
import com.loadbalancer.model.NodeStatus;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

@Slf4j
@Service
@RequiredArgsConstructor
public class NodeRegistryService {
    private final Map<String, Node> nodes = new ConcurrentHashMap<>();
    private final Map<String, Node> addressIndex = new ConcurrentHashMap<>();
    private final LoadBalancerConfig loadBalancerConfig;

    public Node registerNode(String address, int weight) {
        if (addressIndex.containsKey(address)) {
            log.warn("Address {} is already registered", address);
            return null;
        }

        String id = UUID.randomUUID().toString();
        Node node = new Node(id, address, Math.max(weight, 0));
        node.setHealthCheckIntervalSeconds(loadBalancerConfig.getDefaultIntervalSeconds());
        node.setHealthCheckTimeoutSeconds(loadBalancerConfig.getDefaultTimeoutSeconds());
        
        nodes.put(id, node);
        addressIndex.put(address, node);
        log.info("Node registered: id={}, address={}", id, address);
        return node;
    }

    public Node getNode(String id) {
        return nodes.get(id);
    }

    public Collection<Node> getAllNodes() {
        return nodes.values();
    }

    public boolean updateWeight(String id, int weight) {
        Node node = nodes.get(id);
        if (node == null) {
            return false;
        }
        node.getLock().writeLock().lock();
        try {
            node.setWeight(Math.max(weight, 0));
            log.info("Updated weight for node {}: {}", id, weight);
        } finally {
            node.getLock().writeLock().unlock();
        }
        return true;
    }

    public boolean updateHealthConfig(String id, Integer intervalSeconds, Integer timeoutSeconds) {
        Node node = nodes.get(id);
        if (node == null) {
            return false;
        }
        node.getLock().writeLock().lock();
        try {
            if (intervalSeconds != null) {
                node.setHealthCheckIntervalSeconds(intervalSeconds);
            }
            if (timeoutSeconds != null) {
                node.setHealthCheckTimeoutSeconds(timeoutSeconds);
            }
            log.info("Updated health config for node {}: interval={}s, timeout={}s", 
                    id, node.getHealthCheckIntervalSeconds(), node.getHealthCheckTimeoutSeconds());
        } finally {
            node.getLock().writeLock().unlock();
        }
        return true;
    }

    public void processHealthCheckResult(Node node, boolean success) {
        HealthCheckResult result = new HealthCheckResult(java.time.Instant.now(), success);
        node.getLock().writeLock().lock();
        try {
            node.addHealthCheckResult(result);
            NodeStatus currentStatus = node.getStatus();
            
            if (currentStatus == NodeStatus.OFFLINE) {
                return;
            }

            if (success) {
                node.setConsecutiveSuccesses(node.getConsecutiveSuccesses() + 1);
                node.setConsecutiveFailures(0);
                
                if (currentStatus == NodeStatus.NEW_REGISTERED) {
                    node.setStatus(NodeStatus.NORMAL_SERVICE);
                    node.setConsecutiveSuccesses(0);
                    log.info("Node {} transitioned: NEW_REGISTERED -> NORMAL_SERVICE", node.getId());
                } else if (currentStatus == NodeStatus.SUSPECTED_FAILURE && node.getConsecutiveSuccesses() >= 2) {
                    node.setStatus(NodeStatus.NORMAL_SERVICE);
                    node.setConsecutiveSuccesses(0);
                    log.info("Node {} transitioned: SUSPECTED_FAILURE -> NORMAL_SERVICE", node.getId());
                }
            } else {
                node.setConsecutiveFailures(node.getConsecutiveFailures() + 1);
                node.setConsecutiveSuccesses(0);
                
                if (currentStatus == NodeStatus.NORMAL_SERVICE && node.getConsecutiveFailures() >= 2) {
                    node.setStatus(NodeStatus.SUSPECTED_FAILURE);
                    node.setConsecutiveFailures(0);
                    log.warn("Node {} transitioned: NORMAL_SERVICE -> SUSPECTED_FAILURE", node.getId());
                }
            }
        } finally {
            node.getLock().writeLock().unlock();
        }
    }

    public boolean takeNodeOffline(String id) {
        Node node = nodes.get(id);
        if (node == null) {
            return false;
        }
        node.getLock().writeLock().lock();
        try {
            node.setStatus(NodeStatus.OFFLINE);
            node.setConsecutiveSuccesses(0);
            node.setConsecutiveFailures(0);
            log.info("Node {} taken offline", id);
        } finally {
            node.getLock().writeLock().unlock();
        }
        return true;
    }

    public boolean bringNodeOnline(String id) {
        Node node = nodes.get(id);
        if (node == null || node.getStatus() != NodeStatus.OFFLINE) {
            return false;
        }
        node.getLock().writeLock().lock();
        try {
            node.setStatus(NodeStatus.NEW_REGISTERED);
            node.setConsecutiveSuccesses(0);
            node.setConsecutiveFailures(0);
            node.getRecentChecks().clear();
            log.info("Node {} brought back online, status reset to NEW_REGISTERED", id);
        } finally {
            node.getLock().writeLock().unlock();
        }
        return true;
    }
}
