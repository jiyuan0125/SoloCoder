package com.loadbalancer.model;

import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;

import java.time.Instant;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class DistributionHistory {
    private String nodeId;
    private String reason;
    private String requestPath;
    private Instant timestamp;
    private String routeRuleId;

    public DistributionHistory(String nodeId, String reason, String requestPath, Instant timestamp) {
        this.nodeId = nodeId;
        this.reason = reason;
        this.requestPath = requestPath;
        this.timestamp = timestamp;
    }

    public DistributionHistory(String nodeId, String reason, String requestPath, Instant timestamp, String routeRuleId) {
        this.nodeId = nodeId;
        this.reason = reason;
        this.requestPath = requestPath;
        this.timestamp = timestamp;
        this.routeRuleId = routeRuleId;
    }
}
