package com.example.memorymq.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class ConsumeRequest {
    @NotBlank
    private String topic;
    @NotBlank
    private String consumerGroup;
    private Integer maxMessages;
    private Integer timeoutSeconds;
}
