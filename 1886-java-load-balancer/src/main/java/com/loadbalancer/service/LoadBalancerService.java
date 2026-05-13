package com.loadbalancer.service;

import com.loadbalancer.model.Node;
import com.loadbalancer.model.NodeStatus;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.Collection;
import java.util.List;
import java.util.concurrent.locks.ReadWriteLock;
import java.util.concurrent.locks.ReentrantReadWriteLock;

@Slf4j
@Service
@RequiredArgsConstructor
public class LoadBalancerService {
    private final NodeRegistryService nodeRegistryService;
    private final ReadWriteLock lock = new ReentrantReadWriteLock();

    public Node selectNode() {
        lock.writeLock().lock();
        try {
            Collection<Node> availableNodes = getAvailableNodes();
            if (availableNodes.isEmpty()) {
                log.warn("No available nodes");
                return null;
            }

            Node selectedNode = null;
            int totalWeight = 0;
            
            for (Node node : availableNodes) {
                node.getLock().readLock().lock();
                try {
                    totalWeight += node.getWeight();
                    node.setCurrentWeight(node.getCurrentWeight() + node.getWeight());
                    if (selectedNode == null || node.getCurrentWeight() > selectedNode.getCurrentWeight()) {
                        selectedNode = node;
                    }
                } finally {
                    node.getLock().readLock().unlock();
                }
            }

            if (selectedNode != null) {
                selectedNode.getLock().writeLock().lock();
                try {
                    selectedNode.setCurrentWeight(selectedNode.getCurrentWeight() - totalWeight);
                } finally {
                    selectedNode.getLock().writeLock().unlock();
                }
            }

            return selectedNode;
        } finally {
            lock.writeLock().unlock();
        }
    }

    private Collection<Node> getAvailableNodes() {
        List<Node> availableNodes = new ArrayList<>();
        for (Node node : nodeRegistryService.getAllNodes()) {
            if (node.getStatus() == NodeStatus.NORMAL_SERVICE) {
                availableNodes.add(node);
            }
        }
        return availableNodes;
    }
}
