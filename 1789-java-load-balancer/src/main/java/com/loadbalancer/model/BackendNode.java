package com.loadbalancer.model;

import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;

import java.util.Set;
import java.util.HashSet;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class BackendNode {
    private String id;
    private String host;
    private int port;
    private int weight = 1;
    private Set<String> tags = new HashSet<>();
    private boolean healthy = true;
    private int currentWeight = 0;
    private int consecutiveFailures = 0;
    private int consecutiveSuccesses = 0;
    private long lastHealthCheckTime;

    public BackendNode(String id, String host, int port, int weight, Set<String> tags) {
        this.id = id;
        this.host = host;
        this.port = port;
        this.weight = weight;
        this.tags = tags != null ? tags : new HashSet<>();
    }

    public void incrementFailure() {
        this.consecutiveFailures++;
        this.consecutiveSuccesses = 0;
    }

    public void incrementSuccess() {
        this.consecutiveSuccesses++;
        this.consecutiveFailures = 0;
    }

    public void resetFailureCount() {
        this.consecutiveFailures = 0;
    }

    public void resetSuccessCount() {
        this.consecutiveSuccesses = 0;
    }

    public boolean hasTag(String tag) {
        return tags.contains(tag);
    }

    public boolean hasAllTags(Set<String> requiredTags) {
        return tags.containsAll(requiredTags);
    }
}
