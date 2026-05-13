package com.example.healthchecker.model;

import lombok.Data;
import java.time.LocalDateTime;

@Data
public class HealthEvent {
    private String serviceName;
    private HealthStatus previousStatus;
    private HealthStatus newStatus;
    private String reason;
    private LocalDateTime timestamp = LocalDateTime.now();
}
