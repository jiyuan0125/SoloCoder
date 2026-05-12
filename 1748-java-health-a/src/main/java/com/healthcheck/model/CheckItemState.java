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
public class CheckItemState {
    private String serviceId;
    private String checkItemName;
    private CheckStatus currentStatus;
    private CheckStatus previousStatus;
    private Instant lastStateChangeTime;
    private int consecutiveFailures;
    private CheckResult lastResult;
}
