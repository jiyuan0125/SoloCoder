package com.example.proxy.model;

import javax.validation.constraints.Min;
import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;
import java.util.concurrent.atomic.AtomicLong;

public class RewriteRule {

    private Long id;

    @NotBlank(message = "Pattern is required")
    private String pattern;

    @NotBlank(message = "Replacement template is required")
    private String replacement;

    @NotNull(message = "Priority is required")
    @Min(value = 0, message = "Priority must be >= 0")
    private Integer priority;

    private final AtomicLong matchCount = new AtomicLong(0);

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getPattern() {
        return pattern;
    }

    public void setPattern(String pattern) {
        this.pattern = pattern;
    }

    public String getReplacement() {
        return replacement;
    }

    public void setReplacement(String replacement) {
        this.replacement = replacement;
    }

    public Integer getPriority() {
        return priority;
    }

    public void setPriority(Integer priority) {
        this.priority = priority;
    }

    public long getMatchCount() {
        return matchCount.get();
    }

    public long incrementMatchCount() {
        return matchCount.incrementAndGet();
    }
}
