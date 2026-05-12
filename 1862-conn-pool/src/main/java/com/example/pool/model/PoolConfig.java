package com.example.pool.model;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Positive;
import lombok.Data;

import java.util.Map;

@Data
public class PoolConfig {
    
    @NotBlank(message = "name is required")
    private String name;
    
    @NotNull(message = "type is required")
    private DataSourceType type;
    
    @NotNull(message = "connectionParams is required")
    private Map<String, String> connectionParams;
    
    @Positive(message = "maxConnections must be positive")
    private int maxConnections = 10;
    
    @Positive(message = "minIdle must be positive")
    private int minIdle = 2;
    
    @Positive(message = "acquireTimeoutSeconds must be positive")
    private int acquireTimeoutSeconds = 30;
    
    private int idleTimeoutSeconds = 60;
}
