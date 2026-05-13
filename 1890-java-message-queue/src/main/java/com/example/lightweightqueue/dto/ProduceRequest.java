package com.example.lightweightqueue.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class ProduceRequest {
    @NotBlank
    private String topic;
    
    @NotBlank
    private String body;
    
    private String format;
}
