package com.gateway.manager;

import com.gateway.model.Rule;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicReference;

public class RuleManager {
    private final AtomicReference<Map<String, Rule>> rulesRef = new AtomicReference<>(new ConcurrentHashMap<>());

    public void addRule(Rule rule) {
        Map<String, Rule> current = new HashMap<>(rulesRef.get());
        current.put(rule.getId(), rule);
        rulesRef.set(new ConcurrentHashMap<>(current));
    }

    public boolean updateRule(String id, Rule newRule) {
        Map<String, Rule> current = new HashMap<>(rulesRef.get());
        if (!current.containsKey(id)) {
            return false;
        }
        current.put(id, newRule);
        rulesRef.set(new ConcurrentHashMap<>(current));
        return true;
    }

    public boolean removeRule(String id) {
        Map<String, Rule> current = new HashMap<>(rulesRef.get());
        if (!current.containsKey(id)) {
            return false;
        }
        current.remove(id);
        rulesRef.set(new ConcurrentHashMap<>(current));
        return true;
    }

    public Rule getRule(String id) {
        return rulesRef.get().get(id);
    }

    public Collection<Rule> getAllRules() {
        return rulesRef.get().values();
    }

    public boolean hasRule(String id) {
        return rulesRef.get().containsKey(id);
    }

    public List<Rule> getRulesForVersion(String versionName) {
        List<Rule> result = new ArrayList<>();
        for (Rule rule : rulesRef.get().values()) {
            if (versionName.equals(rule.getTargetVersion())) {
                result.add(rule);
            }
        }
        return result;
    }
}
