package com.example.lightweightqueue.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class ResendDeadLetterRequest {
    @NotBlank
    private String topic;
    
    private String messageId;
}
