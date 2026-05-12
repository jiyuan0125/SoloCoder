package com.loadbalancer.model;

import lombok.Data;
import lombok.Builder;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

@Data
@Builder
@AllArgsConstructor
@NoArgsConstructor
public class HealthCheckResult {
    private String nodeId;
    private boolean success;
    private String errorMessage;
    private int consecutiveFailures;
    private int consecutiveSuccesses;
    private LocalDateTime checkedAt;
}
