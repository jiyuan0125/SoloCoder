package com.loadbalancer.service;

import com.loadbalancer.model.RoutingRule;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.regex.Pattern;

@Service
public class RoutingRuleService {
    private final Map<String, RoutingRule> rules = new ConcurrentHashMap<>();
    private final Map<String, Pattern> patternCache = new ConcurrentHashMap<>();

    public void addRule(RoutingRule rule) {
        if (rule.getId() == null || rule.getId().isEmpty()) {
            throw new IllegalArgumentException("Rule id cannot be empty");
        }
        if (rule.getPathPattern() == null || rule.getPathPattern().isEmpty()) {
            throw new IllegalArgumentException("Path pattern cannot be empty");
        }
        rules.put(rule.getId(), rule);
        patternCache.put(rule.getPathPattern(), convertToRegex(rule.getPathPattern()));
    }

    public void removeRule(String ruleId) {
        RoutingRule rule = rules.remove(ruleId);
        if (rule != null) {
            patternCache.remove(rule.getPathPattern());
        }
    }

    public Optional<RoutingRule> getRule(String ruleId) {
        return Optional.ofNullable(rules.get(ruleId));
    }

    public List<RoutingRule> getAllRules() {
        List<RoutingRule> ruleList = new ArrayList<>(rules.values());
        ruleList.sort((r1, r2) -> Integer.compare(r2.getPriority(), r1.getPriority()));
        return ruleList;
    }

    public Optional<RoutingRule> matchRule(String path) {
        return getAllRules().stream()
                .filter(RoutingRule::isEnabled)
                .filter(rule -> matchesPath(rule.getPathPattern(), path))
                .findFirst();
    }

    private boolean matchesPath(String pattern, String path) {
        Pattern regex = patternCache.get(pattern);
        if (regex == null) {
            regex = convertToRegex(pattern);
            patternCache.put(pattern, regex);
        }
        return regex.matcher(path).matches();
    }

    private Pattern convertToRegex(String pathPattern) {
        String regex = pathPattern
                .replace(".", "\\.")
                .replace("*", ".*")
                .replace("?", ".");
        if (!regex.startsWith("^")) {
            regex = "^" + regex;
        }
        if (!regex.endsWith("$")) {
            regex = regex + "$";
        }
        return Pattern.compile(regex);
    }

    public void updateRule(RoutingRule updatedRule) {
        if (!rules.containsKey(updatedRule.getId())) {
            throw new IllegalArgumentException("Rule not found: " + updatedRule.getId());
        }
        RoutingRule oldRule = rules.get(updatedRule.getId());
        if (!oldRule.getPathPattern().equals(updatedRule.getPathPattern())) {
            patternCache.remove(oldRule.getPathPattern());
            patternCache.put(updatedRule.getPathPattern(), convertToRegex(updatedRule.getPathPattern()));
        }
        rules.put(updatedRule.getId(), updatedRule);
    }
}
