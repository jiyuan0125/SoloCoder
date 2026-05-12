package com.example.pool.model;

import lombok.Data;
import lombok.EqualsAndHashCode;

import java.util.Map;

@Data
@EqualsAndHashCode(callSuper = true)
public class PoolDetail extends PoolInfo {
    
    private Map<String, String> connectionParams;
    private int acquireTimeoutSeconds;
    private int idleTimeoutSeconds;
}
