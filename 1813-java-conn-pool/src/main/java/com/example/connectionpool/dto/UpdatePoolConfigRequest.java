package com.example.connectionpool.dto;

import lombok.Data;

@Data
public class UpdatePoolConfigRequest {
    private Integer maxConnections;
    private Long waitTimeoutMs;
    private Integer leakThresholdSeconds;
    private Integer maxQueueSize;
}
