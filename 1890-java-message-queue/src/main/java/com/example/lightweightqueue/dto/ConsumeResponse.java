package com.example.lightweightqueue.dto;

import lombok.Data;
import lombok.Builder;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ConsumeResponse {
    private boolean success;
    private String messageId;
    private String body;
    private String format;
    private String error;
}
