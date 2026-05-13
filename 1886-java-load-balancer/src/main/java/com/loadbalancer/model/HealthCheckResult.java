package com.loadbalancer.model;

import lombok.Data;
import lombok.AllArgsConstructor;
import java.time.Instant;

@Data
@AllArgsConstructor
public class HealthCheckResult {
    private Instant timestamp;
    private boolean success;
}
