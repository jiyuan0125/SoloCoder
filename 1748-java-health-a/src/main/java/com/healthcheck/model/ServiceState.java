package com.healthcheck.model;

import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;
import lombok.Builder;

import java.time.Instant;
import java.util.Map;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class ServiceState {
    private String serviceId;
    private ServiceHealthStatus overallStatus;
    private ServiceHealthStatus previousStatus;
    private Instant lastStateChangeTime;
    private Map<String, CheckItemState> checkItemStates;
}
