package com.health.model;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Positive;
import lombok.Data;

import java.time.LocalDateTime;

@Data
public class ServiceRegistration {
    private String id;
    
    @NotBlank(message = "Service name is required")
    private String name;
    
    @NotNull(message = "Check type is required")
    private CheckType checkType;
    
    @NotBlank(message = "Endpoint is required")
    private String endpoint;
    
    @Positive(message = "Port must be positive")
    private Integer port;
    
    private String path;
    
    private String method = "GET";
    
    private Integer intervalSeconds;
    
    private Integer timeoutMilliseconds;
    
    private Integer unhealthyThreshold;
    
    private LocalDateTime createdAt;
    private LocalDateTime lastCheckedAt;
    private HealthStatus status = HealthStatus.UNKNOWN;
    private Integer consecutiveFailures = 0;
    private String lastMessage;
}