package com.loadbalancer.service;

import com.loadbalancer.model.BackendNode;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import java.io.IOException;
import java.net.Socket;
import java.util.List;

@Service
public class HealthCheckService {
    private static final Logger logger = LoggerFactory.getLogger(HealthCheckService.class);
    private static final int CONSECUTIVE_FAILURES_THRESHOLD = 3;
    private static final int CONSECUTIVE_SUCCESSES_THRESHOLD = 2;
    private static final int CONNECTION_TIMEOUT = 5000;

    private final NodeRegistryService nodeRegistryService;

    public HealthCheckService(NodeRegistryService nodeRegistryService) {
        this.nodeRegistryService = nodeRegistryService;
    }

    @Scheduled(fixedRate = 10000)
    public void performHealthCheck() {
        List<BackendNode> allNodes = nodeRegistryService.getAllNodes();
        for (BackendNode node : allNodes) {
            boolean isHealthy = checkNodeHealth(node);
            processHealthCheckResult(node, isHealthy);
        }
    }

    private boolean checkNodeHealth(BackendNode node) {
        try (Socket socket = new Socket()) {
            socket.connect(new java.net.InetSocketAddress(node.getHost(), node.getPort()), CONNECTION_TIMEOUT);
            return true;
        } catch (IOException e) {
            logger.debug("Health check failed for node {}: {}", node.getId(), e.getMessage());
            return false;
        }
    }

    private void processHealthCheckResult(BackendNode node, boolean isHealthy) {
        node.setLastHealthCheckTime(System.currentTimeMillis());
        
        if (isHealthy) {
            node.incrementSuccess();
            if (!node.isHealthy() && node.getConsecutiveSuccesses() >= CONSECUTIVE_SUCCESSES_THRESHOLD) {
                logger.info("Node {} recovered, marking as healthy", node.getId());
                nodeRegistryService.markNodeHealthy(node.getId());
                node.resetSuccessCount();
            }
        } else {
            node.incrementFailure();
            if (node.isHealthy() && node.getConsecutiveFailures() >= CONSECUTIVE_FAILURES_THRESHOLD) {
                logger.warn("Node {} failed {} consecutive health checks, marking as unhealthy", 
                           node.getId(), CONSECUTIVE_FAILURES_THRESHOLD);
                nodeRegistryService.markNodeUnhealthy(node.getId());
                node.resetFailureCount();
            }
        }
    }

    public void recordForwardingFailure(String nodeId) {
        BackendNode node = nodeRegistryService.getNode(nodeId).orElse(null);
        if (node != null) {
            node.incrementFailure();
            if (node.isHealthy() && node.getConsecutiveFailures() >= CONSECUTIVE_FAILURES_THRESHOLD) {
                logger.warn("Node {} failed {} consecutive forwardings, marking as unhealthy",
                           nodeId, CONSECUTIVE_FAILURES_THRESHOLD);
                nodeRegistryService.markNodeUnhealthy(nodeId);
                node.resetFailureCount();
            }
        }
    }

    public void recordForwardingSuccess(String nodeId) {
        BackendNode node = nodeRegistryService.getNode(nodeId).orElse(null);
        if (node != null && node.getConsecutiveFailures() > 0) {
            node.resetFailureCount();
        }
    }
}
