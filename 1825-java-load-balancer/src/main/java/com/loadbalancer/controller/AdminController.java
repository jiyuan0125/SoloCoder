package com.loadbalancer.controller;

import com.loadbalancer.dto.NodeRegistrationRequest;
import com.loadbalancer.dto.RoutingRuleRequest;
import com.loadbalancer.dto.WeightUpdateRequest;
import com.loadbalancer.model.BackendNode;
import com.loadbalancer.model.HealthCheckResult;
import com.loadbalancer.model.RoutingRule;
import com.loadbalancer.model.TrafficLog;
import com.loadbalancer.service.*;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDateTime;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/lb-api")
public class AdminController {
    private final NodeManager nodeManager;
    private final RoutingRuleManager ruleManager;
    private final TrafficLogService trafficLogService;
    private final HealthCheckService healthCheckService;

    public AdminController(NodeManager nodeManager,
                            RoutingRuleManager ruleManager,
                            TrafficLogService trafficLogService,
                            HealthCheckService healthCheckService) {
        this.nodeManager = nodeManager;
        this.ruleManager = ruleManager;
        this.trafficLogService = trafficLogService;
        this.healthCheckService = healthCheckService;
    }

    @PostMapping("/nodes")
    public ResponseEntity<BackendNode> registerNode(@RequestBody NodeRegistrationRequest request) {
        BackendNode node = nodeManager.register(request);
        return ResponseEntity.ok(node);
    }

    @DeleteMapping("/nodes/{nodeId}")
    public ResponseEntity<Map<String, Object>> deregisterNode(@PathVariable String nodeId) {
        return nodeManager.deregister(nodeId)
                .map(node -> {
                    Map<String, Object> resp = new HashMap<>();
                    resp.put("message", "Node marked as DRAINING, will be removed after 60s");
                    resp.put("node", node);
                    return ResponseEntity.ok(resp);
                })
                .orElse(ResponseEntity.notFound().build());
    }

    @PutMapping("/nodes/{nodeId}/weight")
    public ResponseEntity<BackendNode> updateWeight(@PathVariable String nodeId,
                                                     @RequestBody WeightUpdateRequest request) {
        return nodeManager.updateWeight(nodeId, request.getWeight())
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping("/nodes")
    public ResponseEntity<List<BackendNode>> listNodes() {
        return ResponseEntity.ok(nodeManager.getAllNodes());
    }

    @GetMapping("/nodes/{nodeId}")
    public ResponseEntity<BackendNode> getNode(@PathVariable String nodeId) {
        return nodeManager.getNode(nodeId)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }

    @PostMapping("/rules")
    public ResponseEntity<RoutingRule> addRule(@RequestBody RoutingRuleRequest request) {
        RoutingRule rule = ruleManager.addRule(request);
        return ResponseEntity.ok(rule);
    }

    @DeleteMapping("/rules/{ruleId}")
    public ResponseEntity<Map<String, String>> removeRule(@PathVariable String ruleId) {
        if (ruleManager.removeRule(ruleId)) {
            Map<String, String> resp = new HashMap<>();
            resp.put("message", "Rule removed");
            return ResponseEntity.ok(resp);
        }
        return ResponseEntity.notFound().build();
    }

    @GetMapping("/rules")
    public ResponseEntity<List<RoutingRule>> listRules() {
        return ResponseEntity.ok(ruleManager.getAllRules());
    }

    @GetMapping("/rules/{ruleId}")
    public ResponseEntity<RoutingRule> getRule(@PathVariable String ruleId) {
        return ruleManager.getRule(ruleId)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping("/traffic-logs")
    public ResponseEntity<List<TrafficLog>> getTrafficLogs(
            @RequestParam(required = false) String nodeId,
            @RequestParam(required = false) @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) LocalDateTime start,
            @RequestParam(required = false) @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) LocalDateTime end,
            @RequestParam(defaultValue = "100") int limit) {
        List<TrafficLog> logs = trafficLogService.getLogs(nodeId, start, end, limit);
        return ResponseEntity.ok(logs);
    }

    @GetMapping("/health/results")
    public ResponseEntity<List<HealthCheckResult>> getHealthCheckResults() {
        return ResponseEntity.ok(healthCheckService.getLastResults());
    }

    @GetMapping("/health/nodes")
    public ResponseEntity<List<Map<String, Object>>> getAllNodeHealthStatus() {
        List<Map<String, Object>> statuses = new java.util.ArrayList<>();
        for (BackendNode node : nodeManager.getAllNodes()) {
            statuses.add(healthCheckService.getNodeHealthStatus(node.getId()));
        }
        return ResponseEntity.ok(statuses);
    }

    @GetMapping("/health/nodes/{nodeId}")
    public ResponseEntity<Map<String, Object>> getNodeHealthStatus(@PathVariable String nodeId) {
        if (nodeManager.getNode(nodeId).isEmpty()) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(healthCheckService.getNodeHealthStatus(nodeId));
    }

    @GetMapping("/status")
    public ResponseEntity<Map<String, Object>> getOverallStatus() {
        Map<String, Object> status = new HashMap<>();
        List<BackendNode> allNodes = nodeManager.getAllNodes();
        status.put("totalNodes", allNodes.size());
        status.put("availableNodes", nodeManager.getAvailableNodes().size());
        status.put("totalRules", ruleManager.getAllRules().size());
        return ResponseEntity.ok(status);
    }
}
