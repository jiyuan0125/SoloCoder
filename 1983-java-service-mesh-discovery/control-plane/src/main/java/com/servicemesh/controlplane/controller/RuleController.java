package com.servicemesh.controlplane.controller;

import com.servicemesh.common.model.TrafficRule;
import com.servicemesh.controlplane.service.RuleService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/rules")
public class RuleController {

    private final RuleService ruleService;

    public RuleController(RuleService ruleService) {
        this.ruleService = ruleService;
    }

    @GetMapping
    public ResponseEntity<List<TrafficRule>> getAllRules() {
        return ResponseEntity.ok(ruleService.getAllRules());
    }

    @GetMapping("/{serviceName}")
    public ResponseEntity<TrafficRule> getRule(@PathVariable String serviceName) {
        TrafficRule rule = ruleService.getRule(serviceName);
        return rule != null ? ResponseEntity.ok(rule) : ResponseEntity.notFound().build();
    }

    @PostMapping
    public ResponseEntity<Map<String, String>> setRule(@RequestBody TrafficRule rule) {
        ruleService.setRule(rule);
        return ResponseEntity.ok(Map.of("status", "ok"));
    }

    @DeleteMapping("/{serviceName}")
    public ResponseEntity<Map<String, String>> removeRule(@PathVariable String serviceName) {
        ruleService.removeRule(serviceName);
        return ResponseEntity.ok(Map.of("status", "ok"));
    }
}
