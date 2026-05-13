package com.example.lightweightqueue.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class AckRequest {
    @NotBlank
    private String topic;
    
    @NotBlank
    private String groupId;
    
    @NotBlank
    private String consumerId;
    
    @NotBlank
    private String messageId;
}
