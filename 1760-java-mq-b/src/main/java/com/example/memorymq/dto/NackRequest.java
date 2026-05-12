package com.example.memorymq.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

import java.util.List;

@Data
public class NackRequest {
    @NotBlank
    private String topic;
    @NotBlank
    private String consumerGroup;
    private List<String> messageIds;
    private Boolean requeue;
}
