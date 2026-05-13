package com.example.registry.dto;

import lombok.Data;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.Positive;
import java.util.Map;

@Data
public class RegisterRequest {
    
    @NotBlank
    private String serviceName;
    
    @NotBlank
    private String ip;
    
    @Positive
    private int port;
    
    private Map<String, String> metadata;
}
