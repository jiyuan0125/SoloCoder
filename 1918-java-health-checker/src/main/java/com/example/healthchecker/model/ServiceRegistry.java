package com.example.healthchecker.model;

import lombok.Data;
import java.time.LocalDateTime;
import java.util.HashSet;
import java.util.Set;

@Data
public class ServiceRegistry {
    private String name;
    private HealthStatus status = HealthStatus.HEALTHY;
    private boolean inMaintenance = false;
    private Set<String> dependencies = new HashSet<>();
    private LocalDateTime lastStatusChangeTime = LocalDateTime.now();
}
