package com.loadbalancer.service;

import com.loadbalancer.model.BackendNode;
import com.loadbalancer.model.BackendNode.NodeStatus;
import com.loadbalancer.model.HealthCheckResult;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.net.Socket;
import java.time.LocalDateTime;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class HealthCheckService {
    @Value("${loadbalancer.health-check.interval-ms:10000}")
    private long intervalMs;

    @Value("${loadbalancer.health-check.tcp-timeout-ms:5000}")
    private int tcpTimeoutMs;

    @Value("${loadbalancer.health-check.failure-threshold:3}")
    private int failureThreshold;

    @Value("${loadbalancer.health-check.recovery-threshold:2}")
    private int recoveryThreshold;

    private final NodeManager nodeManager;
    private final Map<String, Integer> consecutiveFailures = new ConcurrentHashMap<>();
    private final Map<String, Integer> consecutiveSuccesses = new ConcurrentHashMap<>();
    private final List<HealthCheckResult> lastResults = Collections.synchronizedList(new ArrayList<>());

    public HealthCheckService(NodeManager nodeManager) {
        this.nodeManager = nodeManager;
    }

    @Scheduled(fixedRateString = "${loadbalancer.health-check.interval-ms:10000}")
    public void runHealthChecks() {
        List<BackendNode> nodes = nodeManager.getAllNodes();
        for (BackendNode node : nodes) {
            if (node.getStatus() == NodeStatus.DRAINING) {
                continue;
            }
            checkNode(node);
        }
    }

    public HealthCheckResult checkNode(BackendNode node) {
        boolean success = tcpProbe(node.getHost(), node.getPort());
        int failures = consecutiveFailures.getOrDefault(node.getId(), 0);
        int successes = consecutiveSuccesses.getOrDefault(node.getId(), 0);

        if (success) {
            failures = 0;
            successes++;
            if (node.getStatus() == NodeStatus.UNHEALTHY && successes >= recoveryThreshold) {
                nodeManager.updateStatus(node.getId(), NodeStatus.HEALTHY);
            }
        } else {
            successes = 0;
            failures++;
            if (node.getStatus() == NodeStatus.HEALTHY && failures >= failureThreshold) {
                nodeManager.updateStatus(node.getId(), NodeStatus.UNHEALTHY);
            }
        }

        consecutiveFailures.put(node.getId(), failures);
        consecutiveSuccesses.put(node.getId(), successes);

        HealthCheckResult result = HealthCheckResult.builder()
                .nodeId(node.getId())
                .success(success)
                .errorMessage(success ? null : "TCP connection failed")
                .consecutiveFailures(failures)
                .consecutiveSuccesses(successes)
                .checkedAt(LocalDateTime.now())
                .build();

        lastResults.add(result);
        if (lastResults.size() > 100) {
            lastResults.remove(0);
        }

        return result;
    }

    private boolean tcpProbe(String host, int port) {
        try (Socket socket = new Socket()) {
            socket.connect(new InetSocketAddress(host, port), tcpTimeoutMs);
            return true;
        } catch (IOException e) {
            return false;
        }
    }

    public void recordForwardFailure(String nodeId) {
        int failures = consecutiveFailures.merge(nodeId, 1, Integer::sum);
        consecutiveSuccesses.put(nodeId, 0);

        Optional<BackendNode> nodeOpt = nodeManager.getNode(nodeId);
        if (nodeOpt.isPresent() && nodeOpt.get().getStatus() == NodeStatus.HEALTHY && failures >= failureThreshold) {
            nodeManager.updateStatus(nodeId, NodeStatus.UNHEALTHY);
        }
    }

    public List<HealthCheckResult> getLastResults() {
        return new ArrayList<>(lastResults);
    }

    public Map<String, Object> getNodeHealthStatus(String nodeId) {
        Map<String, Object> status = new HashMap<>();
        status.put("nodeId", nodeId);
        status.put("consecutiveFailures", consecutiveFailures.getOrDefault(nodeId, 0));
        status.put("consecutiveSuccesses", consecutiveSuccesses.getOrDefault(nodeId, 0));
        nodeManager.getNode(nodeId).ifPresent(node -> status.put("currentStatus", node.getStatus()));
        return status;
    }
}
