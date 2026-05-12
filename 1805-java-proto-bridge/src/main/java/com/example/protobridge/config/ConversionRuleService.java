package com.example.protobridge.config;

import org.springframework.stereotype.Service;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class ConversionRuleService {

    private final Map<String, ConversionRule> rulesById = new ConcurrentHashMap<>();
    private final Map<String, ConversionRule> rulesByPathMethod = new ConcurrentHashMap<>();

    public ConversionRule createRule(ConversionRule rule) {
        String id = UUID.randomUUID().toString();
        rule.setId(id);
        rulesById.put(id, rule);
        String key = buildKey(rule.getPath(), rule.getMethod());
        rulesByPathMethod.put(key, rule);
        return rule;
    }

    public ConversionRule updateRule(String id, ConversionRule rule) {
        if (!rulesById.containsKey(id)) {
            return null;
        }
        ConversionRule existing = rulesById.get(id);
        String oldKey = buildKey(existing.getPath(), existing.getMethod());
        rulesByPathMethod.remove(oldKey);

        rule.setId(id);
        rulesById.put(id, rule);
        String newKey = buildKey(rule.getPath(), rule.getMethod());
        rulesByPathMethod.put(newKey, rule);
        return rule;
    }

    public boolean deleteRule(String id) {
        ConversionRule rule = rulesById.remove(id);
        if (rule != null) {
            String key = buildKey(rule.getPath(), rule.getMethod());
            rulesByPathMethod.remove(key);
            return true;
        }
        return false;
    }

    public ConversionRule getRule(String id) {
        return rulesById.get(id);
    }

    public ConversionRule getRuleByPath(String path, String method) {
        String key = buildKey(path, method != null ? method.toUpperCase() : null);
        return rulesByPathMethod.get(key);
    }

    public List<ConversionRule> getAllRules() {
        return new ArrayList<>(rulesById.values());
    }

    private String buildKey(String path, String method) {
        if (path == null || method == null) {
            return null;
        }
        return method.toUpperCase() + ":" + path;
    }
}
