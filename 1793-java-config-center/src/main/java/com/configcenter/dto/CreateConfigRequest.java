package com.configcenter.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class CreateConfigRequest {
    
    @NotBlank(message = "namespace cannot be blank")
    private String namespace;
    
    @NotBlank(message = "group cannot be blank")
    private String group;
    
    @NotBlank(message = "key cannot be blank")
    private String key;
    
    private String value;
    
    private String operator;
}
