package com.example.memorymq.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class PublishRequest {
    @NotBlank
    private String topic;
    private Object payload;
}
