package com.example.lightweightqueue.dto;

import lombok.Data;
import lombok.Builder;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ProduceResponse {
    private boolean success;
    private String messageId;
    private String error;
}
