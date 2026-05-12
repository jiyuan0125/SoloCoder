package com.gateway.engine;

import com.gateway.manager.RuleManager;
import com.gateway.model.Rule;
import com.gateway.model.Rule.RuleType;

import java.util.Collection;
import java.util.Map;

public class RuleMatchingEngine {
    private final RuleManager ruleManager;

    public RuleMatchingEngine(RuleManager ruleManager) {
        this.ruleManager = ruleManager;
    }

    public String matchVersion(Map<String, String> headers, Map<String, String> cookies, String clientIp) {
        Collection<Rule> rules = ruleManager.getAllRules();

        String matchedVersion = matchByType(rules, RuleType.HEADER, headers, null, null);
        if (matchedVersion != null) {
            return matchedVersion;
        }

        matchedVersion = matchByType(rules, RuleType.COOKIE, null, cookies, null);
        if (matchedVersion != null) {
            return matchedVersion;
        }

        matchedVersion = matchByType(rules, RuleType.IP, null, null, clientIp);
        return matchedVersion;
    }

    private String matchByType(Collection<Rule> rules, RuleType type,
                               Map<String, String> headers,
                               Map<String, String> cookies,
                               String clientIp) {
        for (Rule rule : rules) {
            if (rule.getType() != type) {
                continue;
            }

            boolean matches = false;

            switch (type) {
                case HEADER:
                    if (headers != null) {
                        String headerValue = headers.get(rule.getKey());
                        matches = rule.getValue().equals(headerValue);
                    }
                    break;
                case COOKIE:
                    if (cookies != null) {
                        String cookieValue = cookies.get(rule.getKey());
                        matches = rule.getValue().equals(cookieValue);
                    }
                    break;
                case IP:
                    matches = rule.getValue().equals(clientIp);
                    break;
            }

            if (matches) {
                return rule.getTargetVersion();
            }
        }
        return null;
    }
}
