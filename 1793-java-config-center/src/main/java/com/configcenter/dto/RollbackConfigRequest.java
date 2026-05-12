package com.configcenter.dto;

import lombok.Data;

@Data
public class RollbackConfigRequest {
    
    private Long targetVersion;
    
    private String operator;
}
