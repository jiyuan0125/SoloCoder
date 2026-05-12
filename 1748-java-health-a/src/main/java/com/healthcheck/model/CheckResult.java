package com.healthcheck.model;

import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;
import lombok.Builder;

import java.time.Instant;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class CheckResult {
    private String serviceId;
    private String checkItemName;
    private CheckStatus status;
    private String message;
    private Long responseTime;
    private Instant timestamp;
    private boolean success;
}
