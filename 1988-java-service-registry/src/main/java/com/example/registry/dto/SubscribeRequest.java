package com.example.registry.dto;

import lombok.Data;

import javax.validation.constraints.NotBlank;

@Data
public class SubscribeRequest {
    
    @NotBlank
    private String serviceName;
    
    @NotBlank
    private String callbackUrl;
}
