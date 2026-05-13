package com.poolguard.model;

import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotBlank;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class PoolConfig {
    @NotBlank(message = "Name is required")
    private String name;
    
    @Min(value = 1, message = "max_size must be at least 1")
    private int maxSize;
    
    @Min(value = 0, message = "min_idle must be at least 0")
    private int minIdle;
    
    @Min(value = 1, message = "health_check_interval must be at least 1 second")
    private int healthCheckInterval;
}
