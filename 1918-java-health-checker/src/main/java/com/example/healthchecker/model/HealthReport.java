package com.example.healthchecker.model;

import lombok.Data;
import java.time.LocalDateTime;

@Data
public class HealthReport {
    private String serviceName;
    private HealthStatus status;
    private LocalDateTime reportedAt = LocalDateTime.now();
}
