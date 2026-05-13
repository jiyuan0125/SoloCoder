package com.example.proxy.service;

import com.example.proxy.model.RewriteRule;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;
import java.util.regex.Pattern;
import java.util.stream.Collectors;

@Service
public class RewriteRuleService {

    private final Map<Long, RewriteRule> rules = new ConcurrentHashMap<>();
    private final AtomicLong idGenerator = new AtomicLong(0);

    public List<RewriteRule> getAllRules() {
        return rules.values().stream()
                .sorted(Comparator.comparingInt(RewriteRule::getPriority))
                .collect(Collectors.toList());
    }

    public Optional<RewriteRule> getRuleById(Long id) {
        return Optional.ofNullable(rules.get(id));
    }

    public RewriteRule addRule(RewriteRule rule) {
        Pattern.compile(rule.getPattern());
        long id = idGenerator.incrementAndGet();
        rule.setId(id);
        rules.put(id, rule);
        return rule;
    }

    public Optional<RewriteRule> updateRule(Long id, RewriteRule updatedRule) {
        RewriteRule existing = rules.get(id);
        if (existing == null) {
            return Optional.empty();
        }
        Pattern.compile(updatedRule.getPattern());
        existing.setPattern(updatedRule.getPattern());
        existing.setReplacement(updatedRule.getReplacement());
        existing.setPriority(updatedRule.getPriority());
        return Optional.of(existing);
    }

    public boolean deleteRule(Long id) {
        return rules.remove(id) != null;
    }

    public Optional<RewriteResult> matchAndRewrite(String pathWithQuery) {
        for (RewriteRule rule : getAllRules()) {
            Pattern pattern = Pattern.compile(rule.getPattern());
            java.util.regex.Matcher matcher = pattern.matcher(pathWithQuery);
            if (matcher.matches()) {
                rule.incrementMatchCount();
                String rewritten = matcher.replaceAll(rule.getReplacement());
                return Optional.of(new RewriteResult(rule.getId(), rewritten));
            }
        }
        return Optional.empty();
    }

    public static class RewriteResult {
        private final Long ruleId;
        private final String rewrittenPath;

        public RewriteResult(Long ruleId, String rewrittenPath) {
            this.ruleId = ruleId;
            this.rewrittenPath = rewrittenPath;
        }

        public Long getRuleId() {
            return ruleId;
        }

        public String getRewrittenPath() {
            return rewrittenPath;
        }
    }
}
