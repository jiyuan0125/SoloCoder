package com.configcenter.dto;

import lombok.Data;

@Data
public class UpdateConfigRequest {
    
    private String value;
    
    private String operator;
}
