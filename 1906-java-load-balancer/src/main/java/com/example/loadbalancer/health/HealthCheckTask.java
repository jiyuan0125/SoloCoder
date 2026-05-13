package com.example.loadbalancer.health;

import com.example.loadbalancer.config.HealthCheckProperties;
import com.example.loadbalancer.model.Node;
import com.example.loadbalancer.model.NodeStatus;
import com.example.loadbalancer.registry.NodeRegistry;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestTemplate;

import java.util.List;

@Component
public class HealthCheckTask {

    private static final Logger log = LoggerFactory.getLogger(HealthCheckTask.class);

    private final NodeRegistry nodeRegistry;
    private final HealthCheckProperties properties;
    private final RestTemplate restTemplate;

    public HealthCheckTask(NodeRegistry nodeRegistry,
                           HealthCheckProperties properties,
                           RestTemplate restTemplate) {
        this.nodeRegistry = nodeRegistry;
        this.properties = properties;
        this.restTemplate = restTemplate;
    }

    @Scheduled(fixedDelayString = "${load-balancer.health-check.interval-ms:5000}")
    public void checkAllNodes() {
        List<Node> nodes = nodeRegistry.getAllNodes();
        for (Node node : nodes) {
            if (node.getStatus() == NodeStatus.OFFLINE) {
                continue;
            }
            checkNode(node);
        }
    }

    private void checkNode(Node node) {
        String url = "http://" + node.getAddress() + properties.getPath();
        boolean healthy = false;
        
        try {
            restTemplate.getForEntity(url, String.class);
            healthy = true;
        } catch (Exception e) {
            log.debug("Health check failed for {}: {}", node.getAddress(), e.getMessage());
        }
        
        nodeRegistry.updateNodeHealth(node.getId(), healthy);
        log.debug("Health check for {}: {}", node.getAddress(), healthy ? "PASS" : "FAIL");
    }
}
