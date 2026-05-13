package com.loadbalancer.dto;

import jakarta.validation.constraints.Min;
import lombok.Data;

@Data
public class HealthConfigRequest {
    @Min(value = 1, message = "Interval must be >= 1 second")
    private Integer intervalSeconds;
    
    @Min(value = 1, message = "Timeout must be >= 1 second")
    private Integer timeoutSeconds;
}
