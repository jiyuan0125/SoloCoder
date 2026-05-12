package com.loadbalancer.controller;

import com.loadbalancer.model.BackendNode;
import com.loadbalancer.model.DistributionHistory;
import com.loadbalancer.model.RoutingRule;
import com.loadbalancer.service.DistributionHistoryService;
import com.loadbalancer.service.NodeRegistryService;
import com.loadbalancer.service.RoutingRuleService;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.Instant;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;

@RestController
@RequestMapping("/api/admin")
public class AdminController {
    private final NodeRegistryService nodeRegistryService;
    private final RoutingRuleService routingRuleService;
    private final DistributionHistoryService historyService;

    public AdminController(NodeRegistryService nodeRegistryService,
                           RoutingRuleService routingRuleService,
                           DistributionHistoryService historyService) {
        this.nodeRegistryService = nodeRegistryService;
        this.routingRuleService = routingRuleService;
        this.historyService = historyService;
    }

    @PostMapping("/nodes")
    public ResponseEntity<?> registerNode(@RequestBody BackendNode node) {
        try {
            nodeRegistryService.registerNode(node);
            return ResponseEntity.ok(node);
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        }
    }

    @DeleteMapping("/nodes/{nodeId}")
    public ResponseEntity<?> unregisterNode(@PathVariable String nodeId) {
        nodeRegistryService.unregisterNode(nodeId);
        return ResponseEntity.ok(Map.of("message", "Node unregistered successfully"));
    }

    @GetMapping("/nodes")
    public ResponseEntity<List<BackendNode>> getAllNodes() {
        return ResponseEntity.ok(nodeRegistryService.getAllNodes());
    }

    @GetMapping("/nodes/{nodeId}")
    public ResponseEntity<?> getNode(@PathVariable String nodeId) {
        Optional<BackendNode> node = nodeRegistryService.getNode(nodeId);
        if (node.isPresent()) {
            return ResponseEntity.ok(node.get());
        } else {
            return ResponseEntity.notFound().build();
        }
    }

    @PutMapping("/nodes/{nodeId}/weight")
    public ResponseEntity<?> updateWeight(@PathVariable String nodeId,
                                           @RequestBody Map<String, Integer> body) {
        try {
            Integer weight = body.get("weight");
            if (weight == null) {
                return ResponseEntity.badRequest().body(Map.of("error", "Weight is required"));
            }
            nodeRegistryService.updateWeight(nodeId, weight);
            return ResponseEntity.ok(Map.of("message", "Weight updated successfully"));
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        }
    }

    @PutMapping("/nodes/{nodeId}/tags")
    public ResponseEntity<?> updateTags(@PathVariable String nodeId,
                                         @RequestBody Map<String, Set<String>> body) {
        try {
            Set<String> tags = body.get("tags");
            nodeRegistryService.updateTags(nodeId, tags);
            return ResponseEntity.ok(Map.of("message", "Tags updated successfully"));
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        }
    }

    @PostMapping("/rules")
    public ResponseEntity<?> addRule(@RequestBody RoutingRule rule) {
        try {
            routingRuleService.addRule(rule);
            return ResponseEntity.ok(rule);
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        }
    }

    @DeleteMapping("/rules/{ruleId}")
    public ResponseEntity<?> removeRule(@PathVariable String ruleId) {
        routingRuleService.removeRule(ruleId);
        return ResponseEntity.ok(Map.of("message", "Rule removed successfully"));
    }

    @GetMapping("/rules")
    public ResponseEntity<List<RoutingRule>> getAllRules() {
        return ResponseEntity.ok(routingRuleService.getAllRules());
    }

    @GetMapping("/rules/{ruleId}")
    public ResponseEntity<?> getRule(@PathVariable String ruleId) {
        Optional<RoutingRule> rule = routingRuleService.getRule(ruleId);
        if (rule.isPresent()) {
            return ResponseEntity.ok(rule.get());
        } else {
            return ResponseEntity.notFound().build();
        }
    }

    @PutMapping("/rules/{ruleId}")
    public ResponseEntity<?> updateRule(@PathVariable String ruleId,
                                        @RequestBody RoutingRule updatedRule) {
        try {
            if (!ruleId.equals(updatedRule.getId())) {
                return ResponseEntity.badRequest().body(Map.of("error", "Rule ID mismatch"));
            }
            routingRuleService.updateRule(updatedRule);
            return ResponseEntity.ok(Map.of("message", "Rule updated successfully"));
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        }
    }

    @GetMapping("/history")
    public ResponseEntity<List<DistributionHistory>> getHistory(
            @RequestParam(defaultValue = "100") int limit,
            @RequestParam(required = false) @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) Instant startTime,
            @RequestParam(required = false) @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) Instant endTime,
            @RequestParam(required = false) String nodeId) {
        
        if (nodeId != null) {
            return ResponseEntity.ok(historyService.getHistoryByNode(nodeId, limit));
        }
        
        if (startTime != null && endTime != null) {
            return ResponseEntity.ok(historyService.getHistoryByTimeRange(startTime, endTime));
        }
        
        return ResponseEntity.ok(historyService.getRecentHistory(limit));
    }
}
