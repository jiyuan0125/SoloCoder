package com.example.proxy.controller;

import com.example.proxy.model.RewriteRule;
import com.example.proxy.service.RewriteRuleService;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.List;
import java.util.Optional;

@RestController
@RequestMapping("/rewrite-rules")
public class RewriteRuleController {

    private final RewriteRuleService ruleService;

    public RewriteRuleController(RewriteRuleService ruleService) {
        this.ruleService = ruleService;
    }

    @GetMapping
    public List<RewriteRule> getAllRules() {
        return ruleService.getAllRules();
    }

    @GetMapping("/{id}")
    public ResponseEntity<RewriteRule> getRuleById(@PathVariable Long id) {
        Optional<RewriteRule> rule = ruleService.getRuleById(id);
        return rule.map(ResponseEntity::ok)
                .orElseGet(() -> ResponseEntity.notFound().build());
    }

    @PostMapping
    public ResponseEntity<RewriteRule> addRule(@Valid @RequestBody RewriteRule rule) {
        try {
            RewriteRule created = ruleService.addRule(rule);
            return ResponseEntity.status(HttpStatus.CREATED).body(created);
        } catch (Exception e) {
            return ResponseEntity.badRequest().build();
        }
    }

    @PutMapping("/{id}")
    public ResponseEntity<RewriteRule> updateRule(
            @PathVariable Long id,
            @Valid @RequestBody RewriteRule updatedRule) {
        try {
            Optional<RewriteRule> rule = ruleService.updateRule(id, updatedRule);
            return rule.map(ResponseEntity::ok)
                    .orElseGet(() -> ResponseEntity.notFound().build());
        } catch (Exception e) {
            return ResponseEntity.badRequest().build();
        }
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteRule(@PathVariable Long id) {
        boolean deleted = ruleService.deleteRule(id);
        if (deleted) {
            return ResponseEntity.noContent().build();
        }
        return ResponseEntity.notFound().build();
    }
}
