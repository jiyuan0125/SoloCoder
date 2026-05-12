package com.example.protobridge.controller;

import com.example.protobridge.config.ConversionRule;
import com.example.protobridge.config.ConversionRuleService;
import com.example.protobridge.exception.ResourceNotFoundException;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/admin/rules")
@RequiredArgsConstructor
@Validated
public class ConversionRuleController {

    private final ConversionRuleService ruleService;

    @GetMapping
    public ResponseEntity<List<ConversionRule>> getAllRules() {
        return ResponseEntity.ok(ruleService.getAllRules());
    }

    @GetMapping("/{id}")
    public ResponseEntity<ConversionRule> getRule(@PathVariable String id) {
        ConversionRule rule = ruleService.getRule(id);
        if (rule == null) {
            throw new ResourceNotFoundException("规则不存在: " + id);
        }
        return ResponseEntity.ok(rule);
    }

    @PostMapping
    public ResponseEntity<ConversionRule> createRule(@Valid @RequestBody ConversionRule rule) {
        ConversionRule created = ruleService.createRule(rule);
        return ResponseEntity.status(HttpStatus.CREATED).body(created);
    }

    @PutMapping("/{id}")
    public ResponseEntity<ConversionRule> updateRule(@PathVariable String id, @Valid @RequestBody ConversionRule rule) {
        ConversionRule updated = ruleService.updateRule(id, rule);
        if (updated == null) {
            throw new ResourceNotFoundException("规则不存在: " + id);
        }
        return ResponseEntity.ok(updated);
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Map<String, Object>> deleteRule(@PathVariable String id) {
        boolean deleted = ruleService.deleteRule(id);
        Map<String, Object> result = new HashMap<>();
        result.put("success", deleted);
        result.put("id", id);
        if (!deleted) {
            throw new ResourceNotFoundException("规则不存在: " + id);
        }
        return ResponseEntity.ok(result);
    }

    @GetMapping("/search")
    public ResponseEntity<ConversionRule> getRuleByPath(
            @RequestParam String path,
            @RequestParam(defaultValue = "POST") String method) {
        ConversionRule rule = ruleService.getRuleByPath(path, method);
        if (rule == null) {
            throw new ResourceNotFoundException("未找到路径 " + method + " " + path + " 对应的转换规则");
        }
        return ResponseEntity.ok(rule);
    }
}
