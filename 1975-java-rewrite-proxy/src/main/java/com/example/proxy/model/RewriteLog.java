package com.example.proxy.model;

import java.time.LocalDateTime;

public class RewriteLog {

    private Long ruleId;
    private String originalPath;
    private String rewrittenPath;
    private LocalDateTime timestamp;

    public RewriteLog() {
    }

    public RewriteLog(Long ruleId, String originalPath, String rewrittenPath) {
        this.ruleId = ruleId;
        this.originalPath = originalPath;
        this.rewrittenPath = rewrittenPath;
        this.timestamp = LocalDateTime.now();
    }

    public Long getRuleId() {
        return ruleId;
    }

    public void setRuleId(Long ruleId) {
        this.ruleId = ruleId;
    }

    public String getOriginalPath() {
        return originalPath;
    }

    public void setOriginalPath(String originalPath) {
        this.originalPath = originalPath;
    }

    public String getRewrittenPath() {
        return rewrittenPath;
    }

    public void setRewrittenPath(String rewrittenPath) {
        this.rewrittenPath = rewrittenPath;
    }

    public LocalDateTime getTimestamp() {
        return timestamp;
    }

    public void setTimestamp(LocalDateTime timestamp) {
        this.timestamp = timestamp;
    }
}
