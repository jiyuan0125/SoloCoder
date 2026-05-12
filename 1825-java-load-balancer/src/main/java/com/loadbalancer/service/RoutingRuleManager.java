package com.loadbalancer.service;

import com.loadbalancer.model.RoutingRule;
import com.loadbalancer.dto.RoutingRuleRequest;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.stream.Collectors;

@Service
public class RoutingRuleManager {
    private final Map<String, RoutingRule> rules = new ConcurrentHashMap<>();
    private final AtomicInteger ruleIdGenerator = new AtomicInteger(0);

    public RoutingRule addRule(RoutingRuleRequest request) {
        String ruleId = "rule-" + ruleIdGenerator.incrementAndGet();
        RoutingRule rule = RoutingRule.builder()
                .id(ruleId)
                .pathPattern(request.getPathPattern())
                .requiredTags(request.getRequiredTags() != null ? new ArrayList<>(request.getRequiredTags()) : new ArrayList<>())
                .description(request.getDescription())
                .priority(request.getPriority())
                .build();
        rules.put(ruleId, rule);
        return rule;
    }

    public boolean removeRule(String ruleId) {
        return rules.remove(ruleId) != null;
    }

    public Optional<RoutingRule> getRule(String ruleId) {
        return Optional.ofNullable(rules.get(ruleId));
    }

    public List<RoutingRule> getAllRules() {
        return rules.values().stream()
                .sorted(Comparator.comparingInt(RoutingRule::getPriority).reversed())
                .collect(Collectors.toList());
    }

    public Optional<RoutingRule> findMatchingRule(String path) {
        return getAllRules().stream()
                .filter(rule -> matchesPath(path, rule.getPathPattern()))
                .findFirst();
    }

    private boolean matchesPath(String path, String pattern) {
        if (path == null || pattern == null) {
            return false;
        }
        if (pattern.endsWith("/*")) {
            String prefix = pattern.substring(0, pattern.length() - 2);
            return path.startsWith(prefix);
        }
        if (pattern.endsWith("/**")) {
            String prefix = pattern.substring(0, pattern.length() - 3);
            return path.startsWith(prefix);
        }
        return path.equals(pattern);
    }
}
