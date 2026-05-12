package com.health.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CheckHistory {
    private String id;
    private String serviceId;
    private String serviceName;
    private LocalDateTime checkedAt;
    private HealthStatus status;
    private Integer responseTimeMs;
    private String message;
}