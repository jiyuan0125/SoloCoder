package com.example.proxy.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class Backend {
    
    private String id;
    private String url;
    private int weight;
    private BackendStatus status;
    private int consecutiveFailures;
    private int consecutiveSuccesses;
    private LocalDateTime lastFailureTime;
    private LocalDateTime lastSuccessTime;
    private LocalDateTime lastProbeTime;
    private int consecutive5xxCount;
    
    public enum BackendStatus {
        HEALTHY,
        UNHEALTHY,
        PROBING
    }
    
    public boolean isAvailable() {
        return status == BackendStatus.HEALTHY && weight > 0;
    }
}
